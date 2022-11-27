package util

import (
	"encoding/json"
	"log"

	"github.com/qbradq/sharduo/lib/uo"
)

// Serializeable is the interface all serializeable objects implement.
type Serializeable interface {
	// GetSerial returns the serial of the object
	GetSerial() uo.Serial
	// SetSerial sets the serial of the object
	SetSerial(uo.Serial)
	// Returns a representation of the object as a byte stream
	Serialize() []byte
	// Deserializes the object from a byte stream
	Deserialize([]byte)
}

// BaseSerializeable implements the most common case of the Serializeable
// interface.
type BaseSerializeable struct {
	Serial uo.Serial
}

// GetSerial implements the Serializeable interface
func (s *BaseSerializeable) GetSerial() uo.Serial {
	return s.Serial
}

// SetSerial implements the Serializeable interface
func (s *BaseSerializeable) SetSerial(serial uo.Serial) {
	s.Serial = serial
}

// Serialize implements the Serializeable interface
func (s *BaseSerializeable) Serialize() []byte {
	d, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return d
}

// Deserialize implements the Serializeable interface.
func (s *BaseSerializeable) Deserialize(d []byte) {
	if err := json.Unmarshal(d, s); err != nil {
		log.Fatal(err)
		return
	}
}
