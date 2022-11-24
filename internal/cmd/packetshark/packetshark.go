// Package packetshark implements the a tool used for analysing and dumping
// Ultima Online packets.
package packetshark

import (
	"log"
	"net"
)

// Main is the packetshark main loop
func Main() {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 7774,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Packet shark listening at 127.0.0.0:7774")
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Println("Stopping local packet hook because", err)
			break
		}
		ip, err := net.LookupIP("127.0.0.1")
		if err != nil {
			log.Fatal(err)
		}
		s, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   ip[0],
			Port: 7775,
		})
		if err != nil {
			log.Fatal(err)
		}
		p := &proxy{
			client: c,
			server: s,
		}
		p.start()
	}
}
