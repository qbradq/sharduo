package serverpacket

import (
	"io"
	"net"

	. "github.com/qbradq/sharduo/lib/dataconv"
	"github.com/qbradq/sharduo/lib/uo"
)

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
	PutByte(w, 0xa8)                     // ID
	PutUint16(w, uint16(length))         // Length
	PutByte(w, 0xcc)                     // Client flags
	PutUint16(w, uint16(len(p.Entries))) // Server count
	// Server list
	for idx, entry := range p.Entries {
		PutUint16(w, uint16(idx))     // Server index
		PutStringN(w, entry.Name, 32) // Server name
		Pad(w, 2)                     // Padding and timezone offset
		// The IP is backward
		PutByte(w, entry.IP.To4()[3])
		PutByte(w, entry.IP.To4()[2])
		PutByte(w, entry.IP.To4()[1])
		PutByte(w, entry.IP.To4()[0])
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
	Key uo.Serial
}

// Write implements the Packet interface.
func (p *ConnectToGameServer) Write(w io.Writer) {
	PutByte(w, 0x8c) // ID
	// IP Address (right-way around)
	PutByte(w, p.IP.To4()[0])
	PutByte(w, p.IP.To4()[1])
	PutByte(w, p.IP.To4()[2])
	PutByte(w, p.IP.To4()[3])
	PutUint16(w, p.Port) // Port
	PutUint32(w, uint32(p.Key))
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
	PutByte(w, 0xa9)               // ID
	PutUint16(w, uint16(length))   // Length
	PutByte(w, byte(len(p.Names))) // Number of character slots
	for _, name := range p.Names {
		PutStringN(w, name, 30)
		Pad(w, 30)
	}
	// Starting locations
	PutByte(w, byte(len(StartingLocations))) // Count
	for i, loc := range StartingLocations {
		PutByte(w, byte(i)) // Index
		PutStringN(w, loc.City, 31)
		PutStringN(w, loc.Area, 31)
	}
	// Flags
	PutUint32(w, 0x000001e8)
}

// LoginComplete is sent after character login is successful.
type LoginComplete struct{}

// Write implements the Packet interface.
func (p *LoginComplete) Write(w io.Writer) {
	PutByte(w, 0x55) // ID
}

// LoginDenied is sent when character login is denied for any reason.
type LoginDenied struct {
	// The reason for the login denial
	Reason uo.LoginDeniedReason
}

// Write implements the Packet interface.
func (p *LoginDenied) Write(w io.Writer) {
	PutByte(w, 0x82) // ID
	PutByte(w, byte(p.Reason))
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
	PutByte(w, 0x1b) // ID
	PutUint32(w, uint32(p.Player))
	Pad(w, 4)
	PutUint16(w, uint16(p.Body))
	PutUint16(w, uint16(p.X))
	PutUint16(w, uint16(p.Y))
	PutByte(w, 0)
	PutByte(w, byte(p.Z))
	PutByte(w, byte(p.Facing))
	PutByte(w, 0)
	Fill(w, 0xff, 4)
	Pad(w, 4)
	PutUint16(w, uint16(p.Width))
	PutUint16(w, uint16(p.Height))
	Pad(w, 6)
}

// Version is sent to the client to request the client version of the packet.
type Version struct{}

// Write implements the Packet interface.
func (p *Version) Write(w io.Writer) {
	PutByte(w, 0xbd) // ID
	PutUint16(w, 3)  // Packet length
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
	PutByte(w, 0x1c) // ID
	PutUint16(w, uint16(44+len(p.Text)+1))
	PutUint32(w, uint32(p.Speaker))
	PutUint16(w, uint16(p.Body))
	PutByte(w, byte(p.Type))
	PutUint16(w, uint16(p.Hue))
	PutUint16(w, uint16(p.Font))
	PutStringN(w, p.Name, 30)
	PutString(w, p.Text)
}

// Ping is sent to the client in response to a client ping packet.
type Ping struct {
	// Key byte of the client ping request
	Key byte
}

// Write implements the Packet interface.
func (p *Ping) Write(w io.Writer) {
	PutByte(w, 0x73)  // ID
	PutByte(w, p.Key) // Key
}

// ClientViewRange sets the client's view range
type ClientViewRange struct {
	// The demanded range
	Range byte
}

// Write implements the Packet interface.
func (p *ClientViewRange) Write(w io.Writer) {
	PutByte(w, 0xC8)    // ID
	PutByte(w, p.Range) // View range
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
	PutByte(w, 0x22)              // ID
	PutByte(w, byte(p.Sequence))  // Move sequence number
	PutByte(w, byte(p.Notoriety)) // Player's notoriety
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
	// Running flag
	IsRunning bool
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
	Graphic uo.Graphic
	// Layer of the item
	Layer uo.Layer
	// Hue of the item
	Hue uo.Hue
}

