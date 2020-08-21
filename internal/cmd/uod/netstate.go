package uod

import (
	"log"
	"net"
	"time"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
)

// NetState manages the network state of a single connection.
type NetState struct {
	conn *net.TCPConn
}

// NewNetState constructs a new NetState object.
func NewNetState(conn *net.TCPConn) *NetState {
	return &NetState{
		conn: conn,
	}
}

// Service is the goroutine that services the netstate.
func (n *NetState) Service() {
	defer n.conn.Close()
	n.conn.SetKeepAlive(false)
	n.conn.SetLinger(0)
	n.conn.SetNoDelay(true)
	n.conn.SetReadBuffer(64 * 1024)
	n.conn.SetWriteBuffer(128 * 1024)
	n.conn.SetDeadline(time.Now().Add(time.Minute * 15))
	r := clientpacket.NewReader(n.conn)
	w := serverpacket.NewCompressedWriter()

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
	clp := &serverpacket.CharacterList{
		Names: []string{
			gslp.Username, "", "", "", "", "",
		},
	}
	if err := w.Write(clp, n.conn); err != nil {
		log.Println("Client disconnect writing character list due to", err)
		return
	}

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
}
