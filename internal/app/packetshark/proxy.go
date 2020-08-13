package packetshark

import (
	"log"
	"net"

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
	pr := uo.NewClientPacketReader(p.client)

	// Connection header
	header, err := pr.ReadConnectionHeader()
	if err != nil {
		log.Println("Client proxy closed because", err)
		return
	}
	_, err = p.server.Write(header)
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
		switch pkt := cp.(type) {
		case uo.ClientPacketInvalid:
			log.Printf("Received invalid client packet 0x%02X, closing proxy", pkt.Command())
			return
		case uo.ClientPacketSelectServer:
			p.hack8c = true
		case uo.ClientPacketGameServerLogin:
			p.compressed = true
		}
		//log.Printf(">> 0x%02X\n", cp.Command())
		log.Printf(">> %#v\n", cp.Bytes())
		_, err = p.server.Write(cp.Bytes())
		if err != nil {
			log.Println("Client proxy closed because", err)
			return
		}
	}
}

func (p *proxy) serverProxy() {
	rb := make([]byte, 1024*64, 1024*64)

	for {
		n, err := p.server.Read(rb[:])
		if err != nil {
			log.Println("Server proxy closed on read because", err)
			break
		}
		ob := rb[:n]
		if p.hack8c && rb[0] == 0x8c {
			p.hack8c = false
			pkt := uo.NewServerPacketConnectToServer(make([]byte, 0, 16),
				net.IPAddr{
					IP: net.ParseIP("127.0.0.1"),
				},
				2592)
			ob = pkt.Bytes()
			copy(ob[7:], rb[7:])
		}
		_, err = p.client.Write(ob)
		if err != nil {
			log.Println("Server proxy closed on write because", err)
			return
		}
		if p.compressed {
			cb := make([]byte, 0, 128*1024)
			for len(ob) > 0 {
				cb := cb[:0]
				n, printable := uo.HuffmanDecodePacket(ob, cb)
				if n == 0 {
					log.Println("Server proxy closed because of fragmented huffman packet")
					return
				}
				log.Printf("<< %#v\n", printable)
				ob = ob[n:]
			}
		} else {
			log.Printf("<< %#v\n", ob)
		}
	}
}
