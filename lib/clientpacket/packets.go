package clientpacket

import (
	"encoding/binary"
	"unicode/utf16"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	packetFactory.Register(0x02, newWalkRequest)
	packetFactory.Register(0x06, newDoubleClick)
	packetFactory.Register(0x09, newSingleClick)
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
		return p.(Packet)
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
		length = int(getuint16(data[1:3]))
		pdat = data[3:length]
	} else if length == 0 {
		return newUnknownPacket(packetFactory.GetName(), uo.Serial(data[0]))
	} else {
		pdat = data[1:length]
	}
	return packetFactory.New(uo.Serial(data[0]), pdat)
}

func nullstr(buf []byte) string {
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	return string(buf)
}

func utf16str(b []byte) string {
	var utf [256]uint16
	ib := 0
	iu := 0
	for {
		if ib+1 >= len(b) || iu >= len(utf) {
			break
		}
		if b[ib+1] == 0 && b[ib] == 0 {
			break
		}
		utf[iu] = binary.BigEndian.Uint16(b[ib:])
		ib += 2
		iu++
	}
	return string(utf16.Decode(utf[:iu]))
}

func getuint16(buf []byte) uint16 {
	return binary.BigEndian.Uint16(buf)
}

func getuint32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
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
		Seed:         getuint32(in[0:4]),
		VersionMajor: int(getuint32(in[4:8])),
		VersionMinor: int(getuint32(in[8:12])),
		VersionPatch: int(getuint32(in[12:16])),
		VersionExtra: int(getuint32(in[16:20])),
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
		Username: nullstr(in[0:30]),
		Password: nullstr(in[30:60]),
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
		Index: int(getuint16(in[0:2])),
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
	Key []byte
}

func newGameServerLogin(in []byte) Packet {
	p := &GameServerLogin{
		Key:      in[:4],
		Username: nullstr(in[4:34]),
		Password: nullstr(in[34:64]),
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
		Slot: int(getuint32(in[64:68])),
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
		String: nullstr(in),
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
		Hue:  uo.Hue(getuint16(in[1:3])),
		Font: uo.Font(getuint16(in[3:5])),
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
		s.Text = nullstr(in[12+skip:])
		return s
	}
	s.Text = utf16str(in[9:])
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
		ID: uo.NewSerialFromData(in),
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
	s := uo.NewSerialFromData(in)
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
		PlayerMobileID:    uo.NewSerialFromData(in[5:]),
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
		FastWalkKey: getuint32(in[2:]),
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
	Graphic uo.Item
}

func newTargetResponse(in []byte) Packet {
	p := &TargetResponse{
		TargetType:   uo.TargetType(in[0]),
		TargetSerial: uo.Serial(getuint32(in[1:5])),
		CursorType:   uo.CursorType(in[5]),
		TargetObject: uo.Serial(getuint32(in[6:10])),
		X:            int(getuint16(in[10:12])),
		Y:            int(getuint16(in[12:14])),
		Z:            int(in[15]),
		Graphic:      uo.Item(getuint16(in[16:18])),
	}
	p.SetSerial(0x6C)
	return p
}
