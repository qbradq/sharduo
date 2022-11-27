package uod

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Account manager
var accountManager *game.AccountManager

// Collection of all object serials
var serialManager *uo.SerialManager

// Map of all active netstates.
var netStates sync.Map

// Main is the entry point for uod.
func Main() {
	// Startup
	dataPath := "data"
	if err := os.MkdirAll(dataPath, 0777); err != nil {
		log.Fatal(err)
		return
	}
	accountManager = game.NewAccountManager(path.Join(dataPath, "accounts.json"))
	serialManager = uo.NewSerialManager()

	go LoginServerMain()

	ipstr := "127.0.0.1"
	port := 7777

	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(ipstr),
		Port: port,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("game server listening at %s:%d\n", ipstr, port)
	for {
		c, err := l.AcceptTCP()
		if err != nil {
			log.Fatal(err)
			return
		}
		go handleConnection(c)
	}
}

// LoginServerMain is the entry point for the login server.
func LoginServerMain() {
	ipstr := "127.0.0.1"
	port := 7775

	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(ipstr),
		Port: port,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("login server listening at %s:%d\n", ipstr, port)
	for {
		c, err := l.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		go handleLoginConnection(c)
	}
}

func handleLoginConnection(c *net.TCPConn) {
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
	pwh := sha256.Sum256([]byte(alp.Password))
	account := accountManager.GetOrCreate(alp.Username, string(pwh[:]))
	if account == nil {
		log.Println("user login failed for", alp.Username)
		ldp := &serverpacket.LoginDenied{
			Reason: uo.LoginDeniedReasonBadPass,
		}
		ldp.Write(pw)
		if err := pw.Flush(); err != nil {
			log.Println("error flushing login denied packet", err)
		}
		return
	}
	log.Println("user login successful for", account.Username)

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
	ssp, ok := cp.(*clientpacket.SelectServer)
	if !ok {
		log.Println("client sent wrong packet waiting for select server", cp)
		return
	}
	log.Println("selected server", ssp.Index)

	// Connect to game server packet
	sp = &serverpacket.ConnectToGameServer{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 7777,
		Key:  []byte{0xBA, 0xAD, 0xF0, 0x0D},
	}
	sp.Write(pw)
	if err := pw.Flush(); err != nil {
		log.Println("error flushing game server redirect", err)
		return
	}

	// Giving the client a moment to process the redirect packet. This is needed
	// for ClassicUO compatibility.
	time.Sleep(time.Second * 5)

	// End of login session
}

// Goroutine for handling inbound connections.
func handleConnection(c *net.TCPConn) {
	ns := NewNetState(c)
	netStates.Store(ns, true)
	ns.Service()
	netStates.Delete(ns)
}

// Broadcast sends a system-wide broadcast message to all connected clients.
func Broadcast(format string, args ...interface{}) {
	s := "System Broadcast: " + fmt.Sprintf(format, args...)
	netStates.Range(func(key, value interface{}) bool {
		n := key.(*NetState)
		n.SystemMessage(s)
		return true
	})
}

// GlobalChat sends a global chat message to all connected clients.
func GlobalChat(who, text string) {
	s := fmt.Sprintf("%s: %s", who, text)
	netStates.Range(func(key, value interface{}) bool {
		n := key.(*NetState)
		n.Send(&serverpacket.Speech{
			Speaker: uo.SerialSystem,
			Body:    uo.BodySystem,
			Font:    uo.FontNormal,
			Hue:     1166,
			Name:    "",
			Text:    s,
			Type:    uo.SpeechTypeSystem,
		})
		return true
	})
}
