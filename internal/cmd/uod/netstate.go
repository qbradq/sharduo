package uod

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
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
	log.Printf("%s:error:at %s:%s", n.id, where, err.Error())
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
	log.Println("User login", gslp.Username, gslp.Password)

	// Character list
	n.Send(&serverpacket.CharacterList{
		Names: []string{
			gslp.Username, "", "", "", "", "",
		},
	})

	// Character login
	cp, err = r.ReadPacket()
	if err != nil {
		n.Error("waiting for character login", err)
		return
	}
	clrp, ok := cp.(*clientpacket.CharacterLogin)
	if !ok {
		n.Error("waiting for character login", ErrWrongPacket)
		return
	}
	log.Printf("Character login request slot 0x%08X", clrp.Slot)

	// Request version string
	n.Send(&serverpacket.Version{})

	// Debug
	n.Send(&serverpacket.EnterWorld{
		Player: 0x00000047,
		Body:   400,
		X:      1607,
		Y:      1595,
		Z:      13,
		Facing: uo.DirSouth | uo.DirRunningFlag,
		Width:  7168,
		Height: 4096,
	})
	n.Send(&serverpacket.LoginComplete{})
	n.Send(&serverpacket.Speech{
		Speaker: uo.SerialSystem,
		Body:    uo.BodySystem,
		Font:    uo.FontNormal,
		Hue:     500,
		Name:    "",
		Text:    "Welcome to Trammie Time!",
		Type:    uo.SpeechTypeSystem,
	})

	n.readLoop(r)
}

// SendService is the goroutine that services the send queue.
func (n *NetState) SendService() {
	w := serverpacket.NewCompressedWriter()
	for p := range n.sendQueue {
		if err := w.Write(p, n.conn); err != nil {
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
		handler := PacketHandlerTable[cp.GetID()]
		if handler == nil {
			n.Log("Unhandled client packet 0x%02X:\n%s", cp.GetID(),
				hex.Dump(data))
		} else {
			handler(n, cp)
		}
	}
}
