package uo

import (
	"fmt"
)

// Serial is a 31-bit value with the following characteristics: The zero value
// is also the "invalid value" value No Serial will have a value greater than
// 2^31-1
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
	SerialTheVoid         Serial = 0xd0000000
	SerialTextGUMP        Serial = 0xd0000001
	SerialSystem          Serial = 0xffffffff
)

// SerialType is a classification code for serial values
type SerialType byte

// All valid values for SerialType
const (
	SerialTypeItem   SerialType = 0
	SerialTypeMobile SerialType = 1
)

// String returns the string representation of the serial.
func (s Serial) String() string {
	return fmt.Sprintf("0x%08X", uint32(s))
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

// RandomMobileSerial returns a randomized Serial fit for a mobile
func RandomMobileSerial(rng RandomSource) Serial {
	return Serial(rng.Random(int(SerialFirstMobile), int(SerialLastMobile)))
}

// RandomItemSerial returns a randomized Serial fit for an item
func RandomItemSerial(rng RandomSource) Serial {
	return Serial(rng.Random(int(SerialFirstItem), int(SerialLastItem)))
}
