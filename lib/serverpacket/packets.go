package serverpacket

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"image/color"
	"io"
	"net"
	"strings"
	"time"
	"unicode/utf8"

	dc "github.com/qbradq/sharduo/lib/dataconv"
	"github.com/qbradq/sharduo/lib/uo"
)

// Packet is the interface all server packets implement.
type Packet interface {
	// Write writes the packet data to w.
	Write(w io.Writer)
}

// Utility function to properly write a hue value.
func putHue(w io.Writer, hue uo.Hue) {
	if hue == uo.HueDefault {
		dc.PutUint16(w, 0)
	} else {
		dc.PutUint16(w, uint16(hue+1))
	}
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
	dc.PutByte(w, 0xa8)                     // ID
	dc.PutUint16(w, uint16(length))         // Length
	dc.PutByte(w, 0xcc)                     // Client flags
	dc.PutUint16(w, uint16(len(p.Entries))) // Server count
	// Server list
	for idx, entry := range p.Entries {
		dc.PutUint16(w, uint16(idx))     // Server index
		dc.PutStringN(w, entry.Name, 32) // Server name
		dc.Pad(w, 2)                     // Padding and timezone offset
		// The IP is backward
		dc.PutByte(w, entry.IP.To4()[3])
		dc.PutByte(w, entry.IP.To4()[2])
		dc.PutByte(w, entry.IP.To4()[1])
		dc.PutByte(w, entry.IP.To4()[0])
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
	dc.PutByte(w, 0x8c) // ID
	// IP Address (right-way around)
	dc.PutByte(w, p.IP.To4()[0])
	dc.PutByte(w, p.IP.To4()[1])
	dc.PutByte(w, p.IP.To4()[2])
	dc.PutByte(w, p.IP.To4()[3])
	dc.PutUint16(w, p.Port) // Port
	dc.PutUint32(w, uint32(p.Key))
}

// CharacterList is sent on game server login and lists all characters on the
// account as well as the new character starting locations.
type CharacterList struct {
	// Names of all of the characters, empty string for open slots.
	Names []string
}

// Write implements the Packet interface.
func (p *CharacterList) Write(w io.Writer) {
	length := 4 + len(p.Names)*60 + 1 + 4
	dc.PutByte(w, 0xa9)               // ID
	dc.PutUint16(w, uint16(length))   // Length
	dc.PutByte(w, byte(len(p.Names))) // Number of character slots
	for _, name := range p.Names {
		dc.PutStringN(w, name, 30)
		dc.Pad(w, 30)
	}
	dc.PutByte(w, 0)
	// Flags
	dc.PutUint32(w, 0x0000003c)
}

// LoginComplete is sent after character login is successful.
type LoginComplete struct{}

// Write implements the Packet interface.
func (p *LoginComplete) Write(w io.Writer) {
	dc.PutByte(w, 0x55) // ID
}

// LoginDenied is sent when character login is denied for any reason.
type LoginDenied struct {
	// The reason for the login denial
	Reason uo.LoginDeniedReason
}

// Write implements the Packet interface.
func (p *LoginDenied) Write(w io.Writer) {
	dc.PutByte(w, 0x82) // ID
	dc.PutByte(w, byte(p.Reason))
}

// EnterWorld is sent just after character login to bring them into the world.
type EnterWorld struct {
	// Player serial
	Player uo.Serial
	// Body graphic
	Body uo.Body
	// Position
	Location uo.Location
	// Direction the player is facing and if running.
	Facing uo.Direction
	// Server dimensions
	Width, Height int
}

// Write implements the Packet interface.
func (p *EnterWorld) Write(w io.Writer) {
	dc.PutByte(w, 0x1b) // ID
	dc.PutUint32(w, uint32(p.Player))
	dc.Pad(w, 4)
	dc.PutUint16(w, uint16(p.Body))
	dc.PutUint16(w, uint16(p.Location.X))
	dc.PutUint16(w, uint16(p.Location.Y))
	dc.PutByte(w, 0)
	dc.PutByte(w, byte(p.Location.Z))
	dc.PutByte(w, byte(p.Facing))
	dc.PutByte(w, 0)
	dc.Fill(w, 0xff, 4)
	dc.Pad(w, 4)
	dc.PutUint16(w, uint16(p.Width))
	dc.PutUint16(w, uint16(p.Height))
	dc.Pad(w, 6)
}

// Version is sent to the client to request the client version of the packet.
type Version struct{}

// Write implements the Packet interface.
func (p *Version) Write(w io.Writer) {
	dc.PutByte(w, 0xbd) // ID
	dc.PutUint16(w, 3)  // Packet length
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
	dc.PutByte(w, 0x1c) // ID
	dc.PutUint16(w, uint16(44+len(p.Text)+1))
	dc.PutUint32(w, uint32(p.Speaker))
	dc.PutUint16(w, uint16(p.Body))
	dc.PutByte(w, byte(p.Type))
	putHue(w, p.Hue)
	dc.PutUint16(w, uint16(p.Font))
	dc.PutStringN(w, p.Name, 30)
	dc.PutString(w, p.Text)
}

// Ping is sent to the client in response to a client ping packet.
type Ping struct {
	// Key byte of the client ping request
	Key byte
}

// Write implements the Packet interface.
func (p *Ping) Write(w io.Writer) {
	dc.PutByte(w, 0x73)  // ID
	dc.PutByte(w, p.Key) // Key
}

// ClientViewRange sets the client's view range
type ClientViewRange struct {
	// The demanded range
	Range byte
}

// Write implements the Packet interface.
func (p *ClientViewRange) Write(w io.Writer) {
	dc.PutByte(w, 0xC8)    // ID
	dc.PutByte(w, p.Range) // View range
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
	dc.PutByte(w, 0x22)              // ID
	dc.PutByte(w, byte(p.Sequence))  // Move sequence number
	dc.PutByte(w, byte(p.Notoriety)) // Player's notoriety
}

// EquippedMobile is sent to add or update a mobile with equipment graphics.
type EquippedMobile struct {
	// ID of the mobile
	ID uo.Serial
	// Body of the mobile
	Body uo.Body
	// Position of the mobile
	Location uo.Location
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
	dc.PutByte(w, 0x78) // Packet ID
	dc.PutUint16(w, uint16(19+len(p.Equipment)*9+4))
	dc.PutUint32(w, uint32(p.ID))
	dc.PutUint16(w, uint16(p.Body))
	dc.PutUint16(w, uint16(p.Location.X))
	dc.PutUint16(w, uint16(p.Location.Y))
	dc.PutByte(w, byte(p.Location.Z))
	// Facing
	if p.IsRunning {
		dc.PutByte(w, byte(p.Facing.SetRunningFlag()))
	} else {
		dc.PutByte(w, byte(p.Facing.StripRunningFlag()))
	}
	putHue(w, p.Hue)
	dc.PutByte(w, byte(p.Flags))
	dc.PutByte(w, byte(p.Notoriety))
	for _, item := range p.Equipment {
		dc.PutUint32(w, uint32(item.ID))
		dc.PutUint16(w, uint16(item.Graphic.SetHueFlag()))
		dc.PutByte(w, uint8(item.Layer))
		putHue(w, item.Hue)
	}
	dc.PutUint32(w, 0x00000000) // End of list marker
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
	dc.PutByte(w, 0x6C) // Packet ID
	dc.PutByte(w, byte(p.TargetType))
	dc.PutUint32(w, uint32(p.Serial))
	dc.PutByte(w, byte(p.CursorType))
	dc.Pad(w, 12)
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
	dc.PutByte(w, 0x11) // Packet ID
	dc.PutUint16(w, 70) // Packet length
	dc.PutUint32(w, uint32(p.Mobile))
	dc.PutStringN(w, p.Name, 30)
	dc.PutUint16(w, uint16(p.HP))
	dc.PutUint16(w, uint16(p.MaxHP))
	dc.PutBool(w, p.NameChangeFlag)
	dc.PutByte(w, 0x03) // UO:R status bar information
	dc.PutBool(w, p.Female)
	dc.PutUint16(w, uint16(p.Strength))
	dc.PutUint16(w, uint16(p.Dexterity))
	dc.PutUint16(w, uint16(p.Intelligence))
	dc.PutUint16(w, uint16(p.Stamina))
	dc.PutUint16(w, uint16(p.MaxStamina))
	dc.PutUint16(w, uint16(p.Mana))
	dc.PutUint16(w, uint16(p.MaxMana))
	dc.PutUint32(w, uint32(p.Gold))
	dc.PutUint16(w, uint16(p.ArmorRating))
	dc.PutUint16(w, uint16(p.Weight))
	dc.PutUint16(w, uint16(p.StatsCap))
	dc.PutByte(w, byte(p.Followers))
	dc.PutByte(w, byte(p.MaxFollowers))
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
	// Location of the item or multi
	Location uo.Location
	// Facing of the item - always 0 for multi
	Facing uo.Direction
	// Layer of the item or 0 if not equipable or multi
	Layer uo.Layer
	// Hue - 0 if multi
	Hue uo.Hue
	// If true the object will be moveable even if normally not. Note that even
	// when this is false the client may still treat the object as movable
	// depending on the contents of the tile definition for the graphic.
	Movable bool
}

// Write implements the Packet interface.
func (p *ObjectInfo) Write(w io.Writer) {
	dc.PutByte(w, 0xF3)     // Packet ID
	dc.PutUint16(w, 0x0001) // Always 0x0001 on OSI according to POL
	// Data type
	if p.IsMulti {
		dc.PutByte(w, 0x02)
	} else {
		dc.PutByte(w, 0x00)
	}
	dc.PutUint32(w, uint32(p.Serial))
	dc.PutUint16(w, uint16(p.Graphic))
	dc.PutByte(w, byte(p.GraphicIncrement))
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
	dc.PutUint16(w, uint16(n))
	dc.PutUint16(w, uint16(n))
	// Location
	dc.PutUint16(w, uint16(p.Location.X&0x7FFF))
	dc.PutUint16(w, uint16(p.Location.Y&0x3FFF))
	dc.PutByte(w, byte(p.Location.Z))
	// Facing
	dc.PutByte(w, 0)
	// Hue
	if p.IsMulti {
		dc.PutUint16(w, 0)
	} else {
		putHue(w, p.Hue)
	}
	// Flags
	flags := byte(0)
	if p.Movable {
		flags |= 0x20
	}
	dc.PutByte(w, flags) // Movable if normally not
	// Unknown
	dc.Pad(w, 2)
}

// DeleteObject tells the client to forget about an object
type DeleteObject struct {
	// Serial of the object to remove
	Serial uo.Serial
}

// Write implements the Packet interface.
func (p *DeleteObject) Write(w io.Writer) {
	dc.PutByte(w, 0x1D) // Packet ID
	dc.PutUint32(w, uint32(p.Serial))
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
	dc.PutByte(w, 0x88) // Open paper doll
	dc.PutUint32(w, uint32(p.Serial))
	dc.PutStringN(w, p.Text, 60)
	dc.PutByte(w, flags)
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
	dc.PutByte(w, 0xBF)     // General Information packet
	dc.PutUint16(w, 6)      // Packet length
	dc.PutUint16(w, 0x0026) // MoveSpeed sub-command
	dc.PutByte(w, byte(p.MoveSpeed))
}

// DrawPlayer updates the player's location and appearance
type DrawPlayer struct {
	// Serial of the player
	ID uo.Serial
	// Body graphic
	Body uo.Body
	// Skin hue
	Hue uo.Hue
	// Flags field
	Flags uo.MobileFlags
	// Location of the mobile
	Location uo.Location
	// Direction the mobile is facing
	Facing uo.Direction
}

// Write implements the Packet interface.
func (p *DrawPlayer) Write(w io.Writer) {
	dc.PutByte(w, 0x20) // Packet ID
	dc.PutUint32(w, uint32(p.ID))
	dc.PutUint16(w, uint16(p.Body))
	dc.Pad(w, 1)
	putHue(w, p.Hue)
	dc.PutByte(w, byte(p.Flags))
	dc.PutUint16(w, uint16(p.Location.X))
	dc.PutUint16(w, uint16(p.Location.Y))
	dc.Pad(w, 2)
	dc.PutByte(w, byte(p.Facing))
	dc.PutByte(w, byte(int8(p.Location.Z)))
}

// DropApproved is sent to the client to acknowledge a drop or equip request.
type DropApproved struct{}

// Write implements the Packet interface.
func (p *DropApproved) Write(w io.Writer) {
	dc.PutByte(w, 0x29) // Packet ID
}

// WornItem is sent to clients to inform them of an item added to a mobile's
// equipment.
type WornItem struct {
	// The item being worn
	Item uo.Serial
	// Graphic of the item
	Graphic uo.Graphic
	// Layer of the item
	Layer uo.Layer
	// Mobile wearing the item
	Wearer uo.Serial
	// Hue of the item
	Hue uo.Hue
}

// Write implements the Packet interface.
func (p *WornItem) Write(w io.Writer) {
	dc.PutByte(w, 0x2E) // Packet ID
	dc.PutUint32(w, uint32(p.Item))
	dc.PutUint16(w, uint16(p.Graphic))
	dc.Pad(w, 1)
	dc.PutByte(w, byte(p.Layer))
	dc.PutUint32(w, uint32(p.Wearer))
	putHue(w, p.Hue)
}

// MoveItemReject rejects a pick-up, drop, or equip request
type MoveItemReject struct {
	Reason uo.MoveItemRejectReason
}

// Write implements the Packet interface.
func (p *MoveItemReject) Write(w io.Writer) {
	dc.PutByte(w, 0x27) // Packet ID
	dc.PutByte(w, byte(p.Reason))
}

// MoveMobile moves an existing mobile on the client side
type MoveMobile struct {
	// Serial of the mobile to update
	ID uo.Serial
	// Body of the mobile
	Body uo.Body
	// Location of the mobile
	Location uo.Location
	// Facing
	Facing uo.Direction
	// Running flag
	Running bool
	// Hue
	Hue uo.Hue
	// Mobile flags
	Flags uo.MobileFlags
	// Notoriety
	Notoriety uo.Notoriety
}

// Write implements the Packet interface.
func (p *MoveMobile) Write(w io.Writer) {
	dc.PutByte(w, 0x77) // Packet ID
	dc.PutUint32(w, uint32(p.ID))
	dc.PutUint16(w, uint16(p.Body))
	dc.PutUint16(w, uint16(p.Location.X))
	dc.PutUint16(w, uint16(p.Location.Y))
	dc.PutByte(w, byte(int8(p.Location.Z)))
	// Facing
	if p.Running {
		dc.PutByte(w, byte(p.Facing.SetRunningFlag()))
	} else {
		dc.PutByte(w, byte(p.Facing.StripRunningFlag()))
	}
	putHue(w, p.Hue)
	dc.PutByte(w, byte(p.Flags))
	dc.PutByte(w, byte(p.Notoriety))
}

// DragItem makes the client play an animation of the item being dragged from
// the source to the destination.
type DragItem struct {
	// Graphic is the graphic of the item being moved
	Graphic uo.Graphic
	// Graphic offset, used for stacked graphics. This is truncated to 8-bits
	// and must be positive.
	GraphicOffset int
	// Hue of the item being moved
	Hue uo.Hue
	// Amount in the stack
	Amount int
	// Source mobile serial, uo.SerialSystem for the map
	Source uo.Serial
	// Source position
	SourceLocation uo.Location
	// Destination mobile serial, uo.SerialSystem for the map
	Destination uo.Serial
	// Destination position
	DestinationLocation uo.Location
}

// Write implements the Packet interface.
func (p *DragItem) Write(w io.Writer) {
	dc.PutByte(w, 0x23) // Packet ID
	dc.PutUint16(w, uint16(p.Graphic))
	dc.PutByte(w, byte(p.GraphicOffset))
	putHue(w, p.Hue)
	dc.PutUint16(w, uint16(p.Amount))
	dc.PutUint32(w, uint32(p.Source))
	dc.PutUint16(w, uint16(p.SourceLocation.X))
	dc.PutUint16(w, uint16(p.SourceLocation.Y))
	dc.PutByte(w, byte(int8(p.SourceLocation.Z)))
	dc.PutUint32(w, uint32(p.Destination))
	dc.PutUint16(w, uint16(p.DestinationLocation.X))
	dc.PutUint16(w, uint16(p.DestinationLocation.Y))
	dc.PutByte(w, byte(int8(p.DestinationLocation.Z)))
}

// OpenContainerGump opens a container gump on the client.
type OpenContainerGump struct {
	// The ID of the Gump
	GumpSerial uo.Serial
	// The gump graphic
	Gump uo.GUMP
}

// Write implements the Packet interface.
func (p *OpenContainerGump) Write(w io.Writer) {
	dc.PutByte(w, 0x24) // Packet ID
	dc.PutUint32(w, uint32(p.GumpSerial))
	dc.PutUint16(w, uint16(p.Gump))
	dc.PutUint16(w, uint16(0x007D)) // No idea what this does but it's required > 7.0.9.x
}

// AddItemToContainer adds an item to an already-open container gump.
type AddItemToContainer struct {
	// The ID of the item being added to the container
	Item uo.Serial
	// Graphic of the item being added to the container
	Graphic uo.Graphic
	// Graphic offset of the item
	GraphicOffset int
	// Stack amount, truncated to 0-0xFFFF inclusive
	Amount int
	// Location of the item in the container. X=Y=0xFFFF means random location.
	Location uo.Location
	// The ID of the container and container gump to add this item to
	Container uo.Serial
	// Hue of the item
	Hue uo.Hue
}

// Write implements the Packet interface.
func (p *AddItemToContainer) Write(w io.Writer) {
	dc.PutByte(w, 0x25) // Packet ID
	dc.PutUint32(w, uint32(p.Item))
	dc.PutUint16(w, uint16(p.Graphic))
	dc.PutByte(w, byte(p.GraphicOffset))
	dc.PutUint16(w, uint16(p.Amount))
	dc.PutUint16(w, uint16(p.Location.X))
	dc.PutUint16(w, uint16(p.Location.Y))
	dc.Pad(w, 1) // Grid index
	dc.PutUint32(w, uint32(p.Container))
	putHue(w, p.Hue)
}

// ContentsItem represents one item in a Contents packet.
type ContentsItem struct {
	// Serial of the item
	Serial uo.Serial
	// Item graphic
	Graphic uo.Graphic
	// Item graphic offset, this gets truncated between 0-255 inclusive
	GraphicOffset int
	// Stack amount
	Amount int
	// Location of the item in the container
	Location uo.Location
	// Serial of the container to add the item to
	Container uo.Serial
	// Hue of the item
	Hue uo.Hue
	// Price of the item if it is being sold
	Price uint32
	// Shop description if it is being sold
	Description string
}

// Contents sends the contents of a container to the client.
type Contents struct {
	// The items in the container
	Items []ContentsItem
	// If true items are listed in reverse order
	ReverseOrder bool
}

// Write implements the Packet interface.
func (p *Contents) Write(w io.Writer) {
	dc.PutByte(w, 0x3C)                        // Packet ID
	dc.PutUint16(w, uint16(5+len(p.Items)*20)) // Packet length
	dc.PutUint16(w, uint16(len(p.Items)))
	if p.ReverseOrder {
		for i := len(p.Items) - 1; i >= 0; i-- {
			item := p.Items[i]
			dc.PutUint32(w, uint32(item.Serial))
			dc.PutUint16(w, uint16(item.Graphic))
			dc.PutByte(w, byte(item.GraphicOffset))
			dc.PutUint16(w, uint16(item.Amount))
			dc.PutUint16(w, uint16(item.Location.X))
			dc.PutUint16(w, uint16(item.Location.Y))
			dc.Pad(w, 1) // Grid index
			dc.PutUint32(w, uint32(item.Container))
			putHue(w, item.Hue)
		}
	} else {
		for _, item := range p.Items {
			dc.PutUint32(w, uint32(item.Serial))
			dc.PutUint16(w, uint16(item.Graphic))
			dc.PutByte(w, byte(item.GraphicOffset))
			dc.PutUint16(w, uint16(item.Amount))
			dc.PutUint16(w, uint16(item.Location.X))
			dc.PutUint16(w, uint16(item.Location.Y))
			dc.Pad(w, 1) // Grid index
			dc.PutUint32(w, uint32(item.Container))
			putHue(w, item.Hue)
		}
	}
}

// CloseGump sends a force gump close BF subcommand to forcefully close a gump
// on the client.
type CloseGump struct {
	// Serial of the gump to close
	Gump uo.Serial
	// Button response for the gump response packet, use 0 for close gump
	Button int
}

// Write implements the Packet interface.
func (p *CloseGump) Write(w io.Writer) {
	dc.PutByte(w, 0xBF)             // General information packet ID
	dc.PutUint16(w, uint16(13))     // Length
	dc.PutUint16(w, uint16(0x0004)) // Close gump subcommand
	dc.PutUint32(w, uint32(p.Gump))
	dc.PutUint32(w, uint32(p.Button))
}

// MoveReject sends a movement rejection packet to the client.
type MoveReject struct {
	// Sequence number of the movement request rejected
	Sequence byte
	// Location of the mobile after the rejection
	Location uo.Location
	// Facing of the mobile
	Facing uo.Direction
}

// Write implements the Packet interface.
func (p *MoveReject) Write(w io.Writer) {
	dc.PutByte(w, 0x21) // Packet ID
	dc.PutByte(w, p.Sequence)
	dc.PutUint16(w, uint16(p.Location.X))
	dc.PutUint16(w, uint16(p.Location.Y))
	dc.PutByte(w, byte(p.Facing))
	dc.PutByte(w, byte(int8(p.Location.Z)))
}

// SingleSkillUpdate sends an update for a single skill.
type SingleSkillUpdate struct {
	// Which skill changed
	Skill uo.Skill
	// New raw value of the skill (0-1000)
	Value int
	// Lock state
	Lock uo.SkillLock
}

// Write implements the Packet interface.
func (p *SingleSkillUpdate) Write(w io.Writer) {
	dc.PutByte(w, 0x3A)                       // Packet ID
	dc.PutUint16(w, 13)                       // Packet length
	dc.PutByte(w, byte(uo.SkillUpdateSingle)) // Update type
	dc.PutUint16(w, uint16(p.Skill))          // Skill index
	dc.PutUint16(w, uint16(p.Value))          // Display value
	dc.PutUint16(w, uint16(p.Value))          // Base value
	dc.PutByte(w, byte(p.Lock))               // Lock code
	dc.PutUint16(w, 1000)                     // Skill cap
}

// FullSkillUpdate sends an update for all skills.
type FullSkillUpdate struct {
	// Slice of all skill values
	SkillValues []int16
}

// Write implements the Packet interface.
func (p *FullSkillUpdate) Write(w io.Writer) {
	dc.PutByte(w, 0x3A)                             // Packet ID
	dc.PutUint16(w, uint16(4+len(p.SkillValues)*9)) // Packet length
	dc.PutByte(w, byte(uo.SkillUpdateAll))          // Update type
	for id, value := range p.SkillValues {
		dc.PutUint16(w, uint16(id+1))       // Skill ID - Not sure why this is 1-based in this one packet, but oh well
		dc.PutUint16(w, uint16(value))      // Displayed value
		dc.PutUint16(w, uint16(value))      // Base value
		dc.PutByte(w, byte(uo.SkillLockUp)) // Skill lock
		dc.PutUint16(w, 1000)               // Skill cap
	}
}

// ClilocMessage sends a localized message to the client.
type ClilocMessage struct {
	// Serial of the speaker
	Speaker uo.Serial
	// Body of the speaker
	Body uo.Body
	// Hue of the text
	Hue uo.Hue
	// Font of the text
	Font uo.Font
	// Cliloc message number
	Cliloc uo.Cliloc
	// Name of the speaker
	Name string
	// List of arguments for the message
	Arguments []string
}

// Write implements the Packet interface.
func (p *ClilocMessage) Write(w io.Writer) {
	args := strings.Join(p.Arguments, "\t")
	// Calculate packet length
	l := 50                            // Fixed portions of the packet, including the terminating null
	l += len([]rune(args)) * 2         // Two bytes per rune
	dc.PutByte(w, 0xC1)                // Packet ID
	dc.PutUint16(w, uint16(l))         // Packet length
	dc.PutUint32(w, uint32(p.Speaker)) // Speaker's ID
	dc.PutUint16(w, uint16(p.Body))    // Speaker's body graphic
	// Message type handling
	if p.Speaker == uo.SerialSystem {
		dc.PutByte(w, byte(0x06))
	} else {
		dc.PutByte(w, byte(0x07))
	}
	putHue(w, p.Hue)                  // Message hue
	dc.PutUint16(w, uint16(p.Font))   // Message font
	dc.PutUint32(w, uint32(p.Cliloc)) // Message index number
	dc.PutStringN(w, p.Name, 30)
	dc.PutUTF16LEString(w, args)
}

// Sound tells the client to play a sound from a specific location.
type Sound struct {
	// Which sound to play
	Sound uo.Sound
	// Where the sound is coming from
	Location uo.Location
}

// Write implements the Packet interface.
func (p *Sound) Write(w io.Writer) {
	dc.PutByte(w, 0x54)                     // Packet ID
	dc.PutByte(w, 0x01)                     // Sound type, 0=quiet, 1=normal
	dc.PutUint16(w, uint16(p.Sound))        // Sound ID
	dc.PutUint16(w, 0x00)                   // Volume, ServUO always sets this to 0
	dc.PutUint16(w, uint16(p.Location.X))   // X position
	dc.PutUint16(w, uint16(p.Location.Y))   // Y position
	dc.PutByte(w, 0x00)                     // Facing byte? Always 0
	dc.PutByte(w, byte(int8(p.Location.Z))) // Z position
}

// Music tells the client to start playing the given song.
type Music struct {
	Song uo.Music // Which song to play
}

// Write implements the Packet interface.
func (p *Music) Write(w io.Writer) {
	dc.PutByte(w, 0x6D)             // Packet ID
	dc.PutUint16(w, uint16(p.Song)) // Song ID
}

// Animation tells the client to animate a mobile.
type Animation struct {
	// Serial of the mobile to animate
	Serial uo.Serial
	// Animation type
	AnimationType uo.AnimationType
	// Animation action
	AnimationAction uo.AnimationAction
}

// Write implements the Packet interface.
func (p *Animation) Write(w io.Writer) {
	dc.PutByte(w, 0xE2)                        // Packet ID
	dc.PutUint32(w, uint32(p.Serial))          // Mobile to animate
	dc.PutUint16(w, uint16(p.AnimationType))   // Animation to play
	dc.PutUint16(w, uint16(p.AnimationAction)) // Sub-animation to play
	dc.PutByte(w, 0)                           // Mode
}

// Time tells the client the server time.
type Time struct {
	// Server time
	Time time.Time
}

// Write implements the Packet interface.
func (p *Time) Write(w io.Writer) {
	dc.PutByte(w, 0x5B) // Packet ID
	dc.PutByte(w, byte(p.Time.Hour()))
	dc.PutByte(w, byte(p.Time.Minute()))
	dc.PutByte(w, byte(p.Time.Second()))
}

// GlobalLightLevel sets the overall light level for the client.
type GlobalLightLevel struct {
	// Light level to set
	LightLevel uo.LightLevel
}

// Write implements the Packet interface.
func (p *GlobalLightLevel) Write(w io.Writer) {
	dc.PutByte(w, 0x4F) // Packet ID
	dc.PutByte(w, byte(p.LightLevel))
}

// PersonalLightLevel sets the personal light level for the mobile.
type PersonalLightLevel struct {
	// Serial of the mobile
	Serial uo.Serial
	// Light level to set
	LightLevel uo.LightLevel
}

// Write implements the Packet interface.
func (p *PersonalLightLevel) Write(w io.Writer) {
	dc.PutByte(w, 0x4E) // Packet ID
	dc.PutUint32(w, uint32(p.Serial))
	dc.PutByte(w, byte(p.LightLevel))
}

// ctxMenuEntry is an entry for a context menu.
type ctxMenuEntry struct {
	// Unique ID of the entry
	ID uint16
	// Cliloc ID - 3,000,000
	Cliloc uint16
}

// ContextMenu sends a context menu to the client.
type ContextMenu struct {
	// Serial of the object this context menu is to appear over
	Serial uo.Serial
	// Entries of the menu
	Entries []ctxMenuEntry
}

// Add adds an entry to the context menu. The cliloc parameter must be in the
// range 3,000,000 - 3,060,000 inclusive.
func (p *ContextMenu) Add(id uint16, cliloc uo.Cliloc) {
	cl := uint16(uint32(cliloc) - 3_000_000)
	p.Entries = append(p.Entries, ctxMenuEntry{id, cl})
}

// Write implements the Packet interface.
func (p *ContextMenu) Write(w io.Writer) {
	dc.PutByte(w, 0xBF)                          // General packet ID
	dc.PutUint16(w, uint16(12+len(p.Entries)*6)) // Packet length
	dc.PutUint16(w, 0x0014)                      // Subcommand ID
	dc.Pad(w, 1)
	dc.PutByte(w, 0x01) // Subsubcommand
	dc.PutUint32(w, uint32(p.Serial))
	dc.PutByte(w, byte(len(p.Entries)))
	for _, e := range p.Entries {
		dc.PutUint16(w, uint16(e.ID))
		dc.PutUint16(w, uint16(e.Cliloc))
		dc.PutUint16(w, 0x0000) // Enabled
	}
}

// GUMP sends a non-compressed generic GUMP to the client.
type GUMP struct {
	// Sender code of the GUMP layout
	Sender uo.Serial
	// TypeCode of the GUMP returned in reply packets
	TypeCode uo.Serial
	// Layout string for the GUMP
	Layout string
	// Location of the GUMP on screen
	Location uo.Location
	// Text lines
	Lines []string
	// If true the GUMP data will be sent uncompressed
	DoNotCompress bool
}

// Write implements the Packet interface.
func (p *GUMP) Write(w io.Writer) {
	// Use the old GUMP packet
	if p.DoNotCompress {
		// Calculate length
		l := 23                        // All fixed-width fields
		l += len(p.Lines) * 2          // Length field for all lines
		l += len(p.Layout)             // Layout section
		for _, line := range p.Lines { // Length of the lines of text
			l += utf8.RuneCountInString(line) * 2
		}
		dc.PutByte(w, 0xB0)                       // General packet ID
		dc.PutUint16(w, uint16(l))                // Packet length
		dc.PutUint32(w, uint32(p.Sender))         // Player mobile serial
		dc.PutUint32(w, uint32(p.TypeCode))       // GUMP serial
		dc.PutUint32(w, uint32(p.Location.X))     // Screen location X
		dc.PutUint32(w, uint32(p.Location.Y))     // Screen location Y
		dc.PutUint16(w, uint16(len(p.Layout)))    // Layout data length
		dc.PutStringN(w, p.Layout, len(p.Layout)) // Layout data
		dc.PutUint16(w, uint16(len(p.Lines)))     // Number of text lines
		for _, line := range p.Lines {
			lrc := utf8.RuneCountInString(line)
			dc.PutUint16(w, uint16(lrc))     // Length of the line in runes
			dc.PutUTF16StringN(w, line, lrc) // Line data
		}
	} else {
		// Compress layout data
		fb := bytes.NewBuffer(nil)
		fz := zlib.NewWriter(fb)
		fz.Write([]byte(p.Layout))
		fz.Close()
		fd := fb.Bytes()
		// Build the line data
		lbraw := bytes.NewBuffer(nil)
		for _, line := range p.Lines {
			lrc := utf8.RuneCountInString(line)
			dc.PutUint16(lbraw, uint16(lrc))
			dc.PutUTF16StringN(lbraw, line, lrc)
		}
		ldraw := lbraw.Bytes()
		lb := bytes.NewBuffer(nil)
		lz := zlib.NewWriter(lb)
		lz.Write(ldraw)
		lz.Close()
		ld := lb.Bytes()
		// Calculate packet length
		l := 39 + len(fd) + len(ld)
		// Write the packet
		dc.PutByte(w, 0xDD)                    // Packet ID
		dc.PutUint16(w, uint16(l))             // Packet length
		dc.PutUint32(w, uint32(p.Sender))      // Player mobile's serial
		dc.PutUint32(w, uint32(p.TypeCode))    // GUMP serial
		dc.PutUint32(w, uint32(p.Location.X))  // Screen location X
		dc.PutUint32(w, uint32(p.Location.Y))  // Screen location Y
		dc.PutUint32(w, uint32(len(fd)+4))     // Compressed layout length
		dc.PutUint32(w, uint32(len(p.Layout))) // Decompressed layout length
		w.Write(fd)                            // Compressed layout
		dc.PutUint32(w, uint32(len(p.Lines)))  // Number of lines
		dc.PutUint32(w, uint32(len(ld)+4))     // Compressed lines length
		dc.PutUint32(w, uint32(len(ldraw)))    // Decompressed lines length
		w.Write(ld)                            // Compressed lines
	}
}

// GraphicalEffect sends a graphical effect packet to the client
type GraphicalEffect struct {
	// Behavior of the effect
	GFXType uo.GFXType
	// Serial of the source object
	Source uo.Serial
	// Serial of the target object
	Target uo.Serial
	// First frame of the effect
	Graphic uo.Graphic
	// Source location
	SourceLocation uo.Location
	// Target location
	TargetLocation uo.Location
	// Speed of the animation in FPS?
	Speed uint8
	// Duration of the animation 1=Slowest, 0=Even slower for some reason
	Duration uint8
	// If true the projectile will not attempt to change facing during flight
	Fixed bool
	// If true the projectile will explode on impact
	Explodes bool
	// Hue of the effect
	Hue uo.Hue
	// Render mode of the effect
	GFXBlendMode uo.GFXBlendMode
}

// Write implements the Packet interface.
func (p *GraphicalEffect) Write(w io.Writer) {
	dc.PutByte(w, 0xC0) // Packet ID
	dc.PutByte(w, byte(p.GFXType))
	dc.PutUint32(w, uint32(p.Source))
	dc.PutUint32(w, uint32(p.Target))
	dc.PutUint16(w, uint16(p.Graphic))
	dc.PutUint16(w, uint16(p.SourceLocation.X))
	dc.PutUint16(w, uint16(p.SourceLocation.Y))
	dc.PutByte(w, byte(p.SourceLocation.Z))
	dc.PutUint16(w, uint16(p.TargetLocation.X))
	dc.PutUint16(w, uint16(p.TargetLocation.Y))
	dc.PutByte(w, byte(p.TargetLocation.Z))
	dc.PutByte(w, byte(p.Speed))
	dc.PutByte(w, byte(p.Duration))
	dc.Pad(w, 2)
	dc.PutBool(w, p.Fixed)
	dc.PutBool(w, p.Explodes)
	dc.Pad(w, 2)
	putHue(w, p.Hue)
	dc.PutUint32(w, uint32(p.GFXBlendMode))
}

// BuyWindow transfers the buy window details to the client.
type BuyWindow struct {
	// Serial of the container of the buy window.
	Serial uo.Serial
	// The list of items in the container in normal order. The Write method
	// takes care of reversing the order.
	Items []ContentsItem
}

// Write implements the Packet interface.
func (p *BuyWindow) Write(w io.Writer) {
	// Calculate packet length
	l := 8
	for _, i := range p.Items {
		l += 5 + len(i.Description)
	}
	dc.PutByte(w, 0x74)               // Packet ID
	dc.PutUint16(w, uint16(l))        // Packet length
	dc.PutUint32(w, uint32(p.Serial)) // Container serial
	dc.PutByte(w, byte(len(p.Items))) // Number of items
	for _, i := range p.Items {
		dc.PutUint32(w, i.Price)                            // Item price
		dc.PutByte(w, byte(len(i.Description)))             // Description length
		dc.PutStringN(w, i.Description, len(i.Description)) // Description
	}
}

// VendorBuySequence implements the required sequence of packets to open an NPC
// buy window.
type VendorBuySequence struct {
	// Serial of the vendor
	Vendor uo.Serial
	// Serial of the sell container
	ForSale uo.Serial
	// Serial of the bought container
	Bought uo.Serial
	// List of items in the sell container
	ForSaleItems []ContentsItem
	// List of items in the bought container
	BoughtItems []ContentsItem
}

// Write implements the Packet interface.
func (p *VendorBuySequence) Write(w io.Writer) {
	// Wear ForSale container packet
	wp := WornItem{
		Item:    p.ForSale,
		Graphic: 0x0E75,
		Layer:   uo.LayerNPCBuyRestockContainer,
		Wearer:  p.Vendor,
		Hue:     uo.HueDefault,
	}
	wp.Write(w)
	// Wear Bought container packet
	wp = WornItem{
		Item:    p.Bought,
		Graphic: 0x0E75,
		Layer:   uo.LayerNPCBuyNoRestockContainer,
		Wearer:  p.Vendor,
		Hue:     uo.HueDefault,
	}
	wp.Write(w)
	// Contents packet for the ForSale container
	cp := Contents{
		Items:        p.ForSaleItems,
		ReverseOrder: true,
	}
	cp.Write(w)
	// BuyWindow packet for the ForSale container
	bp := BuyWindow{
		Serial: p.ForSale,
		Items:  p.ForSaleItems,
	}
	bp.Write(w)
	// Contents packet for the Bought container
	cp = Contents{
		Items:        p.BoughtItems,
		ReverseOrder: true,
	}
	cp.Write(w)
	// BuyWindow packet for the ForSale container
	bp = BuyWindow{
		Serial: p.Bought,
		Items:  p.BoughtItems,
	}
	bp.Write(w)
	// Open container packet
	op := OpenContainerGump{
		GumpSerial: p.Vendor,
		Gump:       0x0030,
	}
	op.Write(w)
}

// SellWindow is sent to the client to open the vendor sell window.
type SellWindow struct {
	// Serial of the vendor we are buying from
	Vendor uo.Serial
	// List of items the player is allowed to sell to the vendor
	Items []ContentsItem
}

// Write implements the Packet interface.
func (p *SellWindow) Write(w io.Writer) {
	// Calculate packet length
	l := 9
	for _, i := range p.Items {
		l += 14 + len(i.Description)
	}
	dc.PutByte(w, 0x9E)                   // Packet ID
	dc.PutUint16(w, uint16(l))            // Packet length
	dc.PutUint32(w, uint32(p.Vendor))     // Vendor serial
	dc.PutUint16(w, uint16(len(p.Items))) // Number of items
	for _, i := range p.Items {
		dc.PutUint32(w, uint32(i.Serial))                   // Item serial
		dc.PutUint16(w, uint16(i.Graphic))                  // Item graphic
		putHue(w, i.Hue)                                    // Item hue
		dc.PutUint16(w, uint16(i.Amount))                   // Item amount
		dc.PutUint16(w, uint16(i.Price)/2)                  // Item price per unit
		dc.PutUint16(w, uint16(len(i.Description)))         // Length of the description string
		dc.PutStringN(w, i.Description, len(i.Description)) // Item descrition
	}
}

// NameResponse is sent to the client in response to a NameRequest.
type NameResponse struct {
	Serial uo.Serial // Serial of the object who's name we are sending.
	Name   string    // Name of the object
}

// Write implements the Packet interface.
func (p *NameResponse) Write(w io.Writer) {
	dc.PutByte(w, 0x98)                  // Packet ID
	dc.PutUint16(w, 37)                  // Packet length
	dc.PutUint32(w, uint32(p.Serial))    // Serial of the object
	dc.PutStringNWithNull(w, p.Name, 30) // Name of the object
}

// OPLPacket is sent in response to generic packet 0x10 and populates object
// tooltips.
type OPLPacket struct {
	Serial      uo.Serial // Serial of the object this packet is for
	Hash        uint32    // Hash of the packet
	Entries     []string  // List of all tooltip entries
	TailEntries []string  // List of all tooltip entries that should be appended to the tail
	buf         []byte    // Internal data buffer for caching
}

// Append adds an entry to the OPLPacket in the default font and color.
func (p *OPLPacket) Append(text string, tail bool) {
	if !tail {
		p.Entries = append(p.Entries, text)
	} else {
		p.TailEntries = append(p.TailEntries, text)
	}
}

// AppendColor adds an entry to the OPLPacket in the given color.
func (p *OPLPacket) AppendColor(c color.Color, text string, tail bool) {
	r, g, b, _ := c.RGBA()
	s := fmt.Sprintf("<basefont color=#%02X%02X%02X>%s</basefont>", r&0xFF, g&0xFF, b&0xFF, text)
	if !tail {
		p.Entries = append(p.Entries, s)
	} else {
		p.TailEntries = append(p.TailEntries, s)
	}
}

// Write implements the Packet interface.
func (p *OPLPacket) Write(w io.Writer) {
	p.Compile()
	w.Write(p.buf)
}

// Compile compiles the packet into the internal buffer if needed.
func (p *OPLPacket) Compile() {
	if len(p.buf) > 0 {
		return
	}
	// Write the packet to a temporary buffer
	b := bytes.NewBuffer(nil)
	dc.PutByte(b, 0xD6)               // Packet ID
	dc.Pad(b, 2)                      // Leave room for the packet length
	dc.PutUint16(b, 1)                // Unknown 1
	dc.PutUint32(b, uint32(p.Serial)) // Object serial
	dc.Pad(b, 6)                      // Unknown 2 and padding for hash value
	entries := append(p.Entries, p.TailEntries...)
	for _, e := range entries {
		dc.PutUint32(b, 1042971) // ~1_NOTHING~
		lrc := utf8.RuneCountInString(e)
		dc.PutUint16(b, uint16(lrc*2))  // String length in bytes
		dc.PutUTF16LEStringN(b, e, lrc) // Little-endian string, no NULL termination
	}
	dc.Pad(b, 4) // Terminating NULL cliloc ID
	// Calculate and insert the buffer hash
	p.buf = b.Bytes()
	p.Hash = crc32.ChecksumIEEE(p.buf)
	binary.BigEndian.PutUint32(p.buf[11:15], p.Hash)
	// Insert the buffer length
	binary.BigEndian.PutUint16(p.buf[1:3], uint16(len(p.buf)))
}

// OPLInfo is sent to notify the client of OPL revision changes.
type OPLInfo struct {
	Serial uo.Serial // Serial of the object this packet pertains to.
	Hash   uint32    // Hash of the OPL packet
}

// Write implements the Packet interface.
func (p *OPLInfo) Write(w io.Writer) {
	dc.PutByte(w, 0xDC)               // Packet ID
	dc.PutUint32(w, uint32(p.Serial)) // Object serial
	dc.PutUint32(w, p.Hash)           // OPL hash
}

// UpdateHealth is sent to notify the client of the current HP levels of another
// mobile. The health is normalized to a 4% resolution.
type UpdateHealth struct {
	Serial  uo.Serial // Serial of the object this packet pertains to.
	Hits    int       // Current hit points
	MaxHits int       // Maximum hit points
}

// Write implements the Packet interface.
func (p *UpdateHealth) Write(w io.Writer) {
	dc.PutByte(w, 0xA1) // Packet ID
	dc.PutUint32(w, uint32(p.Serial))
	dc.PutUint16(w, 25)
	r := float64(p.Hits) / float64(p.MaxHits)
	dc.PutUint16(w, uint16(r*25))
}

// TextEntryGUMP is sent to request a string of text from the client via a
// client-side GUMP.
type TextEntryGUMP struct {
	Serial      uo.Serial // Serial of the GUMP
	Value       string    // Current value of the text entry field
	Description string    // Description of the text requested
	CanCancel   bool      // If true allow the client to cancel the GUMP
	MaxLength   int       // Maximum response length
}

// Write implements the Packet interface.
func (p *TextEntryGUMP) Write(w io.Writer) {
	// Calculate length
	l := 19 + len(p.Value) + 1 + len(p.Description) + 1
	dc.PutByte(w, 0xAB)                     // Packet ID
	dc.PutUint16(w, uint16(l))              // Packet length
	dc.PutUint32(w, uint32(p.Serial))       // GUMP serial
	dc.Pad(w, 2)                            // Parent and button IDs?
	dc.PutUint16(w, uint16(len(p.Value)+1)) // Value length
	dc.PutString(w, p.Value)                // Value
	if p.CanCancel {                        // Cancel flag
		dc.PutByte(w, 1)
	} else {
		dc.PutByte(w, 0)
	}
	dc.PutByte(w, 1)                              // Style normal
	dc.PutUint32(w, uint32(p.MaxLength))          // Maximum length
	dc.PutUint16(w, uint16(len(p.Description)+1)) // Description text
	dc.PutString(w, p.Description)
}
