package clientpacket

import (
	. "github.com/qbradq/sharduo/lib/dataconv"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	packetFactory.Register(0x02, newWalkRequest)
	packetFactory.Register(0x06, newDoubleClick)
	packetFactory.Register(0x07, newLiftRequest)
	packetFactory.Register(0x08, newDropRequest)
	packetFactory.Register(0x09, newSingleClick)
	packetFactory.Register(0x13, newWearItemRequest)
	packetFactory.Register(0x34, newPlayerStatusRequest)
	packetFactory.Register(0x5D, newCharacterLogin)
	packetFactory.Register(0x6C, newTargetResponse)
	packetFactory.Register(0x73, newPing)
	packetFactory.Register(0x80, newAccountLogin)
	packetFactory.Register(0x91, newGameServerLogin)
	packetFactory.Register(0xA0, newSelectServer)
	packetFactory.Register(0xAD, newSpeech)
	packetFactory.Ignore(0xB5) // Open chat window request
	packetFactory.Register(0xBD, newVersion)
	packetFactory.Register(0xBF, newGeneralInformation)
	packetFactory.Register(0xC8, newClientViewRange)
	packetFactory.Register(0xEF, newLoginSeed)
	packetFactory.Register(0xF0, newProtocolExtension)
}

// Packet is the interface all client packets implement.
type Packet interface {
	util.Serialer
}

// PacketFactory creates client packets from slices of bytes
type PacketFactory struct {
	util.Factory[uo.Serial, []byte, Packet]
}

// NewPacketFactory creates a new PacketFactory ready for use
func NewPacketFactory(name string) *PacketFactory {
	return &PacketFactory{
		*util.NewFactory[uo.Serial, []byte, Packet](name),
	}
}

// Ignore ignores the given packet ID
func (f *PacketFactory) Ignore(id uo.Serial) {
	f.Add(id, func(in []byte) Packet {
		p := &IgnoredPacket{}
		p.SetSerial(id)
		return p
	})
}

// New creates a new client packet.
func (f *PacketFactory) New(id uo.Serial, in []byte) Packet {
	if p := f.Factory.New(id, in); p != nil {
		return p
	}
	up := NewUnsupportedPacket(f.GetName(), in)
	up.SetSerial(id)
	return up
}

var packetFactory = NewPacketFactory("client")

// New creates a new client packet based on data.
func New(data []byte) Packet {
	var pdat []byte

	length := InfoTable[data[0]].Length
	if length == -1 {
		length = int(GetUint16(data[1:3]))
		pdat = data[3:length]
	} else if length == 0 {
		return newUnknownPacket(packetFactory.GetName(), uo.Serial(data[0]))
	} else {
		pdat = data[1:length]
	}
	return packetFactory.New(uo.Serial(data[0]), pdat)
}

// UnsupportedPacket is sent when the packet being decoded does not have a
// constructor function yet.
type UnsupportedPacket struct {
	util.BaseSerialer
	PType string
}

// NewUnsupportedPacket creates a new UnsupportedPacket properly initialized.
func NewUnsupportedPacket(ptype string, in []byte) *UnsupportedPacket {

	p := &UnsupportedPacket{
		PType: ptype,
	}
	p.SetSerial(uo.Serial(in[0]))
	return p
}

// UnknownPacket is sent when the packet being decoded has no length
// information. This puts the packet stream in an inconsistent state.
type UnknownPacket struct {
	util.BaseSerialer
	PType string
}

func newUnknownPacket(ptype string, id uo.Serial) Packet {
	p := &UnknownPacket{
		PType: ptype,
	}
	p.SetSerial(id)
	return p
}

// MalformedPacket is sent when there is a non-specific decoding error.
type MalformedPacket struct {
	util.BaseSerialer
}

// newMalformedPacket returns a new initialized MalformedPacket
func newMalformedPacket(id uo.Serial) *MalformedPacket {
	p := &MalformedPacket{}
	p.SetSerial(id)
	return p
}

// IgnoredPacket is a packet that we could fetch all the data for, but we do
// not have a struct nor a constructor, but it's OK for the server to ignore
// this.
type IgnoredPacket struct {
	util.BaseSerialer
}

// LoginSeed is the first packet sent to the login server
type LoginSeed struct {
	util.BaseSerialer
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
	p := &LoginSeed{
		Seed:         GetUint32(in[0:4]),
		VersionMajor: int(GetUint32(in[4:8])),
		VersionMinor: int(GetUint32(in[8:12])),
		VersionPatch: int(GetUint32(in[12:16])),
		VersionExtra: int(GetUint32(in[16:20])),
	}
	p.SetSerial(0xEF)
	return p
}

