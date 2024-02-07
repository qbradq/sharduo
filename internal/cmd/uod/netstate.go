package uod

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// ErrWrongPacket is the error logged when the client sends an unexpected
// packet during character login.
var ErrWrongPacket = errors.New("wrong packet")

// gumpDescription describes a GUMP instance.
type gumpDescription struct {
	g gumps.GUMP // The GUMP being managed
	t uo.Serial  // Serial of the target object
	p uo.Serial  // Serial of the parameter object
}

// NetState manages the network state of a single connection.
type NetState struct {
	conn               *net.TCPConn                       // TCP connection we are connected through
	sendQueue          chan serverpacket.Packet           // Queue of packets to send on conn
	m                  *game.Mobile                       // Mobile being controlled by this client, if any
	account            *game.Account                      // Account of the player or a mock account, never nil
	observedContainers map[uo.Serial]*game.Item           // All containers being observed by the player
	targetCallback     func(*clientpacket.TargetResponse) // Function to execute when the next targeting request comes in
	targetDeadline     uo.Time                            // When the outstanding targeting request will expire
	updateGroup        int                                // Used in load balancing
	deadline           uo.Time                            // When this connection should be closed due to inactivity
	disconnectLock     sync.Once                          // Make sure we don't try to close conn more than once
	gumps              map[uo.Serial]*gumpDescription     // All open GUMPs on the client side
	nextActionTime     uo.Time                            // When the next action can be taken
	textReplyFn        func(string)                       // Function to trigger in response to a text GUMP reply (packet 0xAC)
}

// NewNetState constructs a new NetState object.
func NewNetState(conn *net.TCPConn) *NetState {
	return &NetState{
		conn:               conn,
		sendQueue:          make(chan serverpacket.Packet, 1024*16),
		observedContainers: make(map[uo.Serial]*game.Item),
		updateGroup:        util.Random(0, int(uo.DurationSecond)-1),
		deadline:           world.Time() + uo.DurationMinute*5,
		gumps:              make(map[uo.Serial]*gumpDescription),
	}
}

// Mobile returns the mobile associated with the state if any.
func (n *NetState) Mobile() *game.Mobile { return n.m }

// Update should be called once per real-world second to search for stale net
// states, expired targeting cursors, etc.
func (n *NetState) Update() {
	if world.Time() > n.deadline {
		n.Disconnect()
		return
	}
	if n.targetCallback != nil && world.Time() > n.targetDeadline {
		n.targetCallback = nil
		n.targetDeadline = uo.TimeNever
		n.Send(&serverpacket.Target{
			Serial:     uo.SerialZero,
			TargetType: uo.TargetTypeObject,
			CursorType: uo.CursorTypeCancel,
		})
	}
}

// Send attempts to add a packet to the client's send queue and returns false if
// the queue is full.
func (n *NetState) Send(sp serverpacket.Packet) bool {
	if sp == nil {
		return true
	}
	if n.conn != nil {
		select {
		case n.sendQueue <- sp:
			return true
		default:
			return false
		}
	} else {
		// Packet filtering for internal net states
		switch p := sp.(type) {
		case *serverpacket.Speech:
			// Log all messages
			if p.Name == "" {
				log.Printf("info: %s", p.Text)
			} else {
				log.Printf("info: %s: %s", p.Name, p.Text)
			}
		}
	}
	return true
}

// Disconnect disconnects the NetState.
func (n *NetState) Disconnect() {
	if n == nil {
		return
	}
	n.disconnectLock.Do(func() {
		if n == nil {
			return
		}
		for _, c := range n.observedContainers {
			c.RemoveObserver(n)
		}
		if n.conn != nil {
			n.conn.Close()
		}
		if n.sendQueue != nil {
			close(n.sendQueue)
			n.sendQueue = nil
		}
		gameNetStates.Delete(n)
	})
}

// TakeAction returns true if an action is allowed at this time. Examples of
// actions are double-clicking basically anything or lifting an item. This
// method assumes that the action will be taken after this call and sets
// internal states to limit action speed.
func (n *NetState) TakeAction() bool {
	now := world.Time()
	if now < n.nextActionTime {
		return false
	}
	n.nextActionTime = now + uo.DurationSecond/2
	return true
}

