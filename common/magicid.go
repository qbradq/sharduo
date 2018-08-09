package common

import "math/rand"

// A Serial is a 31-bit value with the following characteristics:
// The zero value is also the "invalid value" value
// No Serial will have a value greater than 2^31-1
// A Serial can always be cast to a uint32 without data loss
type Serial int32

// A MagicIDPool produces MagicID values with the following characteristics:
// Values issued are always valid
// Values issued are always unique within the pool
type MagicIDPool struct {
	issued map[Serial]interface{}
}

// NewMagicIDPool creates a new MagicIDPool object
func NewMagicIDPool() *MagicIDPool {
	return &MagicIDPool{
		issued: make(map[Serial]interface{}),
	}
}

// Get creates a new MagicID
func (m *MagicIDPool) Get() Serial {
	for {
		id := Serial(rand.Int31())
		if _, ok := m.issued[id]; ok == false {
			m.issued[id] = struct{}{}
			return id
		}
	}
}

// Return tells the pool that a MagicID is no longer in use
func (m *MagicIDPool) Return(id Serial) {
	delete(m.issued, id)
}

// Reserve asks the pool to reserve a specific value in the pool
func (m *MagicIDPool) Reserve(id Serial) bool {
	if _, ok := m.issued[id]; ok == false {
		m.issued[id] = struct{}{}
		return true
	}
	return false
}
