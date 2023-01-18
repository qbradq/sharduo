package util

import (
	"github.com/qbradq/sharduo/lib/uo"
)

// Serialer is the interface all objects implement that can be identified by a
// uo.Serial value.
type Serialer interface {
	// GetSerial returns the serial of the object
	Serial() uo.Serial
	// SetSerial sets the serial of the object
	SetSerial(uo.Serial)
}

// BaseSerialer implements the most command use case of the Serialer interface.
type BaseSerialer struct {
	// Serial of the object
	serial uo.Serial
}

// Serial implements the Serializeable interface
func (s *BaseSerialer) Serial() uo.Serial {
	if s == nil {
		return uo.SerialZero
	}
	return s.serial
}

// SetSerial implements the Serializeable interface
func (s *BaseSerialer) SetSerial(serial uo.Serial) {
	if s != nil {
		s.serial = serial
	}
}

// Serializeable is the interface all serializeable objects implement.
type Serializeable interface {
	Serialer
	// SerialType returns the type of serial number used by the object
	SerialType() uo.SerialType
	// Deserializes the object from a tag file object. DO NOT CREATE new game
	// objects during deserialization!
	Deserialize(*TagFileObject)
	// Called on all objects after Deserialize has been called on all objects.
	// It is safe to create game objects in this function.
	OnAfterDeserialize(*TagFileObject)
}

// BaseSerializeable implements the most common case of the Serializeable
// interface. GetTypeName() and GetSerialType() are purposefully omitted to
// force includers of this base struct to register their own.
type BaseSerializeable struct {
	BaseSerialer
}

// Deserialize implements the util.Serializeable interface.
func (s *BaseSerializeable) Deserialize(f *TagFileObject) {
	s.serial = uo.Serial(f.GetHex("Serial", uint32(uo.SerialSystem)))
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (s *BaseSerializeable) OnAfterDeserialize(f *TagFileObject) {
	// Do nothing
}