// Service is the goroutine that services the net state.
func (n *NetState) Service() {
	// When this goroutine ends so will the TCP connection.
	defer n.Disconnect()
	// Give the player 15 minutes at the login / character create screen
	n.deadline = world.Time() + uo.DurationMinute*15
	// Start SendService
	go n.SendService()
	// Configure TCP QoS
	n.conn.SetKeepAlive(false)
	n.conn.SetLinger(0)
	n.conn.SetNoDelay(true)
	n.conn.SetReadBuffer(64 * 1024)
	n.conn.SetWriteBuffer(128 * 1024)
	n.conn.SetDeadline(time.Now().Add(time.Minute * 15))
	r := clientpacket.NewReader(n.conn)
	log.Printf("info: connection from %s", n.conn.RemoteAddr().String())
	// Connection header
	if err := r.ReadConnectionHeader(); err != nil {
		log.Printf("error: %s", err.Error())
		return
	}
	// Game server login packet
	cp, err := r.ReadPacket()
	if err != nil {
		log.Printf("error: %s", err.Error())
		return
	}
	gslP, ok := cp.(*clientpacket.GameServerLogin)
	if !ok {
		log.Printf("error: expected GameServerLogin packet")
		return
	}
	account := world.AuthenticateAccount(gslP.Username, util.HashPassword(gslP.Password))
	if account == nil {
		log.Println("error: failed to create new account, reason unknown")
		return
	}
	n.account = account
	// Character list
	n.Send(&serverpacket.CharacterList{
		Names: []string{
			account.Username, "", "", "", "", "",
		},
	})
	// Character login
	cp, err = r.ReadPacket()
	if err != nil {
		log.Printf("error: %s", err.Error())
		return
	}
	_, ok = cp.(*clientpacket.CharacterLogin)
	if !ok {
		log.Println("error: unexpected packet during character login")
		return
	}
	world.SendPacket(&CharacterLogin{}, n)
	// Start the read loop
	n.readLoop(r)
}

// SendService is the goroutine that services the send queue.
func (n *NetState) SendService() {
	w := serverpacket.NewCompressedWriter()
	pw := bufio.NewWriterSize(n.conn, 128*1024)
	ticker := time.NewTicker(time.Millisecond * 50)
	for {
		select {
		case p := <-n.sendQueue:
			// Prioritize writing outbound packets
			if p == nil {
				return
			}
			if err := w.Write(p, pw); err != nil {
				log.Printf("error: %s", err.Error())
				return
			}
		case <-ticker.C:
			// Flush the buffer every 50ms
			if pw.Buffered() > 0 {
				if err := pw.Flush(); err != nil {
					log.Printf("error: %s", err.Error())
					return
				}
			}
		}
	}
}

func (n *NetState) readLoop(r *clientpacket.Reader) {
	for {
		data, err := r.Read()
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "closed network connection") {
				log.Printf("error: %s, disconnecting client", err.Error())
			}
			n.Disconnect()
			return
		}
		// 1 minute timeout, should never be hit due to client ping packets
		n.conn.SetDeadline(time.Now().Add(time.Minute))
		// 1 minute deadline to allow for crazy stuff like connection
		// interruptions, very long save times, etc.
		n.deadline = world.Time() + uo.DurationMinute

		cp := clientpacket.New(data)
		switch p := cp.(type) {
		case nil:
			log.Printf("error: unknown packet 0x%02X", data[0])
		case *clientpacket.MalformedPacket:
			log.Printf("error: malformed packet 0x%02X", p.ID())
		case *clientpacket.UnknownPacket:
			log.Printf("error: unknown %s packet 0x%02X", p.PType, cp.ID())
			return
		case *clientpacket.UnsupportedPacket:
			log.Printf("trace: unsupported %s packet 0x%02X:\n%s", p.PType, cp.ID(), hex.Dump(data))
		case *clientpacket.IgnoredPacket:
			// Do nothing
		default:
			// Let the world goroutine handle the packet
			world.SendPacket(cp, n)
		}
	}
}

