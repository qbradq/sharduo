package uod

import (
	"log"
	"net"
	"time"

	"github.com/qbradq/sharduo/lib/clientpacket"
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
	/*
		var ok bool

		cc := network.NewClientConnection(c)
		defer cc.Disconnect()

		// Ignore the header
		cc.GetHeader()

		// Account login packet
		p := cc.GetPacket()
		gslp, ok := p.(uo.ClientPacketGameServerLogin)
		if !ok {
			cc.Error(errors.New("Expected game server login packet"))
			return
		}
		log.Println("User login", gslp.Username(), gslp.Password())

		// Server packet buffer
		buf := network.GetBuffer()
		defer network.PutBuffer(buf)

		// Respond with features packet
		ecfp := uo.NewServerPacketEnableClientFeatures(buf.B, 0x03)
		cc.SendPacket(ecfp)
		if cc.Closed() {
			return
		}
		buf.Reset()

		// Build character list packet
		clp := uo.NewServerPacketCharacterList(buf.B)
		clp.AddCharacter(gslp.Username())
		clp.FinishCharacterList()
		clp.AddStartingLocation("Haven", "The Middle of F'ing Town")
		clp.Finish(uo.FeatureFlagSiege)
		cc.SendPacket(clp)
		if cc.Closed() {
			return
		}
		buf.Reset()

		// Wait for character login packet
		p = cc.GetPacket()
		lp, ok := p.(uo.ClientPacketCharacterLogin)
		if !ok {
			cc.Error(errors.New("Expected character login packet"))
			return
		}
		log.Println("Character login", lp.CharacterSlot())
	*/
}
