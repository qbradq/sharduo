package uo

import (
	"encoding/binary"
	"net"
)

// ServerPacket is the interface all server packet objects implement
type ServerPacket interface {
	// Bytes returns the underlying byte slice
	Bytes() []byte
}

func putByte(p *[]byte, v byte) {
	*p = append(*p, v)
}

func putUInt16(p *[]byte, v uint16) {
	l := len(*p)
	binary.BigEndian.PutUint16((*p)[l:l+2], v)
	*p = (*p)[:l+2]
}

func backPatchLength(p *[]byte) {
	binary.BigEndian.PutUint16((*p)[1:3], uint16(len(*p)))
}

func putUInt32(p *[]byte, v uint32) {
	l := len(*p)
	binary.BigEndian.PutUint32((*p)[l:l+4], v)
	*p = (*p)[:l+4]
}

func putASCII(p *[]byte, v string, length int) {
	var left = length
	for i, cp := range v {
		if i >= length {
			break
		}
		*p = append(*p, byte(cp&0x7f))
		left--
	}
	for left > 0 {
		*p = append(*p, 0)
		left--
	}
}

// PutBytes writes a byte slice
func putBytes(p *[]byte, b []byte) {
	*p = append(*p, b...)
}

// Fill writes a byte a number of times
func fill(p *[]byte, v byte, n int) {
	for n > 0 {
		*p = append(*p, v)
		n--
	}
}

// ServerPacketSetMap is used to set the map file loaded by the client
type ServerPacketSetMap []byte

// Bytes returns the underlying byte slice
func (p ServerPacketSetMap) Bytes() []byte {
	return ([]byte)(p)
}

// NewServerPacketSetMap creates a new ServerPacketSetMap
func NewServerPacketSetMap(p []byte, mapID byte) ServerPacketSetMap {
	putByte(&p, 0xbf)
	putUInt16(&p, 6)
	putUInt16(&p, 0x08)
	putByte(&p, mapID)
	return ServerPacketSetMap(p)
}

// ServerPacketPlayerBody is used to set the Body ID of the player's mobile
type ServerPacketPlayerBody []byte

// Bytes returns the underlying byte slice
func (p ServerPacketPlayerBody) Bytes() []byte {
	return ([]byte)(p)
}

// NewServerPacketPlayerBody creates a new packet
func NewServerPacketPlayerBody(p []byte, id Serial, body, x, y uint16, z int8, dir Dir, serverWidth, serverHeight uint16) ServerPacketPlayerBody {
	putByte(&p, 0x1b)
	putUInt32(&p, uint32(id))
	fill(&p, 0, 4)
	putUInt16(&p, body)
	putUInt16(&p, x)
	putUInt16(&p, y)
	putByte(&p, 0)
	putByte(&p, byte(z))
	putByte(&p, byte(dir))
	fill(&p, 0, 9)
	putUInt16(&p, serverWidth-8)
	putUInt16(&p, serverHeight)
	fill(&p, 0, 6)
	return ServerPacketPlayerBody(p)
}

// ServerPacketLoginDenied is used to inform the client that a login attempt
// has been rejected
type ServerPacketLoginDenied []byte

// Bytes returns the underlying byte slice
func (p ServerPacketLoginDenied) Bytes() []byte {
	return ([]byte)(p)
}

// NewServerPacketLoginDenied creates a new packet
func NewServerPacketLoginDenied(p []byte, reason LoginDeniedReason) ServerPacketLoginDenied {
	putByte(&p, 0x82)
	putByte(&p, byte(reason))
	return ServerPacketLoginDenied(p)
}

// ServerPacketLoginComplete is used to inform the client that the character
// login process is complete
type ServerPacketLoginComplete []byte

// Bytes returns the underlying byte slice
func (p ServerPacketLoginComplete) Bytes() []byte {
	return ([]byte)(p)
}

