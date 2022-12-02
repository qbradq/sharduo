package uo

import (
	"encoding/binary"
	"strconv"
)

// A Serial is a 31-bit value with the following characteristics:
// The zero value is also the "invalid value" value
// No Serial will have a value greater than 2^31-1
type Serial uint32

// Pre-defined values of Serial
const (
	SerialZero            Serial = 0x00000000
	SerialMobileNil       Serial = 0x00000000
	SerialFirstMobile     Serial = 0x00000001
	SerialLastMobile      Serial = 0x3fffffff
	SerialItemNil         Serial = 0x40000000
	SerialFirstItem       Serial = 0x40000001
	SerialLastItem        Serial = 0x7fffffff
	SerialSelfMask        Serial = 0x7fffffff
	SerialMobileSelfNil   Serial = 0x80000000
	SerialFirstMobileSelf Serial = 0x80000001
	SerialLastMobileSelf  Serial = 0xcfffffff
	SerialSystem          Serial = 0xffffffff
)

// NewSerialFromData creates a new Serial from a []byte slice of at least length
// four.
func NewSerialFromData(in []byte) Serial {
	return Serial(binary.BigEndian.Uint32(in))
}

// Data creates a new byte slice of the serial for transmission on the wire.
func (s Serial) Data() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(s))
	return buf
}

// NewSerialFromString returns a new Serial parsed as a hex number.
func NewSerialFromString(in string) Serial {
	s, err := strconv.ParseUint(in, 0, 31)
	if err != nil {
		panic(err)
	}
	return Serial(s)
}

// String returns the string representation of the serial.
func (s Serial) String() string {
	return string(strconv.AppendInt(nil, int64(s), 16))
}

// IsMobile returns true if the serial refers to a mobile
func (s Serial) IsMobile() bool {
	return s <= SerialLastMobile && s >= SerialFirstMobile
}

// IsItem returns true if the serial refers to a item
func (s Serial) IsItem() bool {
	return s <= SerialLastItem && s >= SerialFirstItem
}

// IsSelf returns true if the serial is a valid client self-reference
func (s Serial) IsSelf() bool {
	return s <= SerialLastMobileSelf && s >= SerialFirstMobileSelf
}

// StripSelfFlag strips the MobileSelf flag out of the serial
func (s Serial) StripSelfFlag() Serial {
	if s.IsSelf() {
		return s & SerialSelfMask
	}
	return s
}

// IsNull returns true if the serial refers to a null mobile or item
func (s Serial) IsNil() bool {
	return s == SerialZero || s == SerialMobileNil || s == SerialItemNil || s == SerialMobileSelfNil
}
