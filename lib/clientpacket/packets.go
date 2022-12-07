package clientpacket

import (
	"encoding/binary"
	"unicode/utf16"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	packetFactory.Add(0x02, newWalkRequest)
	packetFactory.Add(0x06, newDoubleClick)
	packetFactory.Add(0x09, newSingleClick)
	packetFactory.Add(0x34, newPlayerStatusRequest)
	packetFactory.Add(0x5D, newCharacterLogin)
	packetFactory.Add(0x73, newPing)
	packetFactory.Add(0x80, newAccountLogin)
	packetFactory.Add(0x91, newGameServerLogin)
	packetFactory.Add(0xA0, newSelectServer)
	packetFactory.Add(0xAD, newSpeech)
	packetFactory.Ignore(0xB5) // Open chat window request
	packetFactory.Add(0xBD, newVersion)
	packetFactory.Add(0xBF, newGeneralInformation)
	packetFactory.Add(0xC8, newClientViewRange)
	packetFactory.Add(0xEF, newLoginSeed)
}

// Packet is the interface all client packets implement.
type Packet interface {
	// GetID returns the packet ID byte.
	GetID() int
	// SetID sets the packet ID.
	setId(id int)
}

// PacketFactory creates client packets from slices of bytes
type PacketFactory struct {
	util.Factory[int, []byte]
}

// NewPacketFactory creates a new PacketFactory ready for use
func NewPacketFactory(name string) *PacketFactory {
	return &PacketFactory{
		*util.NewFactory[int, []byte](name),
	}
}

// Ignore ignores the given packet ID
func (f *PacketFactory) Ignore(id int) {
	f.Add(id, func(in []byte) any {
		return &IgnoredPacket{
			Base: Base{ID: id},
		}
	})
}

// New creates a new client packet.
func (f *PacketFactory) New(id int, in []byte) Packet {
	if p := f.Factory.New(id, in); p != nil {
		return p.(Packet)
	}
	up := NewUnsupportedPacket(f.GetName(), in)
	up.ID = id
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
		return newUnknownPacket(packetFactory.GetName(), int(data[0]))
	} else {
		pdat = data[1:length]
	}
	return packetFactory.New(int(data[0]), pdat)
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

// Base is the base struct for all client packets.
type Base struct {
	// ID is the packet ID byte.
	ID int
}

// GetID implements the Packet interface.
func (p *Base) GetID() int {
	return p.ID
}

// setId implements the Packet interface.
func (p *Base) setId(id int) {
	p.ID = id
}

// UnsupportedPacket is sent when the packet being decoded does not have a
// constructor function yet.
type UnsupportedPacket struct {
	Base
	PType string
}

// NewUnsupportedPacket creates a new UnsupportedPacket properly initialized.
func NewUnsupportedPacket(ptype string, in []byte) *UnsupportedPacket {
	return &UnsupportedPacket{
		Base:  Base{ID: int(in[0])},
		PType: ptype,
	}
}

// UnknownPacket is sent when the packet being decoded has no length
// information. This puts the packet stream in an inconsistent state.
type UnknownPacket struct {
	Base
	PType string
}

func newUnknownPacket(ptype string, id int) Packet {
	return &UnknownPacket{
		Base:  Base{ID: id},
		PType: ptype,
	}
}

// MalformedPacket is sent when there is a non-specific decoding error.
type MalformedPacket struct {
	Base
}

// IgnoredPacket is a packet that we could fetch all the data for, but we do
// not have a struct nor a constructor, but it's OK for the server to ignore
// this.
type IgnoredPacket struct {
	Base
}

