package clientpacket

import (
	dc "github.com/qbradq/sharduo/lib/dataconv"
	"github.com/qbradq/sharduo/lib/uo"
)

// Global packet registry
var pf = &packetFactory{}

func init() {
	pf.Add(0x02, newWalkRequest)
	pf.Add(0x06, newDoubleClick)
	pf.Add(0x07, newLiftRequest)
	pf.Add(0x08, newDropRequest)
	pf.Add(0x09, newSingleClick)
	pf.Add(0x13, newWearItemRequest)
	pf.Add(0x34, newPlayerStatusRequest)
	pf.Add(0x5D, newCharacterLogin)
	pf.Add(0x6C, newTargetResponse)
	pf.Add(0x73, newPing)
	pf.Add(0x80, newAccountLogin)
	pf.Add(0x91, newGameServerLogin)
	pf.Add(0xA0, newSelectServer)
	pf.Add(0xAD, newSpeech)
	pf.Ignore(0xB5) // Open chat window request
	pf.Add(0xBD, newVersion)
	pf.Add(0xBF, newGeneralInformation)
	pf.Add(0xC8, newClientViewRange)
	pf.Add(0xEF, newLoginSeed)
	pf.Add(0xF0, newProtocolExtension)
}

// Packet is the interface all client packets implement.
type Packet interface {
	ID() byte
}

// New creates a new client packet based on data.
func New(data []byte) Packet {
	var pdat []byte

	length := InfoTable[data[0]].Length
	if length == -1 {
		length = int(dc.GetUint16(data[1:3]))
		pdat = data[3:length]
	} else if length == 0 {
		return newUnknownPacket("client-packets", data[0])
	} else {
		pdat = data[1:length]
	}
	return pf.New(data[0], pdat)
}

// basePacket provides common functionality to all client packets.
type basePacket struct {
	id byte
}

// ID implements the Packet interface.
func (p *basePacket) ID() byte { return p.id }

// UnsupportedPacket is sent when the packet being decoded does not have a
// constructor function yet.
type UnsupportedPacket struct {
	basePacket
	PType string
}

// NewUnsupportedPacket creates a new UnsupportedPacket properly initialized.
func NewUnsupportedPacket(ptype string, in []byte) *UnsupportedPacket {
	p := &UnsupportedPacket{
		basePacket: basePacket{id: in[0]},
		PType:      ptype,
	}
	return p
}

// UnknownPacket is sent when the packet being decoded has no length
// information. This dc.Puts the packet stream in an inconsistent state.
type UnknownPacket struct {
	basePacket
	PType string
}

func newUnknownPacket(ptype string, id byte) Packet {
	p := &UnknownPacket{
		basePacket: basePacket{id: id},
		PType:      ptype,
	}
	return p
}

// MalformedPacket is sent when there is a non-specific decoding error.
type MalformedPacket struct {
	basePacket
}

// newMalformedPacket returns a new initialized MalformedPacket
func newMalformedPacket(id byte) *MalformedPacket {
	return &MalformedPacket{
		basePacket: basePacket{id: id},
	}
}

// IgnoredPacket is a packet that we could fetch all the data for, but we do
// not have a struct nor a constructor, but it's OK for the server to ignore
// this.
type IgnoredPacket struct {
	basePacket
}

// LoginSeed is the first packet sent to the login server
type LoginSeed struct {
	basePacket
	// Connection seed
	Seed uint32
	// Version major part
	VersionMajor int
	// Version minor part
	VersionMinor int
	// Version patch part
	VersionPatch int
	// Version extra part
	VersionExtra int
}

func newLoginSeed(in []byte) Packet {
	return &LoginSeed{
		basePacket:   basePacket{id: 0xEF},
		Seed:         dc.GetUint32(in[0:4]),
		VersionMajor: int(dc.GetUint32(in[4:8])),
		VersionMinor: int(dc.GetUint32(in[8:12])),
		VersionPatch: int(dc.GetUint32(in[12:16])),
		VersionExtra: int(dc.GetUint32(in[16:20])),
	}
}

// AccountLogin is the second packet sent to the login server and attempts to
// authenticate with a clear-text username and password o_O
type AccountLogin struct {
	basePacket
	// Account username
	Username string
	// Account password in plain-text
	Password string
}

func newAccountLogin(in []byte) Packet {
	return &AccountLogin{
		basePacket: basePacket{id: 0x80},
		Username:   dc.NullString(in[0:30]),
		Password:   dc.NullString(in[30:60]),
	}
}

// SelectServer is used during the login process to request the connection
// details of one of the servers on the list.
type SelectServer struct {
	basePacket
	// Index is the index of the server on the list.
	Index int
}

