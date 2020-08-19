package clientpacket

import (
	"encoding/binary"
)

// Packet is the interface all client packets implement.
type Packet interface {
	// GetID returns the packet ID byte.
	GetID() byte
}

// New creates a new based on data.
func New(ID byte, length int, data []byte) Packet {
	ctor := ctorTable[ID]
	if ctor == nil {
		return nil
	}
	return ctor(data)
}

// Base is the base struct for all client packets.
type Base struct {
	// ID is the packet ID byte.
	ID byte
}

// GetID implements the Packet interface.
func (p *Base) GetID() byte {
	return p.ID
}

// AccountLogin is the first packet sent to the login server and attempts to
// authenticate with a clear-text username and password o_O
type AccountLogin struct {
	Base
	// Account username
	Username string
	// Account password in plain-text
	Password string
}

func newAccountLogin(in []byte) Packet {
	return &AccountLogin{
		Base: Base{
			ID: 0x80,
		},
		Username: string(in[0:30]),
		Password: string(in[30:60]),
	}
}

// SelectServer is used during the login process to request the connection
// details of one of the servers on the list.
type SelectServer struct {
	Base
	// Index is the index of the server on the list.
	Index int
}

func newSelectServer(in []byte) Packet {
	return &SelectServer{
		Base: Base{
			ID: 0xA0,
		},
		Index: int(binary.LittleEndian.Uint16(in[0:2])),
	}
}
