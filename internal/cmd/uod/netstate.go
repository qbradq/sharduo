package uod

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// ErrWrongPacket is the error logged when the client sends an unexpected
// packet during character login.
var ErrWrongPacket = errors.New("wrong packet")

// NetState manages the network state of a single connection.
type NetState struct {
	conn               *net.TCPConn
	sendQueue          chan serverpacket.Packet
	id                 string
	m                  game.Mobile
	account            *game.Account
	observedContainers map[uo.Serial]game.Container
}

// NewNetState constructs a new NetState object.
func NewNetState(conn *net.TCPConn) *NetState {
	uuid, _ := uuid.NewRandom()
	return &NetState{
		conn:               conn,
		sendQueue:          make(chan serverpacket.Packet, 1024*16),
		id:                 uuid.String(),
		observedContainers: make(map[uo.Serial]game.Container),
	}
}

// Log logs a message from this netstate as in fmt.Sprintf.
func (n *NetState) Log(fmtstr string, args ...interface{}) {
	s := fmt.Sprintf(fmtstr, args...)
	log.Printf("%s:log:%s", n.id, s)
}

// Error logs an message from this netstate and disconnect it.
func (n *NetState) Error(where string, err error) {
	if err != nil {
		log.Printf("%s:error:at %s:%s", n.id, where, err.Error())
	} else {
		log.Printf("%s:error:at %s", n.id, where)
	}
	n.Disconnect()
}

// Send attempts to add a packet to the client's send queue and returns false if
// the queue is full.
func (n *NetState) Send(p serverpacket.Packet) bool {
	select {
	case n.sendQueue <- p:
		return true
	default:
		return false
	}
}

// Disconnect disconnects the NetState.
func (n *NetState) Disconnect() {
	if n != nil {
		if n.conn != nil {
			n.conn.Close()
		}
		if n.sendQueue != nil {
			close(n.sendQueue)
			n.sendQueue = nil
		}
	}
}

// Service is the goroutine that services the netstate.
func (n *NetState) Service() {
	// When this goroutine ends so will the TCP connection.
	defer n.Disconnect()

	// Start SendService
	go n.SendService()

	// Configure TCP QoS
	n.conn.SetKeepAlive(false)
	n.conn.SetLinger(0)
	n.conn.SetNoDelay(true)
	n.conn.SetReadBuffer(64 * 1024)
	n.conn.SetWriteBuffer(128 * 1024)
	n.conn.SetDeadline(time.Now().Add(time.Minute * 5))
	r := clientpacket.NewReader(n.conn)
	n.Log("connection from %s", n.conn.RemoteAddr().String())

	// Connection header
	if err := r.ReadConnectionHeader(); err != nil {
		n.Error("read header", err)
		return
	}

	// Game server login packet
	cp, err := r.ReadPacket()
	if err != nil {
		n.Error("waiting for game server login", err)
		return
	}
	gslp, ok := cp.(*clientpacket.GameServerLogin)
	if !ok {
		n.Error("waiting for game server login", ErrWrongPacket)
		return
	}
	account := world.AuthenticateLoginSession(gslp.Username, game.HashPassword(gslp.Password), gslp.Key)
	if account == nil {
		n.Error(fmt.Sprintf("bad login seed 0x%08X", gslp.Key), nil)
		return
	}
	n.account = account
	n.id = account.Username()

	// Character list
	n.Send(&serverpacket.CharacterList{
		Names: []string{
			account.Username(), "", "", "", "", "",
		},
	})

	// Character login
	cp, err = r.ReadPacket()
	if err != nil {
		n.Error("waiting for character login", err)
		return
	}
	_, ok = cp.(*clientpacket.CharacterLogin)
	if !ok {
		n.Error("waiting for character login", ErrWrongPacket)
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
				n.Error("writing packet", err)
				return
			}
		default:
			if pw.Buffered() > 0 {
				if err := pw.Flush(); err != nil {
					n.Error("sending packet", err)
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
			if errors.Is(err, os.ErrDeadlineExceeded) {
				n.Error("reading packet timeout", err)
				return
			}
			n.Error("reading packet", err)
			return
		}
		// 5 minute timeout, should never be hit due to client ping packets
		n.conn.SetDeadline(time.Now().Add(time.Minute * 5))

		cp := clientpacket.New(data)
		switch p := cp.(type) {
		case nil:
			n.Error("decoding packet",
				fmt.Errorf("unknown packet 0x%02X", data[0]))
		case *clientpacket.MalformedPacket:
			n.Error("decoding packet",
				fmt.Errorf("malformed packet %s", p.Serial().String()))
		case *clientpacket.UnknownPacket:
			n.Log("unknown %s packet %s", p.PType, cp.Serial().String())
			return
		case *clientpacket.UnsupportedPacket:
			n.Log("unsupported %s packet %s:\n%s", p.PType, cp.Serial().String(),
				hex.Dump(data))
		case *clientpacket.IgnoredPacket:
			// Do nothing
		default:
			handler, ok := embeddedHandlers.Get(cp.Serial())
			if !ok || handler == nil {
				// This packet is handled by the world goroutine, so forward it
				// on.
				world.SendRequest(&ClientPacketRequest{
					BaseWorldRequest: BaseWorldRequest{
						NetState: n,
					},
					Packet: cp,
				})
			} else {
				// This packet is handled inside the net state goroutine, go
				// ahead and handle it.
				handler(&PacketContext{
					NetState: n,
					Packet:   cp,
				})
			}
		}
	}
}

// SystemMessage sends a system message to the connected client. This is a
// wrapper around n.SendSpeech.
func (n *NetState) SystemMessage(fmtstr string, args ...interface{}) {
	n.Speech(nil, fmtstr, args...)
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
			X:             item.Location().X,
			Y:             item.Location().Y,
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
			X:                item.Location().X,
			Y:                item.Location().Y,
			Z:                item.Location().Z,
			Layer:            layer,
			Hue:              item.Hue(),
		})
	}
}

// sendMobile sends packets to send a mobile to the client.
func (n *NetState) sendMobile(mobile game.Mobile) {
	flags := uo.MobileFlagNone
	if mobile.IsFemale() {
		flags |= uo.MobileFlagFemale
	}
	notoriety := uo.NotorietyEnemy
	if n.m != nil {
		notoriety = n.m.GetNotorietyFor(mobile)
	}
	if mobile.IsHumanBody() {
		p := &serverpacket.EquippedMobile{
			ID:        mobile.Serial(),
			Body:      mobile.Body(),
			X:         mobile.Location().X,
			Y:         mobile.Location().Y,
			Z:         mobile.Location().Z,
			Facing:    mobile.Facing(),
			IsRunning: mobile.IsRunning(),
			Hue:       mobile.Hue(),
			Flags:     flags,
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
	} else {
		n.MoveMobile(mobile)
	}
}

// updateMobile sends a StatusBarInfo packet for the mobile.
func (n *NetState) updateMobile(mobile game.Mobile) {
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

// DropReject sends a move reject packet
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
		Gump:       uo.Gump(c.GumpGraphic()),
	})
	if c.ItemCount() > 0 {
		p := &serverpacket.Contents{}
		c.MapContents(func(item game.Item) error {
			p.Items = append(p.Items, &serverpacket.ContentsItem{
				Serial:        item.Serial(),
				Graphic:       item.BaseGraphic(),
				GraphicOffset: item.GraphicOffset(),
				Amount:        item.Amount(),
				X:             item.Location().X,
				Y:             item.Location().Y,
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
		X:             item.Location().X,
		Y:             item.Location().Y,
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
		root := c.RootParent()
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