// Speech sends a speech packet to the attached client.
func (n *NetState) Speech(speaker any, fs string, args ...interface{}) {
	sid := uo.SerialSystem
	body := uo.BodySystem
	font := uo.FontNormal
	hue := uo.Hue(1153)
	name := ""
	text := fmt.Sprintf(fs, args...)
	sType := uo.SpeechTypeSystem
	if speaker != nil {
		switch s := speaker.(type) {
		case *game.Item:
			sid = s.Serial
			sType = uo.SpeechTypeNormal
			name = s.DisplayName()
			body = uo.Body(s.CurrentGraphic())
		case *game.Mobile:
			sid = s.Serial
			sType = uo.SpeechTypeNormal
			name = s.DisplayName()
			body = s.Body
		default:
			panic("unknown object type")
		}
	}
	n.Send(&serverpacket.Speech{
		Speaker: sid,
		Body:    body,
		Font:    font,
		Hue:     hue,
		Name:    name,
		Text:    text,
		Type:    sType,
	})
}

// Cliloc sends a localized client message packet to the attached client.
func (n *NetState) Cliloc(speaker any, cliloc uo.Cliloc, args ...string) {
	sid := uo.SerialSystem
	body := uo.BodySystem
	font := uo.FontNormal
	hue := uo.Hue(1153)
	name := ""
	if speaker != nil {
		switch s := speaker.(type) {
		case *game.Item:
			sid = s.Serial
			name = s.DisplayName()
			body = uo.Body(s.CurrentGraphic())
		case *game.Mobile:
			sid = s.Serial
			name = s.DisplayName()
			body = s.Body
		default:
			panic("unknown object type")
		}
	}
	n.Send(&serverpacket.ClilocMessage{
		Speaker:   sid,
		Body:      body,
		Font:      font,
		Hue:       hue,
		Name:      name,
		Cliloc:    cliloc,
		Arguments: []string(args),
	})
}

// itemInfo sends ObjectInfo or AddItemToContainer packets for the item
func (n *NetState) itemInfo(item *game.Item) {
	if item == nil || item.Removed {
		return
	}
	var layer uo.Layer
	if item.Layer == uo.LayerInvalid || item.Layer > uo.LayerLastVisible {
		// Dirty hack to prevent things like mount items from being sent
		// like a normal item.
		return
	}
	if item.Container != nil {
		// Item in container
		n.Send(&serverpacket.AddItemToContainer{
			Item:      item.Serial,
			Graphic:   item.CurrentGraphic(),
			Amount:    item.Amount,
			Location:  item.Location,
			Container: item.Container.Serial,
			Hue:       item.Hue,
		})
	} else {
		// Item on ground
		n.Send(&serverpacket.ObjectInfo{
			Serial:   item.Serial,
			Graphic:  item.CurrentGraphic(),
			Amount:   item.Amount,
			Location: item.Location,
			Layer:    layer,
			Hue:      item.Hue,
			Movable:  !item.HasFlags(game.ItemFlagsFixed),
		})
	}
	// OPL support
	_, oi := item.OPLPackets()
	if oi != nil {
		n.Send(oi)
	}
}

// sendMobile sends packets to send a mobile to the client.
func (n *NetState) sendMobile(mobile *game.Mobile) {
	// Skip disconnected net states, mobiles that have been removed, and other
	// non-removed mobiles that are no longer on the map, such as mounts within
	// mount items.
	if n.m == nil || mobile == nil || mobile.Removed {
		return
	}
	p := &serverpacket.EquippedMobile{
		ID:        mobile.Serial,
		Body:      mobile.Body,
		Location:  mobile.Location,
		Facing:    mobile.Facing,
		IsRunning: mobile.Running,
		Hue:       mobile.Hue,
		Flags:     mobile.Flags(),
		Notoriety: n.m.NotorietyFor(mobile),
	}
	for i, e := range mobile.Equipment {
		if i < int(uo.LayerFirstValid) {
			continue
		}
		if i > int(uo.LayerLastValid) {
			break
		}
		if e == nil {
			continue
		}
		p.Equipment = append(p.Equipment, &serverpacket.EquippedMobileItem{
			ID:      e.Serial,
			Graphic: e.CurrentGraphic(),
			Layer:   e.Layer,
			Hue:     e.Hue,
		})
	}
	n.Send(p)
	// OPL support
	_, oi := mobile.OPLPackets()
	if oi != nil {
		n.Send(oi)
	}
}

