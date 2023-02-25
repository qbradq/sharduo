package marshal

import "github.com/qbradq/sharduo/lib/uo"

// Serialer is an object that can identify itself with a uo.Serial value.
type Serialer interface {
	// Serial returns the uo.Serial value identifying this object.
	Serial() uo.Serial
}

// Marshaler is an object that knows how to marshal itself into a binary format.
type Marshaler interface {
	// Marshal writes the binary representation of the object to the segment.
	Marshal(*TagFileSegment)
}

// Unmarshaler is an object that knows how to unmarshal itself from binary.
type Unmarshaler interface {
	// Unmarshal reads the binary representation of the object from the tag
	// object.
	Unmarshal(*TagFileSegment) *TagCollection
	// AfterUnmarshal is called after Unmarshal.
	AfterUnmarshal(*TagCollection)
}
