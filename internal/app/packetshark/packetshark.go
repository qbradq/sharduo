// Package packetshark implements the a tool used for analysing and dumping
// Ultima Online packets.
package packetshark

import (
	"log"
	"net"

	"github.com/qbradq/sharduo/pkg/uo"
)

// Main is the packetshark main loop
func Main() {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 2592,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Local packet hook installed at 127.0.0.1:2592")
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Println("Stopping local packet hook because", err)
			break
		}
		go installProxy(c)
	}
}

func installProxy(c *net.TCPConn) {
	ip, err := net.LookupIP("login.uosecondage.com")
	if err != nil {
		log.Fatal(err)
	}
	sc, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   ip[0],
		Port: 2593,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Installing client proxy to 127.0.0.1:2593")
	go clientProxy(c, sc)
	go serverProxy(c, sc)
}

func clientProxy(l, r *net.TCPConn) {
	pr := uo.NewClientPacketReader(l)

	// Connection header
	header, err := pr.ReadConnectionHeader()
	if err != nil {
		log.Println("Client proxy closed because", err)
		return
	}
	_, err = r.Write(header)
	if err != nil {
		log.Println("Client proxy closed because", err)
		return
	}

	// Packets
	for {
		cp, err := pr.ReadClientPacket()
		if err != nil {
			log.Println("Client proxy closed because", err)
			return
		}
		switch p := cp.(type) {
		case uo.ClientPacketInvalid:
			log.Printf("Received invalid client packet 0x%02X, closing proxy", p.Command())
			return
		}
		//log.Printf(">> 0x%02X\n", cp.Command())
		log.Printf(">> %#v\n", cp.Bytes())
		_, err = r.Write(cp.Bytes())
		if err != nil {
			log.Println("Client proxy closed because", err)
			return
		}
	}
}

func serverProxy(l, r *net.TCPConn) {
	rb := make([]byte, 1024*64, 1024*64)

	for {
		n, err := r.Read(rb[:])
		if err != nil {
			log.Println("Server proxy closed on read because", err)
			break
		}
		_, err = l.Write(rb[:n])
		if err != nil {
			log.Println("Server proxy closed on write because", err)
			break
		}
		log.Printf("<< %#v\n", rb[:n])
	}
}