// updateMobile sends a StatusBarInfo packet for the mobile.
func (n *NetState) updateMobile(mobile *game.Mobile) {
	// Skip disconnected net states, mobiles that have been removed, and other
	// non-removed mobiles that are no longer on the map, such as mounts within
	// mount items.
	if n.m == nil || mobile == nil || mobile.Removed {
		return
	}
	if n.m == mobile {
		// Full status update for the player
		n.Send(&serverpacket.StatusBarInfo{
			Mobile:         mobile.Serial,
			Name:           mobile.DisplayName(),
			Female:         mobile.Female,
			HP:             mobile.Hits,
			MaxHP:          mobile.MaxHits,
			NameChangeFlag: false,
			Strength:       mobile.Strength,
			Dexterity:      mobile.Dexterity,
			Intelligence:   mobile.Intelligence,
			Stamina:        mobile.Stamina,
			MaxStamina:     mobile.MaxStamina,
			Mana:           mobile.Mana,
			MaxMana:        mobile.MaxMana,
			Gold:           mobile.Equipment[uo.LayerBackpack].Gold,
			ArmorRating:    0,
			Weight:         int(mobile.Weight),
			StatsCap:       uo.StatsCapDefault,
			Followers:      0,
			MaxFollowers:   uo.MaxFollowers,
		})
		return
	} else if mobile.ControlMaster == n.m {
		// Send rename-able status bar
		n.Send(&serverpacket.StatusBarInfo{
			Mobile:         mobile.Serial,
			Name:           mobile.DisplayName(),
			Female:         mobile.Female,
			HP:             mobile.Hits,
			MaxHP:          mobile.MaxHits,
			NameChangeFlag: true,
		})
	} else {
		// Send hp delta for other mobiles
		n.Send(&serverpacket.UpdateHealth{
			Serial:  mobile.Serial,
			Hits:    mobile.Hits,
			MaxHits: mobile.MaxHits,
		})
	}
}

// UpdateMobile sends an update packet for the mobile.
func (n *NetState) UpdateMobile(m *game.Mobile) {
	if n.m == nil || m == nil {
		return
	}
	if m.Removed || !n.m.CanSee(&m.Object) {
		return
	}
	n.updateMobile(m)
}

// UpdateItem sends an update packet for the item.
func (n *NetState) UpdateItem(i *game.Item) {
	if n.m == nil || i == nil {
		return
	}
	if i.Removed || !n.m.CanSee(&i.Object) {
		return
	}
	n.itemInfo(i)
}

// SendMobile sends an initial information packet for the mobile.
func (n *NetState) SendMobile(m *game.Mobile) {
	if n.m == nil || m == nil {
		return
	}
	if m.Removed || !n.m.CanSee(&m.Object) {
		return
	}
	n.sendMobile(m)
}

// SendItem sends an initial information packet for the item.
func (n *NetState) SendItem(i *game.Item) {
	if n.m == nil || i == nil {
		return
	}
	if i.Removed || !n.m.CanSee(&i.Object) {
		return
	}
	n.itemInfo(i)
}

// MoveMobile sends a packet to inform the client that the mobile moved.
func (n *NetState) MoveMobile(mob *game.Mobile) {
	nt := uo.NotorietyAttackable
	if n.m != nil {
		nt = mob.NotorietyFor(n.m)
		if !mob.CanSee(&n.m.Object) {
			return
		}
	}
	n.Send(&serverpacket.MoveMobile{
		ID:        mob.Serial,
		Body:      mob.Body,
		Location:  mob.Location,
		Facing:    mob.Facing,
		Running:   mob.Running,
		Hue:       mob.Hue,
		Flags:     mob.Flags(),
		Notoriety: nt,
	})
}

// RemoveMobile sends a packet to the client that removes the mobile from the
// client's view of the game.
func (n *NetState) RemoveMobile(m *game.Mobile) {
	n.Send(&serverpacket.DeleteObject{
		Serial: m.Serial,
	})
}

// RemoveItem sends a packet to the client that removes the item from the
// client's view of the game.
func (n *NetState) RemoveItem(i *game.Item) {
	n.Send(&serverpacket.DeleteObject{
		Serial: i.Serial,
	})
}

