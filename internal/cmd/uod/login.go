package uod

import (
	"bufio"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Login server listener
var loginServerListener *net.TCPListener

// Login server connections
var loginServerConnections sync.Map

// StopLoginService attempts to gracefully stop the login service.
func StopLoginService() {
	if loginServerListener != nil {
		loginServerListener.Close()
	}
}

// LoginServerMain is the entry point for the login server.
func LoginServerMain(wg *sync.WaitGroup) {
	var err error

	wg.Add(1)
	defer wg.Done()

	defaultUsername := configuration.DefaultAdminUsername
	defaultPassword := game.HashPassword(configuration.DefaultAdminPassword)

	admin := world.AuthenticateAccount(defaultUsername, defaultPassword)
	log.Println("default admin username", admin.Username())

	loginServerListener, err = net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(configuration.LoginServerAddress),
		Port: configuration.LoginServerPort,
	})
	if err != nil {
		log.Printf("error: %s", err.Error())
		return
	}
	defer loginServerListener.Close()
	log.Printf("login server listening at %s:%d\n", configuration.LoginServerAddress, configuration.LoginServerPort)

	for {
		c, err := loginServerListener.AcceptTCP()
		if err != nil {
			if !strings.Contains(err.Error(), "closed network connection") {
				log.Printf("error: %s", err.Error())
			}
			break
		}
		go handleLoginConnection(c)
	}

	loginServerListener.Close()
	loginServerConnections.Range(func(key, value interface{}) bool {
		c := key.(*net.TCPConn)
		c.Close()
		return true
	})
}

func handleLoginConnection(c *net.TCPConn) {
	var err error

	loginServerConnections.Store(c, true)
	defer loginServerConnections.Delete(c)

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
	var vMajor, vMinor, vPatch, vExtra int = 7, 0, 15, 1
	cp, err := r.ReadPacket()
	if err != nil {
		log.Println("client disconnected waiting for login seed", err)
		return
	}
	lsp, ok := cp.(*clientpacket.LoginSeed)
	if !ok {
		log.Println("client sent wrong packet waiting for login seed", cp)
		return
	}
	if lsp.VersionMajor != vMajor || lsp.VersionMinor != vMinor || lsp.VersionPatch != vPatch || lsp.VersionExtra != vExtra {
		log.Printf("bad client version %d.%d.%d.%d wanted %d.%d.%d.%d\n",
			lsp.VersionMajor, lsp.VersionMinor, lsp.VersionPatch, lsp.VersionExtra,
			vMajor, vMinor, vPatch, vExtra)
		return
	}

	// Account login
	cp, err = r.ReadPacket()
	if err != nil {
		log.Println("client disconnected waiting for account login", err)
		return
	}
	alp, ok := cp.(*clientpacket.AccountLogin)
	if !ok {
		log.Println("client sent wrong packet waiting for account login", cp)
		return
	}
	account := world.AuthenticateAccount(alp.Username, game.HashPassword(alp.Password))
	if account == nil {
		log.Println("user login failed for", alp.Username)
		ldp := &serverpacket.LoginDenied{
			Reason: uo.LoginDeniedReasonBadPass,
		}
		ldp.Write(pw)
		if err := pw.Flush(); err != nil {
			log.Println("error flushing login denied packet", err)
		}
		// Giving the client a moment to process all the network traffic. This
		// is required for ClassicUO compatibility.
		time.Sleep(time.Second * 5)
		return
	}
	log.Printf("user login successful for %s %s", account.Username(), account.Serial().String())

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
		log.Println("error flushing server list packet", err)
		return
	}

	// Select server packet
	cp, err = r.ReadPacket()
	if err != nil {
		log.Println("client disconnected waiting for select server", err)
		return
	}
	_, ok = cp.(*clientpacket.SelectServer)
	if !ok {
		log.Println("client sent wrong packet waiting for select server", cp)
		return
	}

	// Connect to game server packet
	sp = &serverpacket.ConnectToGameServer{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 7777,
		Key:  account.Serial(),
	}
	sp.Write(pw)
	if err := pw.Flush(); err != nil {
		log.Println("error flushing game server redirect", err)
		return
	}

	// Giving the client a moment to process all the network traffic. This is
	// required for ClassicUO compatibility.
	time.Sleep(time.Second * 5)

	// End of login session
}
