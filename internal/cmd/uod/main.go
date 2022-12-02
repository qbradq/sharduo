package uod

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Account manager
var accountManager *game.AccountManager

// Object manager
var objectManager *game.ObjectManager

// Save manager
var saveManager *SaveManager

// Map of all active netstates.
var netStates sync.Map

// trap is used to trap all of the system signals.
func trap(l *net.TCPListener) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGKILL)
	go func() {
		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGQUIT {
			// Close the main listener and let the graceful shutdown save the
			// data stores.
			l.Close()
		} else if err := saveManager.Save(); err != nil {
			// Last-ditch save attempt failed
			log.Fatal("error while trying to save data stores from signal handler", err)
		}
	}()
}

// Main is the entry point for uod.
func Main() {
	savePath := "saves"
	saveManager = NewSaveManager(savePath)
	if err := saveManager.Load(); err != nil {
		log.Println("error while trying to load data stores from main goroutine", err)
		os.Exit(1)
	}

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
	trap(l)
	for {
		c, err := l.AcceptTCP()
		if err != nil {
			if err := saveManager.Save(); err != nil {
				log.Println("error while trying to save data stores from main goroutine", err)
				os.Exit(1)
			}
			if errors.Is(err, io.EOF) {
				break
			} else {
				log.Fatal(err)
				return
			}
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
	account := accountManager.GetOrCreate(alp.Username, game.HashPassword(alp.Password))
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
	log.Printf("user login successful for %s 0x%08X", account.Username, account.GetSerial())

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
		Key:  account.GetSerial().Data(),
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