// DrawPlayer sends the draw player packet to the client.
func (n *NetState) DrawPlayer() {
	if n.m == nil {
		return
	}
	n.Send(&serverpacket.DrawPlayer{
		ID:       n.m.Serial,
		Body:     n.m.Body,
		Hue:      n.m.Hue,
		Flags:    n.m.Flags(),
		Location: n.m.Location,
		Facing:   n.m.Facing,
	})
}

// WornItem sends the WornItem packet to the given mobile.
func (n *NetState) WornItem(i *game.Item, wearer *game.Mobile) {
	n.Send(&serverpacket.WornItem{
		Item:    i.Serial,
		Graphic: i.CurrentGraphic(),
		Layer:   i.Layer,
		Wearer:  wearer.Serial,
		Hue:     i.Hue,
	})
}

// DropReject sends an item move reject packet.
func (n *NetState) DropReject(reason uo.MoveItemRejectReason) {
	n.Send(&serverpacket.MoveItemReject{
		Reason: reason,
	})
}

// DragItem sends the DragItem packet to the given mobile.
func (n *NetState) DragItem(item *game.Item, srcMob *game.Mobile, srcLoc uo.Point, destMob *game.Mobile, destLoc uo.Point) {
	if item == nil {
		return
	}
	if srcLoc.X == destLoc.X && srcLoc.Y == destLoc.Y && srcLoc.Z == destLoc.Z {
		return
	}
	srcSerial := uo.SerialSystem
	destSerial := uo.SerialSystem
	if srcMob != nil {
		srcSerial = srcMob.Serial
	}
	if destMob != nil {
		destSerial = destMob.Serial
	}
	if srcSerial != uo.SerialSystem && srcSerial == destSerial {
		return
	}
	n.Send(&serverpacket.DragItem{
		Graphic:             item.CurrentGraphic(),
		Amount:              item.Amount,
		Source:              srcSerial,
		SourceLocation:      srcLoc,
		Destination:         destSerial,
		DestinationLocation: destLoc,
	})
}

// CloseGump closes the named gump on the client
func (n *NetState) CloseGump(gump uo.Serial) {
	n.Send(&serverpacket.CloseGump{
		Gump:   gump,
		Button: 0,
	})
}

// ContainerOpen implements the game.ContainerObserver interface
func (n *NetState) ContainerOpen(c *game.Item) {
	if c == nil || !c.HasFlags(game.ItemFlagsContainer) {
		return
	}
	n.observedContainers[c.Serial] = c
	n.Send(&serverpacket.OpenContainerGump{
		GumpSerial: c.Serial,
		Gump:       uo.GUMP(c.Gump),
	})
	if c.ItemCount > 0 {
		p := &serverpacket.Contents{}
		p.Items = make([]serverpacket.ContentsItem, 0, c.ItemCount)
		for _, item := range c.Contents {
			p.Items = append(p.Items, serverpacket.ContentsItem{
				Serial:    item.Serial,
				Graphic:   item.CurrentGraphic(),
				Amount:    item.Amount,
				Location:  item.Location,
				Container: c.Serial,
				Hue:       item.Hue,
			})
		}
		n.Send(p)
		// OPL support
		for _, item := range c.Contents {
			_, oi := item.OPLPackets()
			if oi == nil {
				continue
			}
			n.Send(oi)
		}
	}
}

// ContainerClose implements the game.ContainerObserver interface
func (n *NetState) ContainerClose(c *game.Item) {
	// Ignore containers that aren't being observed
	if !n.ContainerIsObserving(c) {
		return
	}
	// Close this container
	delete(n.observedContainers, c.Serial)
	n.CloseGump(c.Serial)
	c.RemoveObserver(n)
	// Close all child containers
	for _, item := range c.Contents {
		if item.HasFlags(game.ItemFlagsContainer) {
			n.ContainerClose(item)
		}
	}
}

// ContainerItemAdded implements the game.ContainerObserver interface
func (n *NetState) ContainerItemAdded(c *game.Item, item *game.Item) {
	n.Send(&serverpacket.AddItemToContainer{
		Item:      item.Serial,
		Graphic:   item.CurrentGraphic(),
		Amount:    item.Amount,
		Location:  item.Location,
		Container: c.Serial,
		Hue:       item.Hue,
	})
	_, oi := item.OPLPackets()
	if oi != nil {
		n.Send(oi)
	}
}

