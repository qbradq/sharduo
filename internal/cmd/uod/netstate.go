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
	conn                *net.TCPConn
	sendQueue           chan serverpacket.Packet
	id                  string
	m                   game.Mobile
	lastWalkRequestTime int64
}

// NewNetState constructs a new NetState object.
func NewNetState(conn *net.TCPConn) *NetState {
	uuid, _ := uuid.NewRandom()
	return &NetState{
		conn:      conn,
		sendQueue: make(chan serverpacket.Packet, 1024*16),
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
	account := world.AuthenticateLoginSession(gslp.Username, game.HashPassword(gslp.Password), gslp.Key)
	if account == nil {
		n.Error(fmt.Sprintf("bad login seed 0x%08X", gslp.Key), nil)
		return
	}
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
			if pw.Size() > 0 {
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

// SendItem implements the game.NetState interface.
func (n *NetState) SendItem(i game.Item) {
	var layer uo.Layer
	if w, ok := i.(game.Wearable); ok {
		layer = w.Layer()
	}
	n.Send(&serverpacket.ObjectInfo{
		Serial:  i.Serial(),
		Graphic: i.Graphic(),
		Amount:  i.Amount(),
		X:       i.Location().X,
		Y:       i.Location().Y,
		Z:       i.Location().Z,
		Layer:   layer,
		Hue:     i.Hue(),
	})
}

// RemoveObject implements the game.NetState interface.
func (n *NetState) RemoveObject(o game.Object) {
	n.Send(&serverpacket.DeleteObject{
		Serial: o.Serial(),
	})
}
