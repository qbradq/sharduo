package marshal

import (
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

// Marshaler is an object that knows how to marshal itself into a binary format.
type Marshaler interface {
	// Serial returns the uo.Serial value identifying this object.
	Serial() uo.Serial
	// Marshal writes the binary representation of the object to the segment.
	Marshal(*TagFileSegment)
	// ObjectType returns the ObjectType associated with this struct.
	ObjectType() ObjectType
	// TemplateName returns the string describing what template the object was
	// constructed from.
	TemplateName() string
}

// Unmarshaler is an object that knows how to unmarshal itself from binary.
type Unmarshaler interface {
	// SetSerial sets the uo.Serial value identifying this object.
	SetSerial(uo.Serial)
	// Unmarshal reads the binary representation of the object from the tag
	// object.
	Unmarshal(*TagFileSegment)
	// Deserialize takes data from the template object and initializes the
	// object's data structures with it.
	Deserialize(*template.Template, bool)
}
