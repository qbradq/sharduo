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
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// TargetCallback is the function signature for target callbacks
type TargetCallback func(*clientpacket.TargetResponse)

// ErrWrongPacket is the error logged when the client sends an unexpected
// packet during character login.
var ErrWrongPacket = errors.New("wrong packet")

// NetState manages the network state of a single connection.
type NetState struct {
	conn               *net.TCPConn
	sendQueue          chan serverpacket.Packet
	m                  game.Mobile
	account            *game.Account
	observedContainers map[uo.Serial]game.Container
	targetCallback     TargetCallback
	targetDeadline     uo.Time
	updateGroup        int
	deadline           uo.Time
	disconnectLock     sync.Once
}

// NewNetState constructs a new NetState object.
func NewNetState(conn *net.TCPConn) *NetState {
	return &NetState{
		conn:               conn,
		sendQueue:          make(chan serverpacket.Packet, 1024*16),
		observedContainers: make(map[uo.Serial]game.Container),
		updateGroup:        world.Random().Random(0, int(uo.DurationSecond)-1),
		deadline:           world.Time() + uo.DurationMinute*5,
	}
}

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
	if n.conn != nil {
		select {
		case n.sendQueue <- sp:
			return true
		default:
			return false
		}
	}
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
	return true
}

// Disconnect disconnects the NetState.
func (n *NetState) Disconnect() {
	n.disconnectLock.Do(func() {
		if n == nil {
			return
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

// Service is the goroutine that services the netstate.
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
	gslp, ok := cp.(*clientpacket.GameServerLogin)
	if !ok {
		log.Printf("error: %s", err.Error())
		return
	}
	account := world.AuthenticateAccount(gslp.Username, game.HashPassword(gslp.Password))
	if account == nil {
		log.Println("error: failed to create new account, reason unknown")
		return
	}
	n.account = account

	// Character list
	n.Send(&serverpacket.CharacterList{
		Names: []string{
			account.Username(), "", "", "", "", "",
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
		log.Printf("error: %s", err.Error())
		return
	}

	world.SendRequest(&CharacterLoginRequest{
		BaseWorldRequest: BaseWorldRequest{
			NetState: n,
		},
	})

	// Start the read loop
	n.readLoop(r)
}

// SendService is the goroutine that services the send queue.
func (n *NetState) SendService() {
	w := serverpacket.NewCompressedWriter()
	pw := bufio.NewWriterSize(n.conn, 128*1024)
	for {
		select {
		case p := <-n.sendQueue:
			if p == nil {
				return
			}
			if err := w.Write(p, pw); err != nil {
				log.Printf("error: %s", err.Error())
				return
			}
		default:
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
		// 5 minute timeout, should never be hit due to client ping packets
		n.conn.SetDeadline(time.Now().Add(time.Minute * 5))
		// 5 minute deadline to allow for crazy stuff like connection
		// interruptions, very long save times, etc.
		n.deadline = world.Time() + uo.DurationMinute*5

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
			log.Printf("unsupported %s packet 0x%02X:\n%s", p.PType, cp.ID(), hex.Dump(data))
		case *clientpacket.IgnoredPacket:
			// Do nothing
		default:
			// Let the world goroutine handle the packet
			world.SendRequest(&ClientPacketRequest{
				BaseWorldRequest: BaseWorldRequest{
					NetState: n,
				},
				Packet: cp,
			})
		}
	}
}

// Speech sends a speech packet to the attached client.
func (n *NetState) Speech(speaker game.Object, fmtstr string, args ...interface{}) {
	sid := uo.SerialSystem
	body := uo.BodySystem
	font := uo.FontNormal
	hue := uo.Hue(1153)
	name := ""
	text := fmt.Sprintf(fmtstr, args...)
	stype := uo.SpeechTypeSystem
	if speaker != nil {
		sid = speaker.Serial()
		stype = uo.SpeechTypeNormal
		name = speaker.DisplayName()
		if item, ok := speaker.(game.Item); ok {
			body = uo.Body(item.BaseGraphic())
		} else if mob, ok := speaker.(game.Mobile); ok {
			body = mob.Body()
		}
	}
	n.Send(&serverpacket.Speech{
		Speaker: sid,
		Body:    body,
		Font:    font,
		Hue:     hue,
		Name:    name,
		Text:    text,
		Type:    stype,
	})
}

// Cliloc sends a localized client message packet to the attached client.
func (n *NetState) Cliloc(speaker game.Object, cliloc uo.Cliloc, args ...string) {
	sid := uo.SerialSystem
	body := uo.BodySystem
	font := uo.FontNormal
	hue := uo.Hue(1153)
	name := ""
	if speaker != nil {
		sid = speaker.Serial()
		name = speaker.DisplayName()
		if item, ok := speaker.(game.Item); ok {
			body = uo.Body(item.BaseGraphic())
		} else if mob, ok := speaker.(game.Mobile); ok {
			body = mob.Body()
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
func (n *NetState) itemInfo(item game.Item) {
	var layer uo.Layer
	if layerer, ok := item.(game.Layerer); ok {
		layer = layerer.Layer()
	}
	if container, ok := item.Parent().(game.Container); ok {
		// Item in container
		n.Send(&serverpacket.AddItemToContainer{
			Item:          item.Serial(),
			Graphic:       item.BaseGraphic(),
			GraphicOffset: item.GraphicOffset(),
			Amount:        item.Amount(),
			Location:      item.Location(),
			Container:     container.Serial(),
			Hue:           item.Hue(),
		})
	} else {
		// Item on ground
		n.Send(&serverpacket.ObjectInfo{
			Serial:           item.Serial(),
			Graphic:          item.BaseGraphic(),
			GraphicIncrement: item.GraphicOffset(),
			Amount:           item.Amount(),
			Location:         item.Location(),
			Layer:            layer,
			Hue:              item.Hue(),
		})
	}
}

// sendMobile sends packets to send a mobile to the client.
func (n *NetState) sendMobile(mobile game.Mobile) {
	notoriety := uo.NotorietyEnemy
	if n.m != nil {
		notoriety = n.m.GetNotorietyFor(mobile)
	}
	p := &serverpacket.EquippedMobile{
		ID:        mobile.Serial(),
		Body:      mobile.Body(),
		Location:  mobile.Location(),
		Facing:    mobile.Facing(),
		IsRunning: mobile.IsRunning(),
		Hue:       mobile.Hue(),
		Flags:     mobile.MobileFlags(),
		Notoriety: notoriety,
	}
	mobile.MapEquipment(func(w game.Wearable) error {
		p.Equipment = append(p.Equipment, &serverpacket.EquippedMobileItem{
			ID:      w.Serial(),
			Graphic: w.BaseGraphic(),
			Layer:   w.Layer(),
			Hue:     w.Hue(),
		})
		return nil
	})
	n.Send(p)
}

// updateMobile sends a StatusBarInfo packet for the mobile.
func (n *NetState) updateMobile(mobile game.Mobile) {
	if n.m == nil {
		// Skip disconnected net states
		return
	}
	if n.m.Serial() == mobile.Serial() {
		// TODO try to detect when just sending the hp/mp/sp deltas would be sufficient
		// Full status update for the player
		n.Send(&serverpacket.StatusBarInfo{
			Mobile:         mobile.Serial(),
			Name:           mobile.DisplayName(),
			Female:         mobile.IsFemale(),
			HP:             mobile.HitPoints(),
			MaxHP:          mobile.MaxHitPoints(),
			NameChangeFlag: false,
			Strength:       mobile.Strength(),
			Dexterity:      mobile.Dexterity(),
			Intelligence:   mobile.Intelligence(),
			Stamina:        mobile.Stamina(),
			MaxStamina:     mobile.MaxStamina(),
			Mana:           mobile.Mana(),
			MaxMana:        mobile.MaxMana(),
			Gold:           mobile.Gold(),
			ArmorRating:    0,
			Weight:         int(mobile.Weight()),
			StatsCap:       uo.StatsCapDefault,
			Followers:      0,
			MaxFollowers:   uo.MaxFollowers,
		})
		return
	}
	// TODO send hp delta for other mobiles
}

// UpdateObject implements the game.NetState interface.
func (n *NetState) UpdateObject(o game.Object) {
	if item, ok := o.(game.Item); ok {
		n.itemInfo(item)
	} else if mobile, ok := o.(game.Mobile); ok {
		n.updateMobile(mobile)
	} else {
		log.Printf("error: NetState.SendObject(%s) unknown object interface", o.Serial())
	}
}

// SendObject implements the game.NetState interface.
func (n *NetState) SendObject(o game.Object) {
	if item, ok := o.(game.Item); ok {
		n.itemInfo(item)
	} else if mobile, ok := o.(game.Mobile); ok {
		n.sendMobile(mobile)
	} else {
		log.Printf("error: NetState.SendObject(%s) unknown object interface", o.Serial())
	}
}

// MoveMobile implements the game.NetState interface.
func (n *NetState) MoveMobile(mob game.Mobile) {
	noto := uo.NotorietyAttackable
	if n.m != nil {
		noto = mob.GetNotorietyFor(n.m)
	}
	n.Send(&serverpacket.MoveMobile{
		ID:        mob.Serial(),
		Body:      mob.Body(),
		Location:  mob.Location(),
		Facing:    mob.Facing(),
		Running:   mob.IsRunning(),
		Hue:       mob.Hue(),
		Flags:     mob.MobileFlags(),
		Notoriety: noto,
	})
}

// RemoveObject implements the game.NetState interface.
func (n *NetState) RemoveObject(o game.Object) {
	n.Send(&serverpacket.DeleteObject{
		Serial: o.Serial(),
	})
}

// DrawPlayer implements the game.NetState interface.
func (n *NetState) DrawPlayer() {
	if n.m == nil {
		return
	}
	n.Send(&serverpacket.DrawPlayer{
		ID:       n.m.Serial(),
		Body:     n.m.Body(),
		Hue:      n.m.Hue(),
		Flags:    n.m.MobileFlags(),
		Location: n.m.Location(),
		Facing:   n.m.Facing(),
	})
}

// WornItem sends the WornItem packet to the given mobile
func (n *NetState) WornItem(wearable game.Wearable, wearer game.Mobile) {
	n.Send(&serverpacket.WornItem{
		Item:    wearable.Serial(),
		Graphic: wearable.BaseGraphic(),
		Layer:   wearable.Layer(),
		Wearer:  wearer.Serial(),
		Hue:     wearable.Hue(),
	})
}

// DropReject sends an item move reject packet
func (n *NetState) DropReject(reason uo.MoveItemRejectReason) {
	n.Send(&serverpacket.MoveItemReject{
		Reason: reason,
	})
}

// DragItem sends the DragItem packet to the given mobile
func (n *NetState) DragItem(item game.Item, srcMob game.Mobile,
	srcLoc uo.Location, destMob game.Mobile, destLoc uo.Location) {
	if item == nil {
		return
	}
	if srcLoc.X == destLoc.X && srcLoc.Y == destLoc.Y && srcLoc.Z == destLoc.Z {
		return
	}
	srcSerial := uo.SerialSystem
	destSerial := uo.SerialSystem
	if srcMob != nil {
		srcSerial = srcMob.Serial()
	}
	if destMob != nil {
		destSerial = destMob.Serial()
	}
	if srcSerial != uo.SerialSystem && srcSerial == destSerial {
		return
	}
	n.Send(&serverpacket.DragItem{
		Graphic:             item.BaseGraphic(),
		GraphicOffset:       item.GraphicOffset(),
		Amount:              item.Amount(),
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
func (n *NetState) ContainerOpen(c game.Container) {
	if c == nil {
		return
	}
	n.observedContainers[c.Serial()] = c
	n.Send(&serverpacket.OpenContainerGump{
		GumpSerial: c.Serial(),
		Gump:       uo.GUMP(c.GumpGraphic()),
	})
	if c.ItemCount() > 0 {
		p := &serverpacket.Contents{}
		c.MapContents(func(item game.Item) error {
			p.Items = append(p.Items, &serverpacket.ContentsItem{
				Serial:        item.Serial(),
				Graphic:       item.BaseGraphic(),
				GraphicOffset: item.GraphicOffset(),
				Amount:        item.Amount(),
				Location:      item.Location(),
				Container:     c.Serial(),
				Hue:           item.Hue(),
			})
			return nil
		})
		n.Send(p)
	}
}

// ContainerClose implements the game.ContainerObserver interface
func (n *NetState) ContainerClose(c game.Container) {
	// Ignore containers that aren't being observed
	if !n.ContainerIsObserving(c) {
		return
	}
	// Close this container
	delete(n.observedContainers, c.Serial())
	n.CloseGump(c.Serial())
	c.RemoveObserver(n)
	// Close all child containers
	c.MapContents(func(item game.Item) error {
		if c, ok := item.(game.Container); ok {
			n.ContainerClose(c)
		}
		return nil
	})
}

// ContainerItemAdded implements the game.ContainerObserver interface
func (n *NetState) ContainerItemAdded(c game.Container, item game.Item) {
	n.Send(&serverpacket.AddItemToContainer{
		Item:          item.Serial(),
		Graphic:       item.BaseGraphic(),
		GraphicOffset: item.GraphicOffset(),
		Amount:        item.Amount(),
		Location:      item.Location(),
		Container:     c.Serial(),
		Hue:           item.Hue(),
	})
}

// ContainerItemRemoved implements the game.ContainerObserver interface
func (n *NetState) ContainerItemRemoved(c game.Container, item game.Item) {
	n.Send(&serverpacket.DeleteObject{
		Serial: item.Serial(),
	})
}

// ContainerRangeCheck implements the game.ContainerObserver interface
func (n *NetState) ContainerRangeCheck() {
	if len(n.observedContainers) == 0 || n.m == nil {
		return
	}
	// Make a copy of the map contents so NetState.ContainerClose can modify
	// NetState.observedContainers
	var toObserve = make([]game.Container, len(n.observedContainers))
	idx := 0
	for _, c := range n.observedContainers {
		toObserve[idx] = c
		idx++
	}
	// Observe all containers
	for _, c := range toObserve {
		root := game.RootParent(c)
		if _, ok := root.(game.Container); ok {
			// Container is somewhere on the map
			// TODO Line of sight check, this one might be costly and unnecessary
			// Range check
			if n.m.Location().XYDistance(root.Location()) > uo.MaxContainerViewRange {
				n.ContainerClose(c)
			}
		} else if m, ok := root.(game.Mobile); ok {
			// This is part of the mobile's inventory, so either inside the bank
			// box or the backpack. We always close the bank box and it's every time we
			// move and we never close the backpack.
			bbobj := m.EquipmentInSlot(uo.LayerBankBox)
			if bbobj == nil {
				continue
			}
			thisc := c
			thisp := c.Parent()
			for {
				if _, ok := thisp.(game.Mobile); ok {
					// This is the top-level container
					if thisc.Serial() == bbobj.Serial() {
						// The bank box or a child of it, close instantly
						n.ContainerClose(c)
					}
					// Otherwise this is the backpack or a child of it, leave
					// open.
					break
				} else if container, ok := thisp.(game.Container); ok {
					// This is a sub-container, inspect the parent.
					thisc = container
					thisp = thisc.Parent()
				} else {
					// Something is very wrong.
					log.Printf("error: object %s has a non-container in it's parent chain", c.Serial().String())
					break
				}
			}
		}
	}
}

// ContainerIsObserving implements the game.ContainerObserver interface
func (n *NetState) ContainerIsObserving(o game.Object) bool {
	_, found := n.observedContainers[o.Serial()]
	return found
}

// OpenPaperDoll implements the game.NetState interface
func (n *NetState) OpenPaperDoll(m game.Mobile) {
	if m == nil {
		return
	}
	if n.m != nil && n.m.Serial() == m.Serial() {
		// Player is opening thier own paper doll
		n.Send(&serverpacket.OpenPaperDoll{
			Serial:    m.Serial(),
			Text:      m.DisplayName(),
			WarMode:   false,
			Alterable: true,
		})
	} else {
		// Player is opening someone else's paper doll
		n.Send(&serverpacket.OpenPaperDoll{
			Serial:    m.Serial(),
			Text:      m.DisplayName(),
			WarMode:   false,
			Alterable: false,
		})
	}
}

// TargetSendCursor implements the game.NetState interface
func (n *NetState) TargetSendCursor(ttype uo.TargetType, fn TargetCallback) {
	if n.m == nil {
		return
	}
	n.targetCallback = fn
	n.targetDeadline = world.Time() + uo.DurationSecond*30
	n.Send(&serverpacket.Target{
		Serial:     n.m.Serial(),
		TargetType: ttype,
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
		SkillValues: n.m.Skills(),
	})
}

// Sound makes the client play a sound.
func (n *NetState) Sound(which uo.Sound, from uo.Location) {
	n.Send(&serverpacket.Sound{
		Sound:    which,
		Location: from,
	})
}

// Music makes the client play a song.
func (n *NetState) Music(song uo.Song) {
	n.Send(&serverpacket.Music{
		Song: song,
	})
}

// Animate animates a mobile on the client side
func (n *NetState) Animate(mob game.Mobile, at uo.AnimationType, aa uo.AnimationAction) {
	if mob == nil {
		return
	}
	n.Send(&serverpacket.Animation{
		Serial:          mob.Serial(),
		AnimationType:   at,
		AnimationAction: aa,
	})
}
