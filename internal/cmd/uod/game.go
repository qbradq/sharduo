package uod

import (
	"fmt"
	"log"
	"net"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/qbradq/sharduo/internal/configuration"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Map of all active net states
var gameNetStates sync.Map

// Listener for the game service
var gameServerListener *net.TCPListener

// StopGameService attempts to gracefully shut down the game service
func StopGameService() {
	if gameServerListener != nil {
		gameServerListener.Close()
	}
}

func GameServerMain(wg *sync.WaitGroup) {
	var err error

	defer func() {
		if p := recover(); p != nil {
			log.Printf("panic: %v\n%s\n", p, debug.Stack())
			panic(p)
		}
	}()
	defer wg.Done()

	gameServerListener, err = net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(configuration.GameServerAddress),
		Port: configuration.GameServerPort,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("info: game server listening at %s:%d\n", configuration.GameServerAddress, configuration.GameServerPort)

	for {
		c, err := gameServerListener.AcceptTCP()
		if err != nil {
			if !strings.Contains(err.Error(), "closed network connection") {
				log.Printf("error: %s", err.Error())
			}
			break
		}
		// IP blacklist
		s := strings.Split(c.RemoteAddr().String(), ":")[0]
		a := net.ParseIP(s)
		if a == nil {
			c.Close()
			continue
		}
		if blacklist.Match(a) {
			c.Close()
			log.Printf("info: blacklisted connection to game server from %s", s)
			continue
		}
		go handleGameConnection(c)
	}

	gameServerListener.Close()
	gameNetStates.Range(func(key, value interface{}) bool {
		ns := key.(*NetState)
		ns.Disconnect()
		return true
	})
}

// Goroutine for handling inbound connections.
func handleGameConnection(c *net.TCPConn) {
	var ns *NetState

	defer func() {
		if p := recover(); p != nil {
			log.Printf("panic: %v\n%s\n", p, debug.Stack())
		}
	}()
	defer ns.Disconnect()

	ns = NewNetState(c)
	gameNetStates.Store(ns, true)
	ns.Service()
	m := ns.m
	ns.m = nil
	if m != nil {
		m.SetNetState(nil)
		world.SendRequest(&CharacterLogoutRequest{
			Mobile: m,
		})
	}
}

// Executes the update method on all net states in the numbered update group.
func UpdateNetStates(group int) {
	gameNetStates.Range(func(key, value interface{}) bool {
		n := key.(*NetState)
		if n.updateGroup == group {
			n.Update()
		}
		return true
	})
}

// BroadcastPacket sends a packet to every connected net state with an attached
// mobile.
func BroadcastPacket(p serverpacket.Packet) {
	gameNetStates.Range(func(key, value interface{}) bool {
		n := key.(*NetState)
		n.Send(p)
		return true
	})
}

// Broadcast sends a system-wide broadcast message to all connected clients.
func Broadcast(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	world.BroadcastPacket(&serverpacket.Speech{
		Speaker: uo.SerialSystem,
		Body:    uo.BodySystem,
		Font:    uo.FontNormal,
		Hue:     1166,
		Name:    "SYSTEM",
		Text:    s,
		Type:    uo.SpeechTypeSystem,
	})
}

// GlobalChat sends a global chat message to all connected clients.
func GlobalChat(hue uo.Hue, who, text string) {
	s := fmt.Sprintf("%s: %s", who, text)
	gameNetStates.Range(func(key, value interface{}) bool {
		n := key.(*NetState)
		n.Send(&serverpacket.Speech{
			Speaker: uo.SerialSystem,
			Body:    uo.BodySystem,
			Font:    uo.FontNormal,
			Hue:     hue,
			Name:    "",
			Text:    s,
			Type:    uo.SpeechTypeSystem,
		})
		return true
	})
}
