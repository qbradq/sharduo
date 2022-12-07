package uod

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Account manager
var accountManager *game.AccountManager

// Save manager
var saveManager *SaveManager

// Map of all active net states
var netStates sync.Map

var world *World

// trap is used to trap all of the system signals.
func trap(l *net.TCPListener) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT)
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
	// Initialize our data structures
	world = NewWorld()
	accountManager = game.NewAccountManager(world.Random())

	// Try to load the most recent save
	savePath := "saves"
	saveManager = NewSaveManager(world, accountManager, savePath)
	if err := saveManager.Load(); err != nil {
		log.Println("error while trying to load data stores from main goroutine", err)
		os.Exit(1)
	}

	go world.Process()
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