// LoginSeed is the first packet sent to the login server
type LoginSeed struct {
	Base
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

func newLoginSeed(in []byte) any {
	return &LoginSeed{
		Base:         Base{ID: 0xEF},
		Seed:         getuint32(in[0:4]),
		VersionMajor: int(getuint32(in[4:8])),
		VersionMinor: int(getuint32(in[8:12])),
		VersionPatch: int(getuint32(in[12:16])),
		VersionExtra: int(getuint32(in[16:20])),
	}
}

// AccountLogin is the second packet sent to the login server and attempts to
// authenticate with a clear-text username and password o_O
type AccountLogin struct {
	Base
	// Account username
	Username string
	// Account password in plain-text
	Password string
}

func newAccountLogin(in []byte) any {
	return &AccountLogin{
		Base:     Base{ID: 0x80},
		Username: nullstr(in[0:30]),
		Password: nullstr(in[30:60]),
	}
}

// SelectServer is used during the login process to request the connection
// details of one of the servers on the list.
type SelectServer struct {
	Base
	// Index is the index of the server on the list.
	Index int
}

func newSelectServer(in []byte) any {
	return &SelectServer{
		Base:  Base{ID: 0xA0},
		Index: int(getuint16(in[0:2])),
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

func newGameServerLogin(in []byte) any {
	return &GameServerLogin{
		Base:     Base{ID: 0x91},
		Key:      in[:4],
		Username: nullstr(in[4:34]),
		Password: nullstr(in[34:64]),
	}
}

// CharacterLogin is used to request a character login.
type CharacterLogin struct {
	Base
	// Character slot chosen
	Slot int
}

func newCharacterLogin(in []byte) any {
	return &CharacterLogin{
		Base: Base{ID: 0x5D},
		Slot: int(getuint32(in[64:68])),
	}
}

// Version is used to communicate to the server the client's version string.
type Version struct {
	Base
	// Version string
	String string
}

func newVersion(in []byte) any {
	// Length check no required, in can be nil
	return &Version{
		Base:   Base{ID: 0xBD},
		String: nullstr(in),
	}
}

// Ping is used for TCP keepalive and possibly latency determination.
type Ping struct {
	Base
	// Don't know what this is used for
	Key byte
}

func newPing(in []byte) any {
	return &Ping{
		Base: Base{ID: 0x73},
		Key:  in[0],
	}
}

// Speech is sent by the client to request speech.
type Speech struct {
	Base
	// Type of speech
	Type uo.SpeechType
	// Hue of the text
	Hue uo.Hue
	// Font of the text
	Font uo.Font
	// Text of the message
	Text string
}

func newSpeech(in []byte) any {
	if len(in) < 11 {
		return &MalformedPacket{
			Base: Base{ID: 0xAD},
		}
	}
	s := &Speech{
		Base: Base{ID: 0xAD},
		Type: uo.SpeechType(in[0]),
		Hue:  uo.Hue(getuint16(in[1:3])),
		Font: uo.Font(getuint16(in[3:5])),
	}
	if s.Type >= uo.SpeechTypeClientParsed {
		if len(in) < 13 {
			return &MalformedPacket{
				Base: Base{ID: 0xAD},
			}
		}
		s.Type = s.Type - uo.SpeechTypeClientParsed
		numwords := int(in[9]) << 4
		numwords += int(in[10] >> 4)
		// mulidx := int(in[11]&0xf0) << 4
		// mulidx += int(in[12])
		skip := ((numwords / 2) * 3) + (numwords % 2) - 1
		if len(in) < 13+skip {
			return &MalformedPacket{
				Base: Base{ID: 0xAD},
			}
		}
		s.Text = nullstr(in[12+skip:])
		return s
	}
	s.Text = utf16str(in[9:])
	return s
}

// SingleClick is sent by the client when the player single-clicks an object
type SingleClick struct {
	Base
	// Object ID clicked on
	ID uo.Serial
}

func newSingleClick(in []byte) any {
	return &SingleClick{
		Base: Base{ID: 0x09},
		ID:   uo.NewSerialFromData(in),
	}
}

// DoubleClick is sent by the client when the player double-clicks an object
type DoubleClick struct {
	Base
	// Object ID clicked on
	ID uo.Serial
	// IsSelf is true if the requested object is the player's mobile
	IsSelf bool
}

func newDoubleClick(in []byte) any {
	s := uo.NewSerialFromData(in)
	isSelf := s.IsSelf()
	s = s.StripSelfFlag()
	return &DoubleClick{
		Base:   Base{ID: 0x06},
		ID:     s,
		IsSelf: isSelf,
	}
}

// PlayerStatusRequest is sent by the client to request status and skills
// updates.
type PlayerStatusRequest struct {
	Base
	// Type of the request
	StatusRequestType uo.StatusRequestType
	// ID of the player's mobile
	PlayerMobileID uo.Serial
}

func newPlayerStatusRequest(in []byte) any {
	return &PlayerStatusRequest{
		Base:              Base{ID: 0x34},
		StatusRequestType: uo.StatusRequestType(in[4]),
		PlayerMobileID:    uo.NewSerialFromData(in[5:]),
	}
}

// ClientViewRange is sent by the client to request a new view range.
type ClientViewRange struct {
	Base
	// Requested view range, clamped to between 4 and 18 inclusive
	Range int
}

func newClientViewRange(in []byte) any {
	r := int(in[0])
	if r < 4 {
		r = 4
	}
	if r > 18 {
		r = 18
	}
	return &ClientViewRange{
		Base:  Base{ID: 0xc8},
		Range: r,
	}
}

// WalkRequest is sent by the client to request walking or running in a
// direction.
type WalkRequest struct {
	Base
	// Direction to walk
	Direction uo.Direction
	// If true this is a run request
	IsRunning bool
	// Walk sequence number
	Sequence int
	// Fast-walk prevention key
	FastWalkKey uint32
}

func newWalkRequest(in []byte) any {
	d := uo.Direction(in[0])
	r := d.IsRunning()
	d = d.StripRunningFlag()
	return &WalkRequest{
		Base:        Base{ID: 0x02},
		Direction:   d,
		IsRunning:   r,
		Sequence:    int(in[1]),
		FastWalkKey: getuint32(in[2:]),
	}
}

// TargetResponse is sent by the client to respond to a targeting cursor
type TargetResponse struct {
	Base
	// Target type
	TargetType uo.TargetCursorType
	// Serial of this targeting request
	Serial uo.Serial
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

func newTargetResponse(in []byte) any {
	return &TargetResponse{
		Base: Base{ID: 0x6C},
		TargetType: uo.TargetType(in[0]),
		Serial:  uo.Serial(getuint32(w, in[1:5])),
		CursorType: uo.CursorType(in[5]),
		TargetObject: uo.Serial(getuint32(w, in[6:10])),
		X: int(getuint16(w, in[10:12])),
		Y: int(getuint16(w, in[12:14])),
		Z: int(in[15]),
		Graphic: uo.Item(getuint16(w, [16:18])),
	}
}
