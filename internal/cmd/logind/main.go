package logind

import (
	"bufio"
	"log"
	"net"
	"time"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
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
	var err error

	// Setup QoS options
	defer c.Close()
	c.SetKeepAlive(false)
	c.SetLinger(0)
	c.SetNoDelay(true)
	c.SetReadBuffer(64 * 1024)
	c.SetWriteBuffer(64 * 1024)
	c.SetDeadline(time.Now().Add(time.Minute * 15))
	r := clientpacket.NewReader(c)

	// Packet writer
	pw := bufio.NewWriterSize(c, 64*1024)

	// Login seed packet
	cp, err := r.ReadPacket()
	if err != nil {
		log.Println("Client disconnected waiting for login seed", err)
		return
	}
	lsp, ok := cp.(*clientpacket.LoginSeed)
	if !ok {
		log.Println("Client sent wrong packet waiting for login seed", cp)
		return
	}
	if lsp.VersionMajor != 7 || lsp.VersionMinor != 0 || lsp.VersionPatch != 15 || lsp.VersionExtra != 1 {
		log.Printf("Bad client version %d.%d.%d.%d\n", lsp.VersionMajor, lsp.VersionMinor, lsp.VersionPatch, lsp.VersionExtra)
		return
	}

	// Account login packet
	cp, err = r.ReadPacket()
	if err != nil {
		log.Println("Client disconnected waiting for account login", err)
		return
	}
	alp, ok := cp.(*clientpacket.AccountLogin)
	if !ok {
		log.Println("Client sent wrong packet waiting for account login", cp)
		return
	}
	log.Println("User login", alp.Username, alp.Password)

	// Server list packet
	var sp serverpacket.Packet
	sp = &serverpacket.ServerList{
		Entries: []serverpacket.ServerListEntry{
			{
				Name: "LOCAL DEV",
				IP:   net.ParseIP("127.0.0.1"),
			},
		},
	}
	sp.Write(pw)
	if err := pw.Flush(); err != nil {
		log.Println("Error flushing server list packet", err)
		return
	}

	// Select server packet
	cp, err = r.ReadPacket()
	if err != nil {
		log.Println("Client disconnected waiting for select server", err)
		return
	}
	ssp, ok := cp.(*clientpacket.SelectServer)
	if !ok {
		log.Println("Client sent wrong packet waiting for select server", cp)
		return
	}
	log.Println("Selected server", ssp.Index)

	// Connect to game server packet
	sp = &serverpacket.ConnectToGameServer{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 7777,
		Key:  []byte{0xBA, 0xAD, 0xF0, 0x0D},
	}
	sp.Write(pw)
	if err := pw.Flush(); err != nil {
		log.Println("Error flushing game server redirect", err)
		return
	}

	// Giving the client a moment to process the redirect packet. This is needed
	// for ClassicUO compatibility.
	time.Sleep(time.Second * 5)

	// End of login session
	return
}

/*
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
*/
