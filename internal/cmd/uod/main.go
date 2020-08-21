package uod

import (
	"log"
	"net"
	"time"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
)

// Main is the entry point for uod.
func Main() {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 7777,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Game server listening at 0.0.0.0:7777")
	for {
		c, err := l.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(c)
	}
}

func handleConnection(c *net.TCPConn) {
	defer c.Close()
	c.SetKeepAlive(false)
	c.SetLinger(0)
	c.SetNoDelay(true)
	c.SetReadBuffer(64 * 1024)
	c.SetWriteBuffer(128 * 1024)
	c.SetDeadline(time.Now().Add(time.Minute * 15))
	r := clientpacket.NewReader(c)
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
	if err := w.Write(clp, c); err != nil {
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
