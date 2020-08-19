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
