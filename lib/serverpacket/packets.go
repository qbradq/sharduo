package serverpacket

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/qbradq/sharduo/lib/uo"
)

// Writes a boolean value
func putbool(w io.Writer, v bool) {
	var b [1]byte
	if v {
		b[0] = 1
	} else {
		b[0] = 0
	}
	w.Write(b[:])
}

// Writes a null-terminated string
func putstr(w io.Writer, s string) {
	var b [1]byte
	w.Write([]byte(s))
	w.Write(b[:])
}

// Writes a fixed-length string
func putstrn(w io.Writer, s string, n int) {
	var b = make([]byte, n)
	copy(b, s)
	w.Write(b)
}

func pad(w io.Writer, l int) {
	var buf [1024]byte
	w.Write(buf[:l])
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
		putuint16(w, uint16(idx))  // Server index
		putstrn(w, entry.Name, 32) // Server name
		pad(w, 2)                  // Padding and timezone offset
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
		putstrn(w, name, 30)
		pad(w, 30)
	}
	// Starting locations
	putbyte(w, byte(len(StartingLocations))) // Count
	for i, loc := range StartingLocations {
		putbyte(w, byte(i)) // Index
		putstrn(w, loc.City, 31)
		putstrn(w, loc.Area, 31)
	}
	// Flags
	putuint32(w, 0x000001e8)
}

// LoginComplete is sent after character login is successful.
type LoginComplete struct{}

// Write implements the Packet interface.
func (p *LoginComplete) Write(w io.Writer) {
	putbyte(w, 0x55) // ID
}

// LoginDenied is sent when character login is denied for any reason.
type LoginDenied struct {
	// The reason for the login denial
	Reason uo.LoginDeniedReason
}

// Write implements the Packet interface.
func (p *LoginDenied) Write(w io.Writer) {
	putbyte(w, 0x82) // ID
	putbyte(w, byte(p.Reason))
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
	Facing uo.Direction
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
	putstrn(w, p.Name, 30)
	putstr(w, p.Text)
}

// Ping is sent to the client in response to a client ping packet.
type Ping struct {
	// Key byte of the client ping request
	Key byte
}

// Write implements the Packet interface.
func (p *Ping) Write(w io.Writer) {
	putbyte(w, 0x73)  // ID
	putbyte(w, p.Key) // Key
}

// ClientViewRange sets the client's view range
type ClientViewRange struct {
	// The demanded range
	Range byte
}

// Write implements the Packet interface.
func (p *ClientViewRange) Write(w io.Writer) {
	putbyte(w, 0xC8)    // ID
	putbyte(w, p.Range) // View range
}

// MoveAcknowledge acknowledges a ClientWalkRequest packet.
type MoveAcknowledge struct {
	// Sequence number of the move from the client
	Sequence int
	// Notoriety of the player
	Notoriety uo.Notoriety
}

// Write implements the Packet interface.
func (p *MoveAcknowledge) Write(w io.Writer) {
	putbyte(w, 0x22)              // ID
	putbyte(w, byte(p.Sequence))  // Move sequence number
	putbyte(w, byte(p.Notoriety)) // Player's notoriety
}

// EquippedMobile is sent to add or update a mobile with equipment graphics.
type EquippedMobile struct {
	// ID of the mobile
	ID uo.Serial
	// Body of the mobile
	Body uo.Body
	// X position of the mobile
	X int
	// Y position of the mobile
	Y int
	// Z position of the mobile
	Z int
	// Direction the mobile is facing
	Facing uo.Direction
	// Hue of the mobile
	Hue uo.Hue
	// Flags
	Flags uo.MobileFlags
	// Notoriety type
	Notoriety uo.Notoriety
	// List of equipped items
	Equipment []*EquippedMobileItem
}

