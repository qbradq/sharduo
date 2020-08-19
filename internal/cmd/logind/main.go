package logind

import (
	"errors"
	"log"
	"net"

	"github.com/qbradq/sharduo/internal/network"
	"github.com/qbradq/sharduo/lib/uo"
)

// Main is the entry point for logind.
func Main() {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 7775,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Login server listening at 0.0.0.0:7775")
	for {
		c, err := l.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(c)
	}
}

func handleConnection(c *net.TCPConn) {
	var ok bool

	cc := network.NewClientConnection(c)
	defer cc.Disconnect()
	// Don't care about the connection header
	cc.GetHeader()

	// Account login packet
	p := cc.GetPacket()
	alp, ok := p.(uo.ClientPacketAccountLogin)
	if !ok {
		cc.Error(errors.New("Expected account login packet"))
		return
	}
	log.Println("User login", alp.Username(), alp.Password())

	// Server packet buffer
	buf := network.GetBuffer()
	defer network.PutBuffer(buf)

	// Send server list
	slp := uo.NewServerPacketServerList(buf.B)
	slp.AddServer("TT LOCAL DEV", 0, 0, net.IPAddr{
		IP: net.ParseIP("127.0.0.1"),
	})
	slp.Finish()
	cc.SendPacket(slp)
	if cc.Closed() {
		return
	}
	buf.Reset()

	// Wait for select server packet
	p = cc.GetPacket()
	ssp, ok := p.(uo.ClientPacketSelectServer)
	if !ok {
		cc.Error(errors.New("Expected select server packet"))
		return
	}
	log.Println("Selected shard", ssp.ShardSelected())

	// Send connect to game server packet
	cgsp := uo.NewServerPacketConnectToServer(buf.B, net.IPAddr{
		IP: net.ParseIP("127.0.0.1"),
	}, 7777)
	cc.SendPacket(cgsp)
}