// ContainerItemRemoved implements the game.ContainerObserver interface
func (n *NetState) ContainerItemRemoved(c *game.Item, item *game.Item) {
	n.Send(&serverpacket.DeleteObject{
		Serial: item.Serial,
	})
}

// ContainerItemOPLChanged implements the game.ContainerObserver interface.
func (n *NetState) ContainerItemOPLChanged(c *game.Item, item *game.Item) {
	_, info := item.OPLPackets()
	if info != nil {
		n.Send(info)
	}
}

// ContainerRangeCheck implements the game.ContainerObserver interface
func (n *NetState) ContainerRangeCheck() {
	if len(n.observedContainers) == 0 || n.m == nil {
		return
	}
	// Make a copy of the map contents so NetState.ContainerClose can modify
	// NetState.observedContainers
	var toObserve = make([]*game.Item, len(n.observedContainers))
	idx := 0
	for _, c := range n.observedContainers {
		toObserve[idx] = c
		idx++
	}
	// Observe all containers
	for _, c := range toObserve {
		rc := c.RootContainer()
		if rc == nil {
			if c.Wearer == nil {
				// Container is somewhere on the map
				if n.m.Location.XYDistance(c.Location) > uo.MaxContainerViewRange {
					n.ContainerClose(c)
				}
			} else if rc.Wearer == n.m {
				// Container is being worn by our player, so either the bank box
				// or backpack. We always close the bank box on every step and
				// we never close our own backpack.
				if rc.Layer == uo.LayerBankBox {
					n.ContainerClose(c)
				}
			} else {
				// Container is being worn by another mobile so we are either
				// snooping or accessing a controlled mobile's pack. Enforce
				// normal view range restrictions.
				if n.m.Location.XYDistance(c.Wearer.Location) > uo.MaxContainerViewRange {
					n.ContainerClose(c)
				}
			}
		} else {
			if rc.Layer == uo.LayerBankBox {
				// The container is within someone's bank box, close with every
				// step.
				n.ContainerClose(c)
			} else if rc.Wearer != n.m {
				// Container is either being snooped or is within a controlled
				// mobile's backpack, enforce normal view range restrictions.
				if n.m.Location.XYDistance(rc.Wearer.Location) > uo.MaxContainerViewRange {
					n.ContainerClose(c)
				}
			} // Else the container is within our backpack, never close it.
		}
	}
}

// ContainerIsObserving implements the game.ContainerObserver interface
func (n *NetState) ContainerIsObserving(i *game.Item) bool {
	_, found := n.observedContainers[i.Serial]
	return found
}

// OpenPaperDoll opens a paper doll GUMP for mobile m on the client attached to
// n.
func (n *NetState) OpenPaperDoll(m *game.Mobile) {
	if m == nil {
		return
	}
	if n.m == m {
		// Player is opening their own paper doll
		n.Send(&serverpacket.OpenPaperDoll{
			Serial:    m.Serial,
			Text:      m.DisplayName(),
			WarMode:   false,
			Alterable: true,
		})
	} else {
		// Player is opening someone else's paper doll
		n.Send(&serverpacket.OpenPaperDoll{
			Serial:    m.Serial,
			Text:      m.DisplayName(),
			WarMode:   false,
			Alterable: false,
		})
	}
}

// TargetSendCursor implements the game.NetState interface
func (n *NetState) TargetSendCursor(tType uo.TargetType, fn func(*clientpacket.TargetResponse)) {
	if n.m == nil {
		return
	}
	n.targetCallback = fn
	n.targetDeadline = world.Time() + uo.DurationSecond*30
	n.Send(&serverpacket.Target{
		Serial:     n.m.Serial,
		TargetType: tType,
		CursorType: uo.CursorTypeNeutral,
	})
}

// TargetResponse handles the target response
func (n *NetState) TargetResponse(r *clientpacket.TargetResponse) {
	cb := n.targetCallback
	dl := n.targetDeadline
	n.targetCallback = nil
	n.targetDeadline = uo.TimeNever
	// Target has timed out or never existed
	if cb == nil {
		return
	}
	// Target has timed out before NetState.Update() could notice
	if world.Time() > dl {
		return
	}
	// Target cursor canceled
	if r.Location.X == uo.TargetCanceledX && r.Location.Y == uo.TargetCanceledY {
		return
	}
	// Execute callback
	cb(r)
}