func newSelectServer(in []byte) Packet {
	return &SelectServer{
		basePacket: basePacket{id: 0xA0},
		Index:      int(dc.GetUint16(in[0:2])),
	}
}

// GameServerLogin is used to authenticate to the game server in clear text.
type GameServerLogin struct {
	basePacket
	// Account username
	Username string
	// Account password in plain-text
	Password string
	// Key given by the login server
	Key uo.Serial
}

func newGameServerLogin(in []byte) Packet {
	return &GameServerLogin{
		basePacket: basePacket{id: 91},
		Key:        uo.Serial(dc.GetUint32(in[:4])),
		Username:   dc.NullString(in[4:34]),
		Password:   dc.NullString(in[34:64]),
	}
}

// CharacterLogin is used to request a character login.
type CharacterLogin struct {
	basePacket
	// Character slot chosen
	Slot int
}

func newCharacterLogin(in []byte) Packet {
	p := &CharacterLogin{
		basePacket: basePacket{id: 0x5D},
		Slot:       int(dc.GetUint32(in[64:68])),
	}
	return p
}

// Version is used to communicate to the server the client's version string.
type Version struct {
	basePacket
	// Version string
	String string
}

func newVersion(in []byte) Packet {
	// Length check not required, it can be nil
	p := &Version{
		basePacket: basePacket{id: 0xBD},
		String:     dc.NullString(in),
	}
	return p
}

// Ping is used for TCP keepalive and possibly latency determination.
type Ping struct {
	basePacket
	// Don't know what this is used for
	Key byte
}

func newPing(in []byte) Packet {
	p := &Ping{
		basePacket: basePacket{id: 0x73},
		Key:        in[0],
	}
	return p
}

// Speech is sent by the client to request speech.
type Speech struct {
	basePacket
	// Type of speech
	Type uo.SpeechType
	// Hue of the text
	Hue uo.Hue
	// Font of the text
	Font uo.Font
	// Text of the message
	Text string
}

func newSpeech(in []byte) Packet {
	if len(in) < 11 {
		return newMalformedPacket(0xAD)
	}
	s := &Speech{
		basePacket: basePacket{id: 0xAD},
		Type:       uo.SpeechType(in[0]),
		Hue:        uo.Hue(dc.GetUint16(in[1:3])),
		Font:       uo.Font(dc.GetUint16(in[3:5])),
	}
	if s.Type >= uo.SpeechTypeClientParsed {
		if len(in) < 13 {
			return newMalformedPacket(0xAD)
		}
		s.Type = s.Type - uo.SpeechTypeClientParsed
		numwords := int(in[9]) << 4
		numwords += int(in[10] >> 4)
		skip := ((numwords / 2) * 3) + (numwords % 2) - 1
		if len(in) < 13+skip {
			return newMalformedPacket(0xAD)
		}
		s.Text = dc.NullString(in[12+skip:])
		return s
	}
	s.Text = dc.UTF16String(in[9:])
	return s
}

// SingleClick is sent by the client when the player single-clicks an object
type SingleClick struct {
	basePacket
	// Object ID clicked on
	Object uo.Serial
}

func newSingleClick(in []byte) Packet {
	p := &SingleClick{
		basePacket: basePacket{id: 0x09},
		Object:     uo.Serial(dc.GetUint32(in)),
	}
	return p
}

// DoubleClick is sent by the client when the player double-clicks an object
type DoubleClick struct {
	basePacket
	// Object ID clicked on
	Object uo.Serial
	// WantPaperDoll is true if this is a request for our own paper doll
	WantPaperDoll bool
}

func newDoubleClick(in []byte) Packet {
	s := uo.Serial(dc.GetUint32(in[:4]))
	isPaperDoll := s.IsSelf()
	s = s.StripSelfFlag()
	p := &DoubleClick{
		basePacket:    basePacket{id: 0x06},
		Object:        s,
		WantPaperDoll: isPaperDoll,
	}
	return p
}

// PlayerStatusRequest is sent by the client to request status and skills
// updates.
type PlayerStatusRequest struct {
	basePacket
	// Type of the request
	StatusRequestType uo.StatusRequestType
	// ID of the player's mobile
	PlayerMobileID uo.Serial
}

func newPlayerStatusRequest(in []byte) Packet {
	p := &PlayerStatusRequest{
		basePacket:        basePacket{id: 0x34},
		StatusRequestType: uo.StatusRequestType(in[4]),
		PlayerMobileID:    uo.Serial(dc.GetUint32(in[5:])),
	}
	return p
}

// ViewRange is sent by the client to request a new view range.
type ViewRange struct {
	basePacket
	// Requested view range, clamped to between 4 and 18 inclusive
	Range int16
}

