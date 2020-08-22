package uod

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Map of all active netstates.
var netStates sync.Map

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

// Goroutine for handling inbound connections.
func handleConnection(c *net.TCPConn) {
	ns := NewNetState(c)
	netStates.Store(ns, true)
	ns.Service()
	netStates.Delete(ns)
}

// Broadcast sends a system-wide broadcast message to all connected clients.
func Broadcast(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	netStates.Range(func(key, value interface{}) bool {
		if n, ok := key.(*NetState); ok {
			n.Send(&serverpacket.Speech{
				Speaker: uo.SerialSystem,
				Body:    uo.BodySystem,
				Font:    uo.FontNormal,
				Hue:     1159,
				Name:    "",
				Text:    s,
				Type:    uo.SpeechTypeBroadcast,
			})

		}
		return true
	})
}
