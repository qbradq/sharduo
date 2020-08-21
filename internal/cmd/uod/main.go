package uod

import (
	"log"
	"net"
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
	ns := NewNetState(c)
	go ns.Service()
}
