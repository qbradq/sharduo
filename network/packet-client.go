package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/qbradq/sharduo/core"
	"github.com/satori/go.uuid"
)

const idleTimeoutDuration = time.Minute * 5
const headerReadTimeoutDuration = time.Second * 5
const headerSize = 4
const clientReadBufferSize = 1024 * 64

// PacketClient is a client of PacketServer
type PacketClient struct {
	Conn       *net.TCPConn
	readBuffer [clientReadBufferSize]byte
	uuid       uuid.UUID
	core.Stopper
}

// NewPacketClient creates a new PacketClient
func NewPacketClient(conn *net.TCPConn) *PacketClient {
	return &PacketClient{
		Conn: conn,
		uuid: uuid.Must(uuid.NewV4()),
	}
}

// UUID returns the UUIDv4 identifier assigned to this client
func (p *PacketClient) UUID() uuid.UUID {
	return p.uuid
}

// ReadLoop is the client's packet reading and dispatching loop
func (p *PacketClient) ReadLoop() {
	// Read the connection header
	p.Conn.SetDeadline(time.Now().Add(headerReadTimeoutDuration))
	_, err := io.ReadFull(p.Conn, p.readBuffer[0:4])
	if err != nil {
		return
	}

	// Read packets
	for p.Stopping() == false {
		// Reset packet read deadline
		p.Conn.SetDeadline(time.Now().Add(idleTimeoutDuration))

		// Packet header
		_, err = io.ReadFull(p.Conn, p.readBuffer[0:1])
		if p.logError(err) || p.Stopping() {
			break
		}
		info := clientPacketInfos[p.readBuffer[0]]
		length := info.Length

		// Packet body
		if length == 0 { // Bad packet
			fmt.Printf("Client %s sent invalid packet id %2X\n",
				p.Conn.RemoteAddr().String(),
				p.readBuffer[0])
			break
		} else if length == -1 { // Dynamic length
			_, err := io.ReadFull(p.Conn, p.readBuffer[1:3])
			if p.logError(err) || p.Stopping() {
				break
			}
			length := binary.BigEndian.Uint16(p.readBuffer[1:3])
			_, err = io.ReadFull(p.Conn, p.readBuffer[3:length])
			if p.logError(err) || p.Stopping() {
				break
			}
		} else { // Fixed length
			_, err = io.ReadFull(p.Conn, p.readBuffer[1:length])
			if p.logError(err) || p.Stopping() {
				break
			}
		}

		// Packet dispatch
		if p.Stopping() {
			break
		}
		if info.Decoder != nil {
			pb := &PacketBuffer{
				Buf: p.readBuffer[0:length],
			}
			info.Decoder(pb)
		}
	}
}

// WriteLoop is the client's write goroutine
func (p *PacketClient) WriteLoop() {

}

// Stop requests a full stop of the client goroutine
func (p *PacketClient) Stop() {
	if p.Stopping() == false {
		p.Stopper.Stop()
		p.Conn.SetDeadline(time.Now())
	}
}

func (p *PacketClient) logError(err error) bool {
	if err == nil {
		return false
	}
	if noe, ok := err.(*net.OpError); ok && noe.Timeout() {
		if p.Stopping() == false {
			log.Fatalln("Client",
				p.Conn.RemoteAddr().String(),
				"disconnected due to network timeout")
		}
		return true
	}
	log.Fatalln("Client",
		p.Conn.RemoteAddr().String(),
		"disconnected due to",
		err)
	return false
}
