package uod

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/qbradq/sharduo/lib/uo"
)

// ClientConnection represents a client connected to a server and manages the
// entire lifecycle of that connection.
type ClientConnection struct {
	conn   net.Conn
	cpr    *uo.ClientPacketReader
	spw    *uo.ServerPacketWriter
	closed bool
	header []byte
}

// NewClientConnection creates a new ClientConnection for the given conn and
// configures conn with our QoS parameters.
func NewClientConnection(conn *net.TCPConn) *ClientConnection {
	conn.SetKeepAlive(false)
	conn.SetLinger(0)
	conn.SetNoDelay(true)
	conn.SetReadBuffer(64 * 1024)
	conn.SetWriteBuffer(128 * 1024)
	conn.SetDeadline(time.Now().Add(time.Minute * 5))
	return &ClientConnection{
		conn: conn,
		cpr:  uo.NewClientPacketReader(conn),
		spw:  uo.NewServerPacketWriter(conn),
	}
}

// Log is a conveniance function to emit a log line as the client.
func (c *ClientConnection) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("log:%v: %s", c.conn.LocalAddr().String(), msg)
}

// Error is a conveniance function to emit an error log line as the client.
func (c *ClientConnection) Error(err error) {
	log.Printf("error:%v: %v", c.conn.LocalAddr().String(), err)
}

// Closed returns true if the connection is closed.
func (c *ClientConnection) Closed() bool {
	return c.closed
}

// GetHeader returns the connection header, reading if needed, returning nil
// on error or end of connection.
func (c *ClientConnection) GetHeader() []byte {
	var err error

	if c.header == nil {
		if c.header, err = c.cpr.ReadConnectionHeader(); err != nil {
			c.doReadWriteError(err)
			c.Disconnect()
			return nil
		}
	}
	return c.header
}

// GetPacket returns the next packet from the client, or nil on closed
// connection.
func (c *ClientConnection) GetPacket() uo.ClientPacket {
	var err error

	if c.closed {
		return nil
	}
	if c.header == nil {
		if c.header, err = c.cpr.ReadConnectionHeader(); err != nil {
			c.doReadWriteError(err)
			c.Disconnect()
			return nil
		}
	}
	p, err := c.cpr.ReadClientPacket()
	if err != nil {
		c.doReadWriteError(err)
		c.Disconnect()
		return nil
	}
	c.conn.(*net.TCPConn).SetDeadline(time.Now().Add(time.Minute * 5))
	return p
}

// Disconnect disconnects the client connection.
func (c *ClientConnection) Disconnect() {
	if !c.closed && c.conn != nil {
		c.conn.Close()
		c.closed = true
	}
}

// SendPacket sends packet p to the client.
func (c *ClientConnection) SendPacket(p uo.ServerPacket) {
	err := c.spw.WritePacket(p)
	if err != nil {
		c.doReadWriteError(err)
		c.Disconnect()
	}
}

func (c *ClientConnection) doReadWriteError(err error) {
	if errors.Is(err, os.ErrDeadlineExceeded) {
		c.Log("read/write timeout")
	} else {
		c.Error(err)
	}
}
