package uo

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
}

// NewSerialManager constructs a new SerialManager
func NewSerialManager() *SerialManager {
	return &SerialManager{
		used: make(map[Serial]struct{}),
	}
}

// New creates a new, unique Serial appropriate for the given type
func (m *SerialManager) New(t SerialType) Serial {
	var n Serial
	for {
		switch t {
		case SerialTypeMobile:
			n = RandomMobileSerial()
		case SerialTypeItem:
			n = RandomItemSerial()
		case SerialTypeUnbound:
			n = RandomUnboundSerial()
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
