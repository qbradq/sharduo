package uo

import (
	"encoding/binary"
	"math/rand"
	"time"
)

var serialRng = rand.New(rand.NewSource(time.Now().Unix()))

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

// RandomMobileSerial returns a randomized non-unique Serial fit for a mobile
func RandomMobileSerial() Serial {
	return Serial(serialRng.Int31n(int32(SerialLastMobile-SerialFirstMobile))) + SerialFirstMobile
}

// RandomItemSerial returns a randomized non-unique Serial fit for an item
func RandomItemSerial() Serial {
	return Serial(serialRng.Int31n(int32(SerialLastItem-SerialFirstItem))) + SerialFirstItem
}

// RandomUnboundSerial returns a randomized non-unique Serial that is NOT FIT
// for transmission to the client. These serials should be used internally only.
func RandomUnboundSerial() Serial {
	return Serial(serialRng.Int31())
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