// AccountLogin is the second packet sent to the login server and attempts to
// authenticate with a clear-text username and password o_O
type AccountLogin struct {
	util.BaseSerialer
	// Account username
	Username string
	// Account password in plain-text
	Password string
}

func newAccountLogin(in []byte) Packet {
	p := &AccountLogin{
		Username: NullString(in[0:30]),
		Password: NullString(in[30:60]),
	}
	p.SetSerial(0x80)
	return p
}

// SelectServer is used during the login process to request the connection
// details of one of the servers on the list.
type SelectServer struct {
	util.BaseSerialer
	// Index is the index of the server on the list.
	Index int
}

func newSelectServer(in []byte) Packet {
	p := &SelectServer{
		Index: int(GetUint16(in[0:2])),
	}
	p.SetSerial(0xA0)
	return p
}

// GameServerLogin is used to authenticate to the game server in clear text.
type GameServerLogin struct {
	util.BaseSerialer
	// Account username
	Username string
	// Account password in plain-text
	Password string
	// Key given by the login server
	Key uo.Serial
}

func newGameServerLogin(in []byte) Packet {
	p := &GameServerLogin{
		Key:      uo.Serial(GetUint32(in[:4])),
		Username: NullString(in[4:34]),
		Password: NullString(in[34:64]),
	}
	p.SetSerial(0x91)
	return p
}

// CharacterLogin is used to request a character login.
type CharacterLogin struct {
	util.BaseSerialer
	// Character slot chosen
	Slot int
}

func newCharacterLogin(in []byte) Packet {
	p := &CharacterLogin{
		Slot: int(GetUint32(in[64:68])),
	}
	p.SetSerial(0x5D)
	return p
}

// Version is used to communicate to the server the client's version string.
type Version struct {
	util.BaseSerialer
	// Version string
	String string
}

func newVersion(in []byte) Packet {
	// Length check not required, it can be nil
	p := &Version{
		String: NullString(in),
	}
	p.SetSerial(0xBD)
	return p
}

// Ping is used for TCP keepalive and possibly latency determination.
type Ping struct {
	util.BaseSerialer
	// Don't know what this is used for
	Key byte
}

func newPing(in []byte) Packet {
	p := &Ping{
		Key: in[0],
	}
	p.SetSerial(0x73)
	return p
}

// Speech is sent by the client to request speech.
type Speech struct {
	util.BaseSerialer
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
		Type: uo.SpeechType(in[0]),
		Hue:  uo.Hue(GetUint16(in[1:3])),
		Font: uo.Font(GetUint16(in[3:5])),
	}
	s.SetSerial(0xAD)
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
		s.Text = NullString(in[12+skip:])
		return s
	}
	s.Text = UTF16String(in[9:])
	return s
}

// SingleClick is sent by the client when the player single-clicks an object
type SingleClick struct {
	util.BaseSerialer
	// Object ID clicked on
	ID uo.Serial
}

func newSingleClick(in []byte) Packet {
	p := &SingleClick{
		ID: uo.Serial(GetUint32(in)),
	}
	p.SetSerial(0x09)
	return p
}

// DoubleClick is sent by the client when the player double-clicks an object
type DoubleClick struct {
	util.BaseSerialer
	// Object ID clicked on
	ID uo.Serial
	// IsSelf is true if the requested object is the player's mobile
	IsSelf bool
}

func newDoubleClick(in []byte) Packet {
	s := uo.Serial(GetUint32(in[:4]))
	isSelf := s.IsSelf()
	s = s.StripSelfFlag()
	p := &DoubleClick{
		ID:     s,
		IsSelf: isSelf,
	}
	p.SetSerial(0x06)
	return p
}

// PlayerStatusRequest is sent by the client to request status and skills
// updates.
type PlayerStatusRequest struct {
	util.BaseSerialer
	// Type of the request
	StatusRequestType uo.StatusRequestType
	// ID of the player's mobile
	PlayerMobileID uo.Serial
}

func newPlayerStatusRequest(in []byte) Packet {
	p := &PlayerStatusRequest{
		StatusRequestType: uo.StatusRequestType(in[4]),
		PlayerMobileID:    uo.Serial(GetUint32(in[5:])),
	}
	p.SetSerial(0x34)
	return p
}