// Write implements the Packet interface.
func (p *EquippedMobile) Write(w io.Writer) {
	PutByte(w, 0x78) // Packet ID
	PutUint16(w, uint16(19+len(p.Equipment)*9+4))
	PutUint32(w, uint32(p.ID))
	PutUint16(w, uint16(p.Body))
	PutUint16(w, uint16(p.X))
	PutUint16(w, uint16(p.Y))
	PutByte(w, byte(int8(p.Z)))
	// Facing
	if p.IsRunning {
		PutByte(w, byte(p.Facing.SetRunningFlag()))
	} else {
		PutByte(w, byte(p.Facing.StripRunningFlag()))
	}
	PutUint16(w, uint16(p.Hue))
	PutByte(w, byte(p.Flags))
	PutByte(w, byte(p.Notoriety))
	for _, item := range p.Equipment {
		PutUint32(w, uint32(item.ID))
		PutUint16(w, uint16(item.Graphic.SetHueFlag()))
		PutByte(w, uint8(item.Layer))
		PutUint16(w, uint16(item.Hue))
	}
	PutUint32(w, 0x00000000) // End of list marker
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
	PutByte(w, 0x6C) // Packet ID
	PutByte(w, byte(p.TargetType))
	PutUint32(w, uint32(p.Serial))
	PutByte(w, byte(p.CursorType))
	Pad(w, 12)
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
	PutByte(w, 0x11) // Packet ID
	PutUint16(w, 70) // Packet length
	PutUint32(w, uint32(p.Mobile))
	PutStringN(w, p.Name, 30)
	PutUint16(w, uint16(p.HP))
	PutUint16(w, uint16(p.MaxHP))
	PutBool(w, p.NameChangeFlag)
	PutByte(w, 0x03) // UO:R status bar information
	PutBool(w, p.Female)
	PutUint16(w, uint16(p.Strength))
	PutUint16(w, uint16(p.Dexterity))
	PutUint16(w, uint16(p.Intelligence))
	PutUint16(w, uint16(p.Stamina))
	PutUint16(w, uint16(p.MaxStamina))
	PutUint16(w, uint16(p.Mana))
	PutUint16(w, uint16(p.MaxMana))
	PutUint32(w, uint32(p.Gold))
	PutUint16(w, uint16(p.ArmorRating))
	PutUint16(w, uint16(p.Weight))
	PutUint16(w, uint16(p.StatsCap))
	PutByte(w, byte(p.Followers))
	PutByte(w, byte(p.MaxFollowers))
}

// ObjectInfo sends information about a single item or multi to the client.
type ObjectInfo struct {
	// If true we are sending information about a multi
	IsMulti bool
	// Serial of the item or multi
	Serial uo.Serial
	// Graphic of the item or index of the multi into multi.mul
	Graphic uo.Graphic
	// Add this number to the graphic index if amount > 1
	GraphicIncrement int
	// Amount, must be at least 1, no greater than 60000 - always 1 for multi
	Amount int
	// X location of the item or multi
	X int
	// X location of the item or multi
	Y int
	// Z location of the item or multi
	Z int
	// Facing of the item - always 0 for multi
	Facing uo.Direction
	// Layer of the item or 0 if not equipable or multi
	Layer uo.Layer
	// Hue - 0 if multi
	Hue uo.Hue
}

// Write implements the Packet interface.
func (p *ObjectInfo) Write(w io.Writer) {
	PutByte(w, 0xF3)     // Packet ID
	PutUint16(w, 0x0001) // Always 0x0001 on OSI according to POL
	// Data type
	if p.IsMulti {
		PutByte(w, 0x02)
	} else {
		PutByte(w, 0x00)
	}
	PutUint32(w, uint32(p.Serial))
	PutUint16(w, uint16(p.Graphic))
	PutByte(w, byte(p.GraphicIncrement))
	// Amount POL server documentation says the amount field is repeated,
	// ClassicUO ignores the second as unknown.
	var n int
	if p.Amount < int(uo.MinStackAmount) {
		n = int(uo.MinStackAmount)
	} else if p.Amount > int(uo.MaxStackAmount) {
		n = int(uo.MaxStackAmount)
	} else {
		n = p.Amount
	}
	PutUint16(w, uint16(n))
	PutUint16(w, uint16(n))
	// Location
	PutUint16(w, uint16(p.X))
	PutUint16(w, uint16(p.Y))
	PutByte(w, byte(int8(p.Z)))
	// Facing
	if p.IsMulti {
		PutByte(w, 0)
	} else {
		PutByte(w, byte(p.Facing))
	}
	// Hue
	if p.IsMulti {
		PutUint16(w, 0)
	} else {
		PutUint16(w, uint16(p.Hue))
	}
	// Flags
	PutByte(w, 0)
	// Unknown
	Pad(w, 2)
}

// DeleteObject tells the client to forget about an object
type DeleteObject struct {
	// Serial of the object to remove
	Serial uo.Serial
}

// Write implements the Packet interface.
func (p *DeleteObject) Write(w io.Writer) {
	PutByte(w, 0x1D) // Packet ID
	PutUint32(w, uint32(p.Serial))
}

// OpenPaperDoll tells the client to open the paper doll window for a mobile
type OpenPaperDoll struct {
	// Serial of the mobile to display the paper doll of
	Serial uo.Serial
	// Text displayed in the name and title area. Note this gets truncated to
	// 60 characters when sent to the client.
	Text string
	// If true the character is currently in war mode
	WarMode bool
	// If true the player may alter the paper doll
	Alterable bool
}

// Write implements the Packet interface.
func (p *OpenPaperDoll) Write(w io.Writer) {
	var flags byte
	if p.WarMode {
		flags |= 0x01
	}
	if p.Alterable {
		flags |= 0x02
	}
	PutByte(w, 0x88) // Open paper doll
	PutUint32(w, uint32(p.Serial))
	PutStringN(w, p.Text, 60)
	PutByte(w, flags)
}

// MoveSpeed sets the movement speed of the player on the client. This is a
// psuedo-packet for General Information packet 0xBF-0x0026. Note that this does
// NOT set the walk/run/mount state of the client. This is for God mode stuff I
// guess.
type MoveSpeed struct {
	MoveSpeed uo.MoveSpeed
}

// Write implements the Packet interface.
func (p *MoveSpeed) Write(w io.Writer) {
	PutByte(w, 0xBF)     // General Information packet
	PutUint16(w, 6)      // Packet length
	PutUint16(w, 0x0026) // MoveSpeed sub-command
	PutByte(w, byte(p.MoveSpeed))
}