// EquippedMobileItem is used to send information about the equipment a mobile
// is wearing.
type EquippedMobileItem struct {
	// ID of the item
	ID uo.Serial
	// Graphic of the item
	Graphic uo.Item
	// Layer of the item
	Layer uo.Layer
	// Hue of the item
	Hue uo.Hue
}

// Write implements the Packet interface.
func (p *EquippedMobile) Write(w io.Writer) {
	putbyte(w, 0x78) // Packet ID
	putuint16(w, uint16(19+len(p.Equipment)*9+4))
	putuint32(w, uint32(p.ID))
	putuint16(w, uint16(p.Body))
	putuint16(w, uint16(p.X))
	putuint16(w, uint16(p.Y))
	putbyte(w, byte(int8(p.Z)))
	putbyte(w, byte(p.Facing))
	putuint16(w, uint16(p.Hue))
	putbyte(w, byte(p.Flags))
	putbyte(w, byte(p.Notoriety))
	for _, item := range p.Equipment {
		putuint32(w, uint32(item.ID))
		putuint16(w, uint16(item.Graphic.SetHueFlag()))
		putbyte(w, uint8(item.Layer))
		putuint16(w, uint16(item.Hue))
	}
	putuint32(w, 0x00000000) // End of list marker
}

// Target is used to send and recieve targeting commands to the client
type Target struct {
	// Serial of the targeting cursor
	Serial uo.Serial
	// Type of targeting request
	TargetType uo.TargetType
	// Cursor display type
	CursorType uo.CursorType
}

// Write implements the Packet interface.
func (p *Target) Write(w io.Writer) {
	putbyte(w, 0x6C) // Packet ID
	putbyte(w, byte(p.TargetType))
	putuint32(w, uint32(p.Serial))
	putbyte(w, byte(p.CursorType))
	pad(w, 12)
}

// StatusBarInfo sends basic status info to the client.
type StatusBarInfo struct {
	// Serial of the mobile this status applies to
	Mobile uo.Serial
	// Name of the mobile (this gets truncated to 30 characters)
	Name string
	// Current hit points
	HP int
	// Max hit points
	MaxHP int
	// Can the player change the name of this mobile?
	NameChangeFlag bool
	// If true the mobile is female
	Female bool
	// Strength
	Strength int
	// Dexterity
	Dexterity int
	// Intelligence
	Intelligence int
	// Current stamina
	Stamina int
	// Max stamina
	MaxStamina int
	// Current mana
	Mana int
	// Max mana
	MaxMana int
	// Total amount of gold this mobile is currently holding
	Gold int
	// Armor rating
	ArmorRating int
	// Current weight of all equipment and inventory
	Weight int
	// Total stats cap
	StatsCap int
	// Current number of follower slots used
	Followers int
	// Maximum number of follower slots
	MaxFollowers int
}

// Write implements the Packet interface.
func (p *StatusBarInfo) Write(w io.Writer) {
	putbyte(w, 0x11) // Packet ID
	putuint16(w, 70) // Packet length
	putuint32(w, uint32(p.Mobile))
	putstrn(w, p.Name, 30)
	putuint16(w, uint16(p.HP))
	putuint16(w, uint16(p.MaxHP))
	putbool(w, p.NameChangeFlag)
	putbyte(w, 0x03) // UO:R status bar information
	putbool(w, p.Female)
	putuint16(w, uint16(p.Strength))
	putuint16(w, uint16(p.Dexterity))
	putuint16(w, uint16(p.Intelligence))
	putuint16(w, uint16(p.Stamina))
	putuint16(w, uint16(p.MaxStamina))
	putuint16(w, uint16(p.Mana))
	putuint16(w, uint16(p.MaxMana))
	putuint32(w, uint32(p.Gold))
	putuint16(w, uint16(p.ArmorRating))
	putuint16(w, uint16(p.Weight))
	putuint16(w, uint16(p.StatsCap))
	putbyte(w, byte(p.Followers))
	putbyte(w, byte(p.MaxFollowers))
}
