package uo

import (
	"math/rand"
	"time"
)

// Serial types
type SerialType int

// Valid values for SerialManager
const (
	SerialTypeMobile  SerialType = 0
	SerialTypeItem    SerialType = 1
	SerialTypeUnbound SerialType = 2
)

// SerialManager manages a pool of unique serials.
type SerialManager struct {
	used map[Serial]struct{}
	rng  *rand.Rand
}

// NewSerialManager constructs a new SerialManager
func NewSerialManager() *SerialManager {
	return &SerialManager{
		used: make(map[Serial]struct{}),
		rng:  rand.New(rand.NewSource(time.Now().Unix())),
	}
}

// New creates a new, unique Serial appropriate for the given type
func (m *SerialManager) New(t SerialType) Serial {
	var n Serial
	for {
		switch t {
		case SerialTypeMobile:
			n = m.RandomMobileSerial()
		case SerialTypeItem:
			n = m.RandomItemSerial()
		case SerialTypeUnbound:
			n = m.RandomUnboundSerial()
		default:
			panic("unknown serial manager type")
		}
		if _, duplicate := m.used[n]; !duplicate {
			m.used[n] = struct{}{}
			break
		}
	}
	return n
}

// RandomMobileSerial returns a randomized non-unique Serial fit for a mobile
func (m *SerialManager) RandomMobileSerial() Serial {
	return Serial(m.rng.Int31n(int32(SerialLastMobile-SerialFirstMobile))) + SerialFirstMobile
}

// RandomItemSerial returns a randomized non-unique Serial fit for an item
func (m *SerialManager) RandomItemSerial() Serial {
	return Serial(m.rng.Int31n(int32(SerialLastItem-SerialFirstItem))) + SerialFirstItem
}

// RandomUnboundSerial returns a randomized non-unique Serial that is NOT FIT
// for transmission to the client. These serials should be used internally only.
func (m *SerialManager) RandomUnboundSerial() Serial {
	return Serial(m.rng.Int31())
}

// Forcefully adds the serial to the set
func (m *SerialManager) Add(s Serial) {
	m.used[s] = struct{}{}
}

// Contains returns true if the given Serial is in the set
func (m *SerialManager) Contains(s Serial) bool {
	_, used := m.used[s]
	return used
}

// Remove removes the serial from the set
func (m *SerialManager) Remove(s Serial) {
	delete(m.used, s)
}