// UpdateSkill implements the game.NetState interface.
func (n *NetState) UpdateSkill(which uo.Skill, lock uo.SkillLock, value int) {
	n.Send(&serverpacket.SingleSkillUpdate{
		Skill: which,
		Value: value,
		Lock:  lock,
	})
}

// SendAllSkills sends all skill values and lock states to the client.
func (n *NetState) SendAllSkills() {
	if n.m == nil {
		return
	}
	n.Send(&serverpacket.FullSkillUpdate{
		SkillValues: n.m.Skills,
	})
}

// Sound makes the client play a sound.
func (n *NetState) Sound(which uo.Sound, from uo.Point) {
	n.Send(&serverpacket.Sound{
		Sound:    which,
		Location: from,
	})
}

// Music makes the client play a song.
func (n *NetState) Music(song uo.Music) {
	n.Send(&serverpacket.Music{
		Song: uo.MusicApproach,
	})
	n.Send(&serverpacket.Music{
		Song: song,
	})
}

// Animate animates a mobile on the client side.
func (n *NetState) Animate(mob *game.Mobile, at uo.AnimationType, aa uo.AnimationAction) {
	if mob == nil {
		return
	}
	n.Send(&serverpacket.Animation{
		Serial:          mob.Serial,
		AnimationType:   at,
		AnimationAction: aa,
	})
}

// GUMP sends a generic GUMP to the client.
func (n *NetState) GUMP(gi any, target, param uo.Serial) {
	g, ok := gi.(gumps.GUMP)
	if !ok {
		return
	}
	s := g.TypeCode()
	n.gumps[s] = &gumpDescription{
		g: g,
		t: target,
		p: param,
	}
	n.RefreshGUMP(g)
}

// GUMPReply dispatches a GUMP reply
func (n *NetState) GUMPReply(s uo.Serial, p *clientpacket.GUMPReply) {
	if n.m == nil {
		return
	}
	// Handle close requests
	if p.Button == 0 {
		delete(n.gumps, s)
		return
	}
	// Resolve the GUMP on our end
	d := n.gumps[s]
	if d == nil {
		return
	}
	d.g.HandleReply(n, p)
	n.RefreshGUMP(d.g)
}

// RefreshGUMP refreshes the passed GUMP on the client side.
func (n *NetState) RefreshGUMP(gi any) {
	if n.m == nil || gi == nil {
		return
	}
	g, ok := gi.(gumps.GUMP)
	if !ok {
		return
	}
	s := g.TypeCode()
	d, found := n.gumps[s]
	if !found {
		return
	}
	// Resolve objects
	var tg any
	if d.t != uo.SerialSystem {
		tg = world.Find(d.t)
		if tg == nil {
			// Target of the GUMP has been removed, close the GUMP
			delete(n.gumps, s)
			return
		}
	}
	var pm any
	if d.p != uo.SerialSystem {
		pm = world.Find(d.p)
		if pm == nil {
			// Parameter of the GUMP has been removed, close the GUMP
			delete(n.gumps, s)
			return
		}
	}
	// Re-lay the GUMP
	g.InvalidateLayout()
	g.Layout(tg, pm)
	// Send the packet
	n.Send(g.Packet(0, 0, n.m.Serial, s))
}

// GetText sends a GUMP for text entry.
func (n *NetState) GetText(value, description string, max int, fn func(string)) {
	n.textReplyFn = fn
	n.Send(&serverpacket.TextEntryGUMP{
		Serial:      uo.SerialTextGUMP,
		Value:       value,
		Description: description,
		CanCancel:   false,
		MaxLength:   max,
	})
}

// HandleGUMPTextReply handles text GUMP reply packets.
func (n *NetState) HandleGUMPTextReply(value string) {
	if n.textReplyFn != nil {
		n.textReplyFn(value)
		n.textReplyFn = nil
	}
}

// GetGUMPByID returns a pointer to the identified GUMP or nil if the state does
// not currently have a GUMP of that type open.
func (n *NetState) GetGUMPByID(s uo.Serial) any {
	gd, found := n.gumps[s]
	if !found {
		return nil
	}
	return gd.g
}
