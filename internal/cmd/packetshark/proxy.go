package packetshark

import (
	"bytes"
	"log"
	"net"

	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

type proxy struct {
	compressed, hack8c bool
	client, server     *net.TCPConn
}

func (p *proxy) start() {
	go p.clientProxy()
	go p.serverProxy()
}

func (p *proxy) clientProxy() {
	pr := clientpacket.NewReader(p.client)

	// Packets
	for {
		cp, err := pr.Read()
		if err != nil {
			log.Println("client proxy closed because", err)
			return
		}
		if cp[0] == 0xa0 {
			p.hack8c = true
		}
		if cp[0] == 0x91 {
			p.compressed = true
		}
		_, err = p.server.Write(cp)
		if err != nil {
			log.Println("client proxy closed because", err)
			return
		}
		log.Printf(">> %#v\n", cp)
	}
}

func (p *proxy) serverProxy() {
	buf := make([]byte, 64*1024)
	compressed := bytes.NewBuffer(nil)
	decompressed := bytes.NewBuffer(nil)
	for {
		// Get next TCP packet
		n, err := p.server.Read(buf[:])
		if err != nil {
			log.Println("server proxy closed on read because", err)
			break
		}
		// Hack the 8C packet
		if p.hack8c && buf[0] == 0x8c {
			buf[5] = 7774 >> 8
			buf[6] = 7774 & 0xff
			p.hack8c = false
			log.Printf("0x8C %#v", buf[:n])
		}
		// Write back to the client
		_, err = p.client.Write(buf[:n])
		if err != nil {
			log.Println("server proxy closed on write because", err)
			return
		}
		// Handle compressed server stream
		if p.compressed {
			compressed.Write(buf[:n])
			for {
				decompressed.Reset()
				backup := append([]byte(nil), compressed.Bytes()...)
				if err := uo.HuffmanDecodePacket(compressed, decompressed); err != nil {
					// Fragmented packet
					if err == uo.ErrIncompletePacket {
						log.Println("Fragmented packet")
						compressed.Reset()
						compressed.Write(backup)
						break
					}
					// Other error
					log.Println("Server proxy closed due to error during decompression", err)
					return
				}
				// Whole packet
				log.Printf("<< %#v\n", decompressed.Bytes())
				// Read until no more compressed data
				if compressed.Len() == 0 {
					compressed.Reset()
					break
				}
			}
		} else {
			// Uncompressed login server traffic
			log.Printf("<< %#v\n", buf[:n])
		}
	}
}
