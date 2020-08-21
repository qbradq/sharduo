package serverpacket

import (
	"encoding/binary"
	"io"
	"net"
)

func padstr(w io.Writer, s string, l int) {
	buf := make([]byte, l, l)
	copy(buf, []byte(s))
	w.Write(buf)
}

func pad(w io.Writer, l int) {
	buf := make([]byte, l, l)
	w.Write(buf)
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
	binary.Write(w, binary.BigEndian, byte(0xa8))             // ID
	binary.Write(w, binary.BigEndian, uint16(length))         // Length
	binary.Write(w, binary.BigEndian, byte(0xcc))             // Client flags
	binary.Write(w, binary.BigEndian, uint16(len(p.Entries))) // Server count
	// Server list
	for idx, entry := range p.Entries {
		binary.Write(w, binary.LittleEndian, uint16(idx)) // Server index
		padstr(w, entry.Name, 32)                         // Server name
		pad(w, 2)                                         // Padding and timezone offset
		// The IP is backward
		binary.Write(w, binary.BigEndian, byte(entry.IP.To4()[3]))
		binary.Write(w, binary.BigEndian, byte(entry.IP.To4()[2]))
		binary.Write(w, binary.BigEndian, byte(entry.IP.To4()[1]))
		binary.Write(w, binary.BigEndian, byte(entry.IP.To4()[0]))
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
	binary.Write(w, binary.BigEndian, byte(0x8c)) // ID
	// IP Address (right-way around)
	binary.Write(w, binary.BigEndian, byte(p.IP.To4()[0]))
	binary.Write(w, binary.BigEndian, byte(p.IP.To4()[1]))
	binary.Write(w, binary.BigEndian, byte(p.IP.To4()[2]))
	binary.Write(w, binary.BigEndian, byte(p.IP.To4()[3]))
	binary.Write(w, binary.BigEndian, uint16(p.Port)) // Port
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
	binary.Write(w, binary.BigEndian, byte(0xa9))         // ID
	binary.Write(w, binary.BigEndian, uint16(length))     // Length
	binary.Write(w, binary.BigEndian, byte(len(p.Names))) // Number of character slots
	for _, name := range p.Names {
		padstr(w, name, 30)
		pad(w, 30)
	}
	// Starting locations
	binary.Write(w, binary.BigEndian, byte(len(StartingLocations))) // Count
	for i, loc := range StartingLocations {
		binary.Write(w, binary.BigEndian, byte(i)) // Index
		padstr(w, loc.City, 31)
		padstr(w, loc.Area, 31)
	}
	// Flags
	binary.Write(w, binary.BigEndian, uint32(0x000001e8))
}
