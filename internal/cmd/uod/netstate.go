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
	conn      *net.TCPConn
	sendQueue chan serverpacket.Packet
	id        string
	m         game.Mobile
}

// NewNetState constructs a new NetState object.
func NewNetState(conn *net.TCPConn) *NetState {
	uuid, _ := uuid.NewRandom()
	return &NetState{
		conn:      conn,
		sendQueue: make(chan serverpacket.Packet, 128),
		id:        uuid.String(),
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

// SystemMessage sends a system message to the connected client. This is a
// wrapper around n.Send sending a serverpacket.Speech packet.
func (n *NetState) SystemMessage(fmtstr string, args ...interface{}) {
	n.Send(&serverpacket.Speech{
		Speaker: uo.SerialSystem,
		Body:    uo.BodySystem,
		Font:    uo.FontNormal,
		Hue:     1153,
		Name:    "",
		Text:    fmt.Sprintf(fmtstr, args...),
		Type:    uo.SpeechTypeSystem,
	})
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
	account := accountManager.Get(uo.NewSerialFromData(gslp.Key))
	if account == nil {
		n.Error(fmt.Sprintf("bad login seed 0x%08X", gslp.Key), nil)
		return
	}
	// TODO Account authentication
	account = accountManager.GetOrCreate(gslp.Username, game.HashPassword(gslp.Password))

	// Character list
	n.Send(&serverpacket.CharacterList{
		Names: []string{
			account.Username, "", "", "", "", "",
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

	// TODO Character load
	isFemale := world.Random().RandomBool()
	n.m = world.NewMobile(&game.BaseMobile{
		BaseObject: game.BaseObject{
			Name: gslp.Username,
			Hue:  uo.RandomSkinHue(world.Random()),
			Location: game.Location{
				X: 1607,
				Y: 1595,
				Z: 13,
			},
		},
		IsFemale:  isFemale,
		Body:      uo.GetHumanBody(isFemale),
		Notoriety: uo.NotorietyInnocent,
	})
	n.m.Equip(world.NewItem(&game.BaseItem{
		BaseObject: game.BaseObject{
			Name:     "shirt",
			ArticleA: true,
			Hue:      uo.RandomDyeHue(world.Random()),
		},
		Graphic:  0x1517,
		Wearable: true,
		Layer:    uo.LayerShirt,
	}))
	n.m.Equip(world.NewItem(&game.BaseItem{
		BaseObject: game.BaseObject{
			Name: "pants",
			Hue:  uo.RandomDyeHue(world.Random()),
		},
		Graphic:  0x152E,
		Wearable: true,
		Layer:    uo.LayerPants,
	}))

	// Request version string
	n.Send(&serverpacket.Version{})

	// Debug
	n.Send(&serverpacket.EnterWorld{
		Player: n.m.GetSerial(),
		Body:   n.m.GetBody(),
		X:      n.m.GetLocation().X,
		Y:      n.m.GetLocation().Y,
		Z:      n.m.GetLocation().Z,
		Facing: uo.DirectionSouth | uo.DirectionRunningFlag,
		Width:  7168,
		Height: 4096,
	})
	n.Send(&serverpacket.LoginComplete{})
	Broadcast("Welcome %s to Trammel Time!", n.m.GetDisplayName())
	n.Send(n.m.EquippedMobilePacket())

	n.readLoop(r)
}

// SendService is the goroutine that services the send queue.
func (n *NetState) SendService() {
	w := serverpacket.NewCompressedWriter()
	pw := bufio.NewWriterSize(n.conn, 128*1024)
	for p := range n.sendQueue {
		if err := w.Write(p, pw); err != nil {
			n.Error("writing packet", err)
			return
		}
		if err := pw.Flush(); err != nil {
			n.Error("sending packet", err)
			return
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
				fmt.Errorf("unknown packet 0x%04X", data[0]))
		case *clientpacket.MalformedPacket:
			n.Error("decoding packet",
				fmt.Errorf("malformed packet 0x%04X", p.GetID()))
		case *clientpacket.UnknownPacket:
			n.Log("unknown %s packet 0x%04X", p.PType, cp.GetID())
			return
		case *clientpacket.UnsupportedPacket:
			n.Log("unsupported %s packet 0x%04X:\n%s", p.PType, cp.GetID(),
				hex.Dump(data))
		case *clientpacket.IgnoredPacket:
			// Do nothing
		default:
			handler, ok := clientPacketFactory.Get(cp.GetID())
			if !ok || handler == nil {
				// This packet is handled by the world goroutine, so forward it
				// on.
				world.SendRequest(&ClientPacketRequest{
					BaseWorldRequest: BaseWorldRequest{
						NetState: n,
						Packet:   cp,
					},
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
