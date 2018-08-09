package network

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/packets/client"
	"github.com/qbradq/sharduo/internal/packets/server"
)

const tcpReadBufferSize = 1024 * 128
const tcpWriteBufferSize = 1024 * 384
const tcpKeepAliveTimeout = time.Second * 5
const idleTimeoutDuration = time.Minute * 5
const headerReadTimeoutDuration = time.Second * 5
const headerSize = 4
const clientReadBufferSize = 1024 * 64
const clientWriteBufferSize = 1024 * 64
const clientCompressionBufferSize = clientWriteBufferSize * 2
const clientOutboundPacketQueueDepth = 100

// packetClient is a client of PacketServer
type packetClient struct {
	Conn              *net.TCPConn
	readBuffer        []byte
	writeBuffer       []byte
	compressionBuffer []byte
	sendChannel       chan server.Compiler
	stopLock          *sync.Mutex
	stop              bool
	state             *server.NetState
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
	v := &packetClient{
		Conn:              conn,
		readBuffer:        make([]byte, clientReadBufferSize, clientReadBufferSize),
		writeBuffer:       make([]byte, 0, clientWriteBufferSize),
		compressionBuffer: make([]byte, 0, clientCompressionBufferSize),
		sendChannel:       make(chan server.Compiler, clientOutboundPacketQueueDepth),
		stopLock:          new(sync.Mutex),
	}
	v.state = server.NewNetState(v)
	return v
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
			log.Printf("Client %s sent invalid packet id %2X\n",
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

		//log.Printf("Client Packet: %#v\n", p.readBuffer[0:length])

		// Packet dispatch
		if info.Decoder == nil {
			log.Printf("Client %s sent unhandled packet 0x%02X",
				p.Conn.RemoteAddr().String(),
				p.readBuffer[0])
		} else if info.IsLoginPacket == false && p.state.Authenticated() == false {
			log.Printf("Client %s disconnected because it sent packet 0x%02X before sending an authentication packet (0x91)",
				p.Conn.RemoteAddr().String(),
				p.readBuffer[0])
			break
		} else {
			r := &client.PacketReader{
				Buf: p.readBuffer[0:length],
			}
			info.Decoder(r, p.state)
		}
	}
}

// WriteLoop is the client's write goroutine
func (p *packetClient) WriteLoop(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		pkt, open := <-p.sendChannel
		if open == false {
			break
		}
		p.Conn.SetWriteDeadline(time.Now().Add(idleTimeoutDuration))
		p.writeBuffer = p.writeBuffer[:0]
		w := &server.PacketWriter{
			Buf: p.writeBuffer,
		}
		pkt.Compile(w)
		// log.Printf("Server Packet: %d %#v\n", len(w.Buf), w.Buf)
		var err error
		if p.state.CompressOutput() {
			p.compressionBuffer = compressUOHuffman(w.Buf, p.compressionBuffer)
			_, err = p.Conn.Write(p.compressionBuffer)
		} else {
			_, err = p.Conn.Write(w.Buf)
		}
		if p.logClientError(err) {
			break
		}
	}
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
func (p *packetClient) PacketSend(pkt server.Compiler) {
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
