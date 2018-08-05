package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/qbradq/sharduo/packets/client"
	"github.com/qbradq/sharduo/packets/server"
)

const tcpReadBufferSize = 1024 * 128
const tcpWriteBufferSize = 1024 * 384
const tcpKeepAliveTimeout = time.Second * 5
const idleTimeoutDuration = time.Minute * 5
const headerReadTimeoutDuration = time.Second * 5
const headerSize = 4
const clientReadBufferSize = 1024 * 64
const clientWriteBufferSize = 1024 * 64
const clientOutboundPacketQueueDepth = 100

// packetClient is a client of PacketServer
type packetClient struct {
	Conn        *net.TCPConn
	readBuffer  [clientReadBufferSize]byte
	writeBuffer [clientWriteBufferSize]byte
	sendChannel chan server.Packet
	stopLock    *sync.Mutex
	stop        bool
}

func newPacketClient(conn *net.TCPConn) *packetClient {
	// Configure connection
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(tcpKeepAliveTimeout)
	conn.SetLinger(0)
	conn.SetNoDelay(true)
	conn.SetReadBuffer(tcpReadBufferSize)
	conn.SetWriteBuffer(tcpWriteBufferSize)

	// Create the object
	return &packetClient{
		Conn:        conn,
		sendChannel: make(chan server.Packet, clientOutboundPacketQueueDepth),
		stopLock:    new(sync.Mutex),
	}
}

// ReadLoop is the client's packet reading and dispatching loop
func (p *packetClient) ReadLoop(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// Read the connection header
	p.Conn.SetDeadline(time.Now().Add(headerReadTimeoutDuration))
	_, err := io.ReadFull(p.Conn, p.readBuffer[0:4])
	if p.logClientError(err) {
		return
	}

	// Read packets
	for {
		// Reset packet read deadline
		p.Conn.SetDeadline(time.Now().Add(idleTimeoutDuration))

		// Packet header
		_, err = io.ReadFull(p.Conn, p.readBuffer[0:1])
		if p.logClientError(err) {
			break
		}
		info := client.PacketInfos[p.readBuffer[0]]
		length := info.Length

		// Packet body
		if length == 0 { // Bad packet
			fmt.Printf("Client %s sent invalid packet id %2X\n",
				p.Conn.RemoteAddr().String(),
				p.readBuffer[0])
			break
		} else if length == -1 { // Dynamic length
			_, err := io.ReadFull(p.Conn, p.readBuffer[1:3])
			if p.logClientError(err) {
				break
			}
			length := binary.BigEndian.Uint16(p.readBuffer[1:3])
			_, err = io.ReadFull(p.Conn, p.readBuffer[3:length])
			if p.logClientError(err) {
				break
			}
		} else { // Fixed length
			_, err = io.ReadFull(p.Conn, p.readBuffer[1:length])
			if p.logClientError(err) {
				break
			}
		}

		// Packet dispatch
		if info.Decoder != nil {
			r := &client.PacketReader{
				Buf: p.readBuffer[0:length],
			}
			info.Decoder(r, p)
		}
	}
	log.Println("End of ReadLoop")
}

// WriteLoop is the client's write goroutine
func (p *packetClient) WriteLoop(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		pkt, open := <-p.sendChannel
		log.Println("WriteLoop open=", open)
		if open == false {
			break
		}
		p.Conn.SetWriteDeadline(time.Now().Add(idleTimeoutDuration))
		_, err := p.Conn.Write(pkt.Compile(p.writeBuffer[:]))
		log.Println("WriteLoop err=", err)
		if p.logClientError(err) {
			break
		}
	}
	log.Println("End of WriteLoop")
}

// Stop requests a full stop of the client goroutine
func (p *packetClient) Stop() {
	p.stopLock.Lock()
	defer p.stopLock.Unlock()

	if p.stop == false {
		p.stop = true
		p.Conn.Close()
		close(p.sendChannel)
	}
}

// PacketSend adds a server.Packet object to the queue of outbound packets. If
// the queue is full the client is disconnected and the function returns
// immediately. PacketSend will never block or panic.
func (p *packetClient) PacketSend(pkt server.Packet) {
	select {
	case p.sendChannel <- pkt:
		// Success, nothing to do
	default:
		// Channel full, bail
		p.Stop()
	}
}

func (p *packetClient) logClientError(err error) bool {
	if err == nil {
		return false
	} else if strings.Contains(err.Error(), "use of closed network connection") ||
		err == io.EOF || err == io.ErrUnexpectedEOF {
		return true
	}
	log.Println("Client",
		p.Conn.RemoteAddr().String(),
		"disconnected due to",
		err)
	return true
}
