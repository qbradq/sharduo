package clientpacket

import (
	"encoding/binary"
	"log"
	"strconv"
	"strings"
)

// Packet is the interface all client packets implement.
type Packet interface {
	// GetID returns the packet ID byte.
	GetID() byte
}

// New creates a new client packet based on data.
func New(data []byte) Packet {
	var pdat []byte

	length := InfoTable[data[0]].Length
	if length == -1 {
		length = int(binary.BigEndian.Uint16(data[1:3]))
		pdat = data[3:length]
	} else if length == 0 {
		return nil
	} else {
		pdat = data[1:length]
	}
	ctor := ctorTable[data[0]]
	if ctor == nil {
		return newUnsupportedPacket(data)
	}
	return ctor(pdat)
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

// UnsupportedPacket is sent when the packet being decoded does not have a
// constructor function yet.
type UnsupportedPacket struct {
	Base
	// Bytes is a copy of the bytes of the unsupported packet for debugging and
	// may be nil.
	Bytes []byte
}

func newUnsupportedPacket(in []byte) Packet {
	return &UnsupportedPacket{
		Base:  Base{ID: in[0]},
		Bytes: append([]byte(nil), in...),
	}
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
		Base:     Base{ID: 0x80},
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
		Base:  Base{ID: 0xA0},
		Index: int(binary.BigEndian.Uint16(in[0:2])),
	}
}

// GameServerLogin is used to authenticate to the game server in clear text.
type GameServerLogin struct {
	Base
	// Account username
	Username string
	// Account password in plain-text
	Password string
	// Key given by the login server
	Key []byte
}

func newGameServerLogin(in []byte) Packet {
	return &GameServerLogin{
		Base:     Base{ID: 0x91},
		Key:      in[:4],
		Username: string(in[4:34]),
		Password: string(in[34:64]),
	}
}

// CharacterLogin is used to request a character login.
type CharacterLogin struct {
	Base
	// Character slot chosen
	Slot int
}

func newCharacterLogin(in []byte) Packet {
	return &CharacterLogin{
		Base: Base{ID: 0x5D},
		Slot: int(binary.BigEndian.Uint32(in[64:68])),
	}
}

// Version is used to communicate to the server the client's version string.
type Version struct {
	Base
	// Major version number
	Major int
	// Minor version number
	Minor int
	// Revision number
	Revision int
	// Patch number
	Patch int
}

func newVersion(in []byte) Packet {
	var n int64
	var err error

	p := &Version{
		Base: Base{ID: 0xBD},
	}

	if len(in) <= 1 {
		log.Println("Empty version string")
		return p
	}

	s := string(in[:len(in)-1])
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		log.Printf("Failed to parse version string %s due to bad format", s)
		return p
	}
	if n, err = strconv.ParseInt(parts[0], 10, 32); err != nil {
		log.Printf("Failed to parse version string %s due to %s", s, err)
	}
	p.Major = int(n)
	if n, err = strconv.ParseInt(parts[1], 10, 32); err != nil {
		log.Printf("Failed to parse version string %s due to %s", s, err)
	}
	p.Minor = int(n)
	if n, err = strconv.ParseInt(parts[2], 10, 32); err != nil {
		log.Printf("Failed to parse version string %s due to %s", s, err)
	}
	p.Revision = int(n)
	if n, err = strconv.ParseInt(parts[3], 10, 32); err != nil {
		log.Printf("Failed to parse version string %s due to %s", s, err)
	}
	p.Patch = int(n)

	return p
}
