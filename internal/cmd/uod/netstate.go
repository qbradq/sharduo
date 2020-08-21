package uod

import (
	"encoding/hex"
	"errors"
	"log"
	"net"
	"os"
	"time"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// NetState manages the network state of a single connection.
type NetState struct {
	conn      *net.TCPConn
	sendQueue chan serverpacket.Packet
}

// NewNetState constructs a new NetState object.
func NewNetState(conn *net.TCPConn) *NetState {
	return &NetState{
		conn:      conn,
		sendQueue: make(chan serverpacket.Packet, 128),
	}
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
	n.conn.Close()
	close(n.sendQueue)
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
		log.Println("Client disconnected during header due to", err)
		return
	}

	// Game server login packet
	cp, err := r.ReadPacket()
	if err != nil {
		log.Println("Client disconnected waiting for game server login", err)
		return
	}
	gslp, ok := cp.(*clientpacket.GameServerLogin)
	if !ok {
		log.Println("Client sent wrong packet waiting for game server login")
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
		log.Println("Client disconnected waiting for character login", err)
		return
	}
	clrp, ok := cp.(*clientpacket.CharacterLogin)
	if !ok {
		log.Println("Client sent wrong packet waiting for character login")
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

	n.readLoop(r)
}

// SendService is the goroutine that services the send queue.
func (n *NetState) SendService() {
	w := serverpacket.NewCompressedWriter()
	for p := range n.sendQueue {
		if err := w.Write(p, n.conn); err != nil {
			log.Println("Client disconnected due to send error", err)
			n.conn.Close()
			return
		}
	}
}

func (n *NetState) readLoop(r *clientpacket.Reader) {
	for {
		data, err := r.Read()
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				log.Println("Client disconnected due to read timeout", err)
				return
			}
			// Just wait for more data
			log.Println("Client disconnected during read due to", err)
			return
		}
		// 5 minute timeout, should never be hit due to client ping packets
		n.conn.SetDeadline(time.Now().Add(time.Minute * 5))

		cp := clientpacket.New(data)
		handler := PacketHandlerTable[cp.GetID()]
		if handler == nil {
			log.Printf("Unhandled client packet 0x%02X:\n%s", cp.GetID(),
				hex.Dump(data))
		} else {
			handler(n, cp)
		}
	}
}
