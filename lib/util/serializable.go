package util

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/uo"
)

// Ctor represents a serializeable constructor function that returns concrete
// implementations of the interface.
type Ctor func() Serializeable

// RegisterCtor registers a constructor function for a serializeable object.
func RegisterCtor(ctor Ctor) {
	s := ctor()
	if s == nil {
		panic("nil object returned from constructor")
	}
	name := s.GetTypeName()
	if _, duplicate := ctors[name]; duplicate {
		panic(fmt.Sprintf("duplicate type constructor registered for %s", name))
	}
	ctors[name] = ctor
}

// NewFromCtor creates a new Serializeable implementation by type name, or nil
// if the named implementation cannot be found.
func NewFromCtor(name string) Serializeable {
	if ctor, ok := ctors[name]; ok {
		return ctor()
	}
	return nil
}

// ctors is the map of all ctor functions to type names.
var ctors map[string]Ctor = make(map[string]Ctor)

// Serializeable is the interface all serializeable objects implement.
type Serializeable interface {
	// GetSerial returns the serial of the object
	GetSerial() uo.Serial
	// SetSerial sets the serial of the object
	SetSerial(uo.Serial)
	// GetTypeName returns the name of the object's type, which must be unique.
	GetTypeName() string
	// Writes the object to a tag file.
	Serialize(*TagFileWriter)
	// Deserializes the object a tag file.
	Deserialize(*TagFileObject)
}

// BaseSerializeable implements the most common case of the Serializeable
// interface. GetTypeName() is purposefully omitted to force includers of this
// base struct to register their own name. BaseSerializeable implements
// comparable.
type BaseSerializeable struct {
	Serial uo.Serial
}

// compare implements the comparable interface
func (s *BaseSerializeable) compare(other Comparable) int {
	otherS, ok := other.(Serializeable)
	if !ok {
		return 0
	}
	if s.GetSerial() < otherS.GetSerial() {
		return -1
	}
	if s.GetSerial() == otherS.GetSerial() {
		return 0
	}
	return 1
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
