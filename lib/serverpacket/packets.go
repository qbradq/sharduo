package serverpacket

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/qbradq/sharduo/lib/uo"
)

func padstr(w io.Writer, s string, l int) {
	var a [1024]byte
	buf := a[:l]
	copy(buf, []byte(s))
	w.Write(buf)
}

func putstr(w io.Writer, s string) {
	var b [1]byte
	w.Write([]byte(s))
	w.Write(b[:])
}

func pad(w io.Writer, l int) {
	var a [1024]byte
	buf := a[:l]
	w.Write(buf)
}

func fill(w io.Writer, v byte, l int) {
	var b [1]byte
	b[0] = v
	for i := 0; i < l; i++ {
		w.Write(b[:])
	}
}

func putbyte(w io.Writer, v byte) {
	var b [1]byte
	b[0] = v
	w.Write(b[:])
}

func putuint16(w io.Writer, v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	w.Write(b[:])
}

func putuint32(w io.Writer, v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	w.Write(b[:])
}

// Packet is the interface all server packets implement.
type Packet interface {
	// Write writes the packet data to w.
	Write(w io.Writer)
}

// ServerListEntry represents one entry in the server list.
type ServerListEntry struct {
	// Name of the server
	Name string
	// IP address of the server to ping
	IP net.IP
}

// ServerList lists all of the available game servers during login.
type ServerList struct {
	// Entries in the server list (order is important!)
	Entries []ServerListEntry
}

// Write implements the Packet interface.
func (p *ServerList) Write(w io.Writer) {
	length := 6 + 40*len(p.Entries)
	// Header
	putbyte(w, 0xa8)                     // ID
	putuint16(w, uint16(length))         // Length
	putbyte(w, 0xcc)                     // Client flags
	putuint16(w, uint16(len(p.Entries))) // Server count
	// Server list
	for idx, entry := range p.Entries {
		putuint16(w, uint16(idx)) // Server index
		padstr(w, entry.Name, 32) // Server name
		pad(w, 2)                 // Padding and timezone offset
		// The IP is backward
		putbyte(w, entry.IP.To4()[3])
		putbyte(w, entry.IP.To4()[2])
		putbyte(w, entry.IP.To4()[1])
		putbyte(w, entry.IP.To4()[0])
	}
}

// ConnectToGameServer is sent to instruct the client how to connect to a game
// server.
type ConnectToGameServer struct {
	// IP is the IP address of the server.
	IP net.IP
	// Port is the port the server listens on.
	Port uint16
	// Key is the connection key.
	Key []byte
}

// Write implements the Packet interface.
func (p *ConnectToGameServer) Write(w io.Writer) {
	putbyte(w, 0x8c) // ID
	// IP Address (right-way around)
	putbyte(w, p.IP.To4()[0])
	putbyte(w, p.IP.To4()[1])
	putbyte(w, p.IP.To4()[2])
	putbyte(w, p.IP.To4()[3])
	putuint16(w, p.Port) // Port
	w.Write(p.Key)
}

// CharacterList is sent on game server login and lists all characters on the
// account as well as the new character starting locations.
type CharacterList struct {
	// Names of all of the characters, empty string for open slots.
	Names []string
}

// Write implements the Packet interface.
func (p *CharacterList) Write(w io.Writer) {
	length := 4 + len(p.Names)*60 + 1 + 63*len(StartingLocations) + 4
	putbyte(w, 0xa9)               // ID
	putuint16(w, uint16(length))   // Length
	putbyte(w, byte(len(p.Names))) // Number of character slots
	for _, name := range p.Names {
		padstr(w, name, 30)
		pad(w, 30)
	}
	// Starting locations
	putbyte(w, byte(len(StartingLocations))) // Count
	for i, loc := range StartingLocations {
		putbyte(w, byte(i)) // Index
		padstr(w, loc.City, 31)
		padstr(w, loc.Area, 31)
	}
	// Flags
	putuint32(w, 0x000001e8)
}

// LoginComplete is sent after character login is sucessful.
type LoginComplete struct{}

// Write implements the Packet interface.
func (p *LoginComplete) Write(w io.Writer) {
	putbyte(w, 0x55) // ID
}

// EnterWorld is sent just after character login to bring them into the world.
type EnterWorld struct {
	// Player serial
	Player uo.Serial
	// Body graphic
	Body uo.Body
	// Position
	X, Y, Z int
	// Direction the player is facing and if running.
	Facing uo.Dir
	// Server dimensions
	Width, Height int
}

// Write implements the Packet interface.
func (p *EnterWorld) Write(w io.Writer) {
	putbyte(w, 0x1b) // ID
	putuint32(w, uint32(p.Player))
	pad(w, 4)
	putuint16(w, uint16(p.Body))
	putuint16(w, uint16(p.X))
	putuint16(w, uint16(p.Y))
	putbyte(w, 0)
	putbyte(w, byte(p.Z))
	putbyte(w, byte(p.Facing))
	putbyte(w, 0)
	fill(w, 0xff, 4)
	pad(w, 4)
	putuint16(w, uint16(p.Width))
	putuint16(w, uint16(p.Height))
	pad(w, 6)
}

// Version is sent to the client to request the client version of the packet.
type Version struct{}

// Write implements the Packet interface.
func (p *Version) Write(w io.Writer) {
	putbyte(w, 0xbd) // ID
	putuint16(w, 3)  // Packet length
}

// Speech is sent to the client for all kinds of speech including system
// messages and prompts.
type Speech struct {
	// Serial of the speaker
	Speaker uo.Serial
	// Body of the speaker
	Body uo.Body
	// Type of speech
	Type uo.SpeechType
	// Hue of the text
	Hue uo.Hue
	// Font of the text
	Font uo.Font
	// Name of the speaker (truncated to 30 bytes) (empty for system)
	Name string
	// Text of the message spoken
	Text string
}

// Write implements the Packet interface.
func (p *Speech) Write(w io.Writer) {
	putbyte(w, 0x1c) // ID
	putuint16(w, uint16(44+len(p.Text)+1))
	putuint32(w, uint32(p.Speaker))
	putuint16(w, uint16(p.Body))
	putbyte(w, byte(p.Type))
	putuint16(w, uint16(p.Hue))
	putuint16(w, uint16(p.Font))
	padstr(w, p.Name, 30)
	putstr(w, p.Text)
}