// NewServerPacketLoginComplete creates a new packet
func NewServerPacketLoginComplete(p []byte) ServerPacketLoginComplete {
	putByte(&p, 0x55)
	return ServerPacketLoginComplete(p)
}

// ServerPacketServerList is used to display the server list to the client
type ServerPacketServerList []byte

// Bytes returns the underlying byte slice
func (p ServerPacketServerList) Bytes() []byte {
	return ([]byte)(p)
}

// AddServer adds a server entry
func (p *ServerPacketServerList) AddServer(name string, full, tz byte, addr net.IPAddr) {
	ip := addr.IP.To4()
	putUInt16((*[]byte)(p), uint16((*p)[5]))
	(*p)[5]++
	putASCII((*[]byte)(p), name, 32)
	putByte((*[]byte)(p), 0)
	putByte((*[]byte)(p), 0)
	putByte((*[]byte)(p), ip[3])
	putByte((*[]byte)(p), ip[2])
	putByte((*[]byte)(p), ip[1])
	putByte((*[]byte)(p), ip[0])
}

// Finish finished writing the packet
func (p *ServerPacketServerList) Finish() {
	backPatchLength((*[]byte)(p))
}

// NewServerPacketServerList creates a new packet
func NewServerPacketServerList(p []byte) ServerPacketServerList {
	putByte(&p, 0xa8)
	fill(&p, 0, 2)
	putByte(&p, 0x5d)
	fill(&p, 0, 2)
	return ServerPacketServerList(p)
}

// ServerPacketDrawPlayer is used to update player position, facing, and hue
// similar to the 0x77 packet but only sent to the client controlling the
// mobile being updated
type ServerPacketDrawPlayer []byte

// Bytes returns the underlying byte slice
func (p ServerPacketDrawPlayer) Bytes() []byte {
	return ([]byte)(p)
}

// NewServerPacketDrawPlayer creates a new packet
func NewServerPacketDrawPlayer(p []byte, id Serial, body uint16, hue Hue, status StatusFlag, x, y uint16, z int8, dir Dir) ServerPacketDrawPlayer {
	putByte(&p, 0x20)
	putUInt32(&p, uint32(id))
	putByte(&p, 0)
	putUInt16(&p, uint16(hue))
	putByte(&p, byte(status))
	putUInt16(&p, x)
	putUInt16(&p, y)
	fill(&p, 0, 2)
	putByte(&p, byte(dir))
	putByte(&p, byte(z))
	return ServerPacketDrawPlayer(p)
}

// ServerPacketDrawMobile is used to send the animation composit information
// for a mobile to the client.
type ServerPacketDrawMobile []byte

// Bytes returns the underlying byte slice
func (p ServerPacketDrawMobile) Bytes() []byte {
	return ([]byte)(p)
}

// AddLayer adds an animation layer
func (p *ServerPacketDrawPlayer) AddLayer(id Serial, body uint16, layer Layer, hue Hue) {
	if hue != HueDefault {
		body |= 0x8000
	}

	(*p)[22]++
	putUInt32((*[]byte)(p), uint32(id))
	putUInt16((*[]byte)(p), body)
	putByte((*[]byte)(p), byte(layer))
	if hue != HueDefault {
		putUInt16((*[]byte)(p), uint16(hue))
	}
}

// Finish finished writing the packet
func (p *ServerPacketDrawMobile) Finish() {
	if (*p)[22] == 0 {
		fill((*[]byte)(p), 0, 5)
	} else {
		fill((*[]byte)(p), 0, 4)
	}
	backPatchLength((*[]byte)(p))
}

// NewServerPacketDrawMobile creates a new packet
func NewServerPacketDrawMobile(p []byte, id Serial, body, x, y uint16, z int8, dir Dir, hue Hue, status StatusFlag, noto Noto) ServerPacketDrawMobile {
	putByte(&p, 0x78)
	putUInt16(&p, 0)
	putUInt32(&p, uint32(id))
	putUInt16(&p, body)
	putUInt16(&p, x)
	putUInt16(&p, y)
	putByte(&p, byte(z))
	putByte(&p, byte(dir))
	putUInt16(&p, uint16(hue))
	putByte(&p, byte(status))
	putByte(&p, byte(noto))
	putUInt32(&p, 0)
	return ServerPacketDrawMobile(p)
}

