package util

import (
	"github.com/qbradq/sharduo/lib/uo"
)

// Serializeable is the interface all serializeable objects implement.
type Serializeable interface {
	// GetSerial returns the serial of the object
	GetSerial() uo.Serial
	// SetSerial sets the serial of the object
	SetSerial(uo.Serial)
	// GetTypeName returns the name of the object's type, which must be unique
	GetTypeName() string
	// GetSerialType returns the type of serial number used by the object
	GetSerialType() uo.SerialType
	// Writes the object to a tag file.
	Serialize(*TagFileWriter)
	// Deserializes the object a tag file.
	Deserialize(*TagFileObject)
}

// BaseSerializeable implements the most common case of the Serializeable
// interface. GetTypeName() and GetSerialType() are purposefully omitted to
// force includers of this base struct to register their own.
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

// Serialize implements the util.Serializeable interface.
func (s *BaseSerializeable) Serialize(f *TagFileWriter) {
	f.WriteHex("Serial", int(s.Serial))
}

// Deserialize implements the util.Serializeable interface.
func (s *BaseSerializeable) Deserialize(f *TagFileObject) {
	s.Serial = uo.Serial(f.GetNumber("Serial", int(uo.SerialSystem)))
}