// ClientViewRange is sent by the client to request a new view range.
type ClientViewRange struct {
	util.BaseSerialer
	// Requested view range, clamped to between 4 and 18 inclusive
	Range int
}

func newClientViewRange(in []byte) Packet {
	r := int(in[0])
	if r < 4 {
		r = 4
	}
	if r > 18 {
		r = 18
	}
	p := &ClientViewRange{
		Range: r,
	}
	p.SetSerial(0xC8)
	return p
}

// WalkRequest is sent by the client to request walking or running in a
// direction.
type WalkRequest struct {
	util.BaseSerialer
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
		Direction:   d,
		IsRunning:   r,
		Sequence:    int(in[1]),
		FastWalkKey: GetUint32(in[2:]),
	}
	p.SetSerial(0x02)
	return p
}

// TargetResponse is sent by the client to respond to a targeting cursor
type TargetResponse struct {
	util.BaseSerialer
	// Target type
	TargetType uo.TargetType
	// Serial of this targeting request
	TargetSerial uo.Serial
	// Cursor type
	CursorType uo.CursorType
	// TargetObject is the serial of the object clicked on, or uo.SerialZero if
	// no object was targeted.
	TargetObject uo.Serial
	// The X location of the target
	X int
	// The Y location of the target
	Y int
	// The Z location of the target
	Z int
	// Graphic of the object clicked, if any
	Graphic uo.Graphic
}

func newTargetResponse(in []byte) Packet {
	p := &TargetResponse{
		TargetType:   uo.TargetType(in[0]),
		TargetSerial: uo.Serial(GetUint32(in[1:5])),
		CursorType:   uo.CursorType(in[5]),
		TargetObject: uo.Serial(GetUint32(in[6:10])),
		X:            int(GetUint16(in[10:12])),
		Y:            int(GetUint16(in[12:14])),
		Z:            int(in[15]),
		Graphic:      uo.Graphic(GetUint16(in[16:18])),
	}
	p.SetSerial(0x6C)
	return p
}

// LiftRequest is sent when the player lifts an item with the cursor.
type LiftRequest struct {
	util.BaseSerialer
	// Serial of the object to lift
	Item uo.Serial
	// Amount to lift
	Amount int
}

func newLiftRequest(in []byte) Packet {
	p := &LiftRequest{
		Item:   uo.Serial(GetUint32(in[0:4])),
		Amount: int(GetUint16(in[4:6])),
	}
	p.SetSerial(0x07)
	return p
}

// DropRequest is sent when the player drops an item from the cursor.
type DropRequest struct {
	util.BaseSerialer
	// Serial of the object to drop
	Item uo.Serial
	// X location to drop the item
	X int
	// Y location to drop the item
	Y int
	// Z location to drop the item
	Z int
	// Serial of the container to drop the item into. uo.SerialSystem means to
	// drop to the world.
	Container uo.Serial
}

func newDropRequest(in []byte) Packet {
	p := &DropRequest{
		Item: uo.Serial(GetUint32(in[0:4])),
		X:    int(GetUint16(in[4:6])),
		Y:    int(GetUint16(in[6:8])),
		Z:    int(int8(in[8])),
		// Skip one byte for the grid index
		Container: uo.Serial(GetUint32(in[10:14])),
	}
	p.SetSerial(0x08)
	return p
}

// WearItemRequest is sent when the player drops an item onto a paper doll
type WearItemRequest struct {
	util.BaseSerialer
	// Serial of the item to wear
	Item uo.Serial
	// Serial of the mobile to equip the item to
	Wearer uo.Serial
}

func newWearItemRequest(in []byte) Packet {
	p := &WearItemRequest{
		Item:   uo.Serial(GetUint32(in[0:4])),
		Wearer: uo.Serial(GetUint32(in[5:9])),
	}
	p.SetSerial(0x13)
	return p
}

// ProtocolExtension is sent by ClassicUO to query party and guild member
// positions for the built-in world map. I think the latest versions of Razor,
// UO Steam, and UOAM/UOPS do this as well.
type ProtocolExtension struct {
	util.BaseSerialer
	// Type of request
	RequestType uo.ProtocolExtensionRequest
}

func newProtocolExtension(in []byte) Packet {
	p := &ProtocolExtension{
		RequestType: uo.ProtocolExtensionRequest(in[0]),
	}
	p.SetSerial(0xF0)
	return p
}