// ServerPacketConnectToServer is used to tell the client what game server to
// connect to
type ServerPacketConnectToServer []byte

// Bytes returns the underlying byte slice
func (p ServerPacketConnectToServer) Bytes() []byte {
	return ([]byte)(p)
}

// NewServerPacketConnectToServer creates a new packet
func NewServerPacketConnectToServer(p []byte, addr net.IPAddr, port uint16) ServerPacketConnectToServer {
	ip := addr.IP.To4()
	putByte(&p, 0x8c)
	putByte(&p, ip[0])
	putByte(&p, ip[1])
	putByte(&p, ip[2])
	putByte(&p, ip[3])
	putUInt16(&p, port)
	putUInt32(&p, 0xbaadf00d)
	return ServerPacketConnectToServer(p)
}

// ServerPacketCharacterList is used to display the charachter list to the client
type ServerPacketCharacterList []byte

// Bytes returns the underlying byte slice
func (p ServerPacketCharacterList) Bytes() []byte {
	return ([]byte)(p)
}

// NewServerPacketCharacterList creates a new packet
func NewServerPacketCharacterList(p []byte) ServerPacketCharacterList {
	putByte(&p, 0xa9)
	putUInt16(&p, 0)
	putByte(&p, 0)
	return ServerPacketCharacterList(p)
}

// AddCharacter adds a character to the packet (max 7)
func (p *ServerPacketCharacterList) AddCharacter(name string) {
	(*p)[3]++
	putASCII((*[]byte)(p), name, 30)
	fill((*[]byte)(p), 0, 30)
}

// FinishCharacterList finishes the character list portion of the packet
func (p *ServerPacketCharacterList) FinishCharacterList() {
	binary.BigEndian.PutUint16((*p)[1:3], uint16(len(*p)))
	putByte((*[]byte)(p), 0)
}

// AddStartingLocation adds a starting location (need at least 1)
func (p *ServerPacketCharacterList) AddStartingLocation(city, area string) {
	ofs := binary.BigEndian.Uint16((*p)[1:3])
	idx := (*p)[ofs]
	(*p)[ofs]++
	putByte((*[]byte)(p), idx)
	putASCII((*[]byte)(p), city, 31)
	putASCII((*[]byte)(p), area, 31)
}

// Finish finished writing the packet with the given flags
func (p *ServerPacketCharacterList) Finish(flags FeatureFlag) {
	putUInt32((*[]byte)(p), uint32(flags))
	backPatchLength((*[]byte)(p))
}

// ServerPacketMoveAck is used to acknowledge a character move
type ServerPacketMoveAck []byte

// Bytes returns the underlying byte slice
func (p ServerPacketMoveAck) Bytes() []byte {
	return ([]byte)(p)
}

// NewServerPacketMoveAck creates a new packet
func NewServerPacketMoveAck(p []byte, key byte, noto Noto) ServerPacketMoveAck {
	putByte(&p, 0x22)
	putByte(&p, key)
	putByte(&p, byte(noto))
	return ServerPacketMoveAck(p)
}

// ServerPacketEnableClientFeatures is used just after game server login to
// enable certian subscription features in the client.
type ServerPacketEnableClientFeatures []byte

// Bytes returns the underlying byte slice
func (p ServerPacketEnableClientFeatures) Bytes() []byte {
	return ([]byte)(p)
}

// NewServerPacketEnableClientFeatures creates a new packet
func NewServerPacketEnableClientFeatures(p []byte, flags uint16) ServerPacketEnableClientFeatures {
	putByte(&p, 0xb9)
	putUInt16(&p, flags)
	return ServerPacketEnableClientFeatures(p)
}