func newClientViewRange(in []byte) Packet {
	r := uo.BoundViewRange(int16(in[0]))
	p := &ViewRange{
		basePacket: basePacket{id: 0xC8},
		Range:      r,
	}
	return p
}

// WalkRequest is sent by the client to request walking or running in a
// direction.
type WalkRequest struct {
	basePacket
	// Direction to walk
	Direction uo.Direction
	// If true this is a run request
	IsRunning bool
	// Walk sequence number
	Sequence int
	// Fast-walk prevention key
	FastWalkKey uint32
}

func newWalkRequest(in []byte) Packet {
	d := uo.Direction(in[0])
	r := d.IsRunning()
	d = d.StripRunningFlag()
	p := &WalkRequest{
		basePacket:  basePacket{id: 0x02},
		Direction:   d,
		IsRunning:   r,
		Sequence:    int(in[1]),
		FastWalkKey: dc.GetUint32(in[2:]),
	}
	return p
}

// TargetResponse is sent by the client to respond to a tardc.Geting cursor
type TargetResponse struct {
	basePacket
	// Target type
	TargetType uo.TargetType
	// Serial of this targeting request
	TargetSerial uo.Serial
	// Cursor type
	CursorType uo.CursorType
	// TargetObject is the serial of the object clicked on, or uo.SerialZero if
	// no object was targeted.
	TargetObject uo.Serial
	// Location of the target.
	Location uo.Location
	// Graphic of the object clicked, if any
	Graphic uo.Graphic
}

func newTargetResponse(in []byte) Packet {
	p := &TargetResponse{
		basePacket:   basePacket{id: 0x6C},
		TargetType:   uo.TargetType(in[0]),
		TargetSerial: uo.Serial(dc.GetUint32(in[1:5])),
		CursorType:   uo.CursorType(in[5]),
		TargetObject: uo.Serial(dc.GetUint32(in[6:10])),
		Location: uo.Location{
			X: int16(dc.GetUint16(in[10:12])),
			Y: int16(dc.GetUint16(in[12:14])),
			Z: int8(in[15]),
		},
		Graphic: uo.Graphic(dc.GetUint16(in[16:18])),
	}
	return p
}

// LiftRequest is sent when the player lifts an item with the cursor.
type LiftRequest struct {
	basePacket
	// Serial of the object to lift
	Item uo.Serial
	// Amount to lift
	Amount int
}

func newLiftRequest(in []byte) Packet {
	p := &LiftRequest{
		basePacket: basePacket{id: 0x07},
		Item:       uo.Serial(dc.GetUint32(in[0:4])),
		Amount:     int(dc.GetUint16(in[4:6])),
	}
	return p
}

// DropRequest is sent when the player drops an item from the cursor.
type DropRequest struct {
	basePacket
	// Serial of the object to drop
	Item uo.Serial
	// Location to drop the item
	Location uo.Location
	// Serial of the container to drop the item into. uo.SerialSystem means to
	// drop to the world.
	Container uo.Serial
}

func newDropRequest(in []byte) Packet {
	p := &DropRequest{
		basePacket: basePacket{id: 0x08},
		Item:       uo.Serial(dc.GetUint32(in[0:4])),
		Location: uo.Location{
			X: int16(dc.GetUint16(in[4:6])),
			Y: int16(dc.GetUint16(in[6:8])),
			Z: int8(in[8]),
		},
		// Skip one byte for the grid index
		Container: uo.Serial(dc.GetUint32(in[10:14])),
	}
	return p
}

// WearItemRequest is sent when the player drops an item onto a paper doll
type WearItemRequest struct {
	basePacket
	// Serial of the item to wear
	Item uo.Serial
	// Serial of the mobile to equip the item to
	Wearer uo.Serial
}

func newWearItemRequest(in []byte) Packet {
	p := &WearItemRequest{
		basePacket: basePacket{id: 0x13},
		Item:       uo.Serial(dc.GetUint32(in[0:4])),
		Wearer:     uo.Serial(dc.GetUint32(in[5:9])),
	}
	return p
}

// ProtocolExtension is sent by ClassicUO to query party and guild member
// positions for the built-in world map. I think the latest versions of Razor,
// UO Steam, and UOAM/UOPS do this as well.
type ProtocolExtension struct {
	basePacket
	// Type of request
	RequestType uo.ProtocolExtensionRequest
}

func newProtocolExtension(in []byte) Packet {
	p := &ProtocolExtension{
		basePacket:  basePacket{id: 0xF0},
		RequestType: uo.ProtocolExtensionRequest(in[0]),
	}
	return p
}
