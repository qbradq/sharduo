package uo

import (
	"strings"
)

// ClientPacket is the interface all packets implement
type ClientPacket interface {
	// Command returns the command byte of the packet
	Command() byte
}

func getASCII(buf []byte, start, length int) string {
	return strings.Trim(string(buf[start:start+length]), "\000")
}

// ClientPacketInvalid is used to indicate that the data stream from the client
// contained an invalid packet command header. This typically indicates that
// there is an error in the packet information tables or that there is a version
// missmatch between the client and this library.
type ClientPacketInvalid []byte

// Command returns the command byte of the packet
func (p ClientPacketInvalid) Command() byte {
	return p[0]
}

// ClientPacketNotSupported is used to indicate that this is a valid client
// packet, however no decoder yet exists for it
type ClientPacketNotSupported []byte

// Command returns the command byte of the packet
func (p ClientPacketNotSupported) Command() byte {
	return p[0]
}

// ClientPacketCharacterLogin is sent from the character selection page
type ClientPacketCharacterLogin []byte

// Command returns the command byte of the packet
func (p ClientPacketCharacterLogin) Command() byte {
	return p[0]
}

// CharacterSlot returns the character slot chosen
func (p ClientPacketCharacterLogin) CharacterSlot() uint {
	return uint(p[68])
}

func x5d(in []byte) ClientPacket {
	return ClientPacketCharacterLogin(in)
}

// ClientPacketAccountLogin is sent from the client's login page
type ClientPacketAccountLogin []byte

// Command returns the command byte of the packet
func (p ClientPacketAccountLogin) Command() byte {
	return p[0]
}

// RequiresLogin returns true if this packet should only be sent after
// a successful packet 0x91 (ClientPacketGameServerLogin)
func (p ClientPacketAccountLogin) RequiresLogin() bool {
	return false
}

// Username returns the username
func (p ClientPacketAccountLogin) Username() string {
	return getASCII(p, 1, 30)
}

// Password returns the plain-text password
func (p ClientPacketAccountLogin) Password() string {
	return getASCII(p, 31, 30)
}

// NextLoginKey returns the next login key from uo.cfg
func (p ClientPacketAccountLogin) NextLoginKey() byte {
	return p[61]
}

func x80(in []byte) ClientPacket {
	return ClientPacketAccountLogin(in)
}

// ClientPacketGameServerLogin is sent as the first packet to the game server
// socket.
type ClientPacketGameServerLogin []byte

// Command returns the command byte of the packet
func (p ClientPacketGameServerLogin) Command() byte {
	return p[0]
}

// RequiresLogin returns true if this packet should only be sent after
// a successful packet 0x91 (ClientPacketGameServerLogin)
func (p ClientPacketGameServerLogin) RequiresLogin() bool {
	return false
}

// Username returns the username
func (p ClientPacketGameServerLogin) Username() string {
	return getASCII(p, 5, 30)
}

// Password returns the plain-text password
func (p ClientPacketGameServerLogin) Password() string {
	return getASCII(p, 35, 30)
}

func x91(in []byte) ClientPacket {
	return ClientPacketGameServerLogin(in)
}

// ClientPacketSelectServer is sent when the client selects a server in the
// shard list.
type ClientPacketSelectServer []byte

// Command returns the command byte of the packet
func (p ClientPacketSelectServer) Command() byte {
	return p[0]
}

// ShardSelected returns the index of the selected shard
func (p ClientPacketSelectServer) ShardSelected() uint {
	return uint(p[2])
}

func xA0(in []byte) ClientPacket {
	return ClientPacketSelectServer(in)
}
