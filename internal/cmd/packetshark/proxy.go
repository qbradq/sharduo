package packetshark

import (
	"bytes"
	"io"
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

	// Connection header
	err := pr.ReadConnectionHeader()
	if err != nil {
		log.Println("Client proxy closed because", err)
		return
	}
	_, err = p.server.Write(pr.Header)
	if err != nil {
		log.Println("Client proxy closed because", err)
		return
	}

	// Packets
	for {
		cp, err := pr.Read()
		if err != nil {
			log.Println("Client proxy closed because", err)
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
			log.Println("Client proxy closed because", err)
			return
		}
		log.Printf(">> %#v\n", cp)
	}
}

func (p *proxy) serverProxy() {
	buf := make([]byte, 64*1024, 64*1024)
	leftbuf := make([]byte, 64*1024, 64*1024)
	readbuf := bytes.NewBuffer(nil)
	dcbuf := bytes.NewBuffer(nil)
	for {
		// Get next TCP packet
		n, err := p.server.Read(buf[:])
		if err != nil {
			log.Println("Server proxy closed on read because", err)
			break
		}
		// Hack the 8C packet
		if p.hack8c && buf[0] == 0x8c {
			buf[5] = 2592 >> 8
			buf[6] = 2592 & 0xff
			p.hack8c = false
			log.Printf("0x8C %#v", buf[:n])
		}
		// Write back to the client
		_, err = p.client.Write(buf[:n])
		if err != nil {
			log.Println("Server proxy closed on write because", err)
			return
		}
		// Handle compressed server stream
		if p.compressed {
			copy(leftbuf, readbuf.Bytes())
			left := leftbuf[:readbuf.Len()]
			if _, err := readbuf.Write(buf[:n]); err != nil {
				log.Println("Server proxy closed on write to memory buffer because", err)
				return
			}
			for readbuf.Len() > 0 {
				dcbuf.Reset()
				if err := uo.HuffmanDecodePacket(readbuf, dcbuf); err != nil {
					// Fragmented packet
					if err == io.EOF {
						log.Println("Start of fragmented packet")
						if _, err := readbuf.Write(left); err != nil {
							log.Println("Server proxy closed on write to memory buffer because", err)
							return
						}
						break
					}
					log.Println("Server proxy closed due to error during decompression", err)
					return
				}
				if readbuf.Len() == 0 {
					log.Printf("<< %#v\n", dcbuf.Bytes())
				}
			}
			readbuf.Reset()
		} else {
			// Uncompressed
			log.Printf("<< %#v\n", buf[:n])
		}
	}
}

func (p *proxy) serverProxyOld() {
	rb := make([]byte, 1024*64, 1024*64)

	for {
		n, err := p.server.Read(rb[:])
		if err != nil {
			log.Println("Server proxy closed on read because", err)
			break
		}
		ob := rb[:n]
		if p.hack8c && rb[0] == 0x8c {
			ob = []byte{
				0x8c,         // Packet ID
				127, 0, 0, 1, // IP Address
				(2592 >> 8), (2592 & 0xff), // Port
				rb[7], rb[8], rb[9], rb[10], // Key
			}
			p.hack8c = false
		}
		_, err = p.client.Write(ob)
		if err != nil {
			log.Println("Server proxy closed on write because", err)
			return
		}
		if p.compressed {
			out := bytes.NewBuffer(nil)
			cb := bytes.NewReader(ob)
			for {
				err := uo.HuffmanDecodePacket(cb, out)
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Println("Server proxy closed because", err)
					return
				}
				log.Printf("<< %#v\n", out.Bytes())
			}
		} else {
			log.Printf("<< %#v\n", ob)
		}
	}
}
