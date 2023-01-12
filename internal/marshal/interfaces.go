package marshal

// Marshaler is an object that knows how to marshal itself into a binary format.
type Marshaler interface {
	// Marshal writes the binary representation of the object to the segment.
	Marshal(*TagFileSegment)
}

// Unmarshaler is an object that knows how to unmarshal itself from binary.
type Unmarshaler interface {
	// Unmarshal reads the binary representation of the object from the tag
	// object.
	Unmarshal(*TagObject)
	// AfterUnmarshal is called after Unmarshal.
	AfterUnmarshal(*TagObject)
}
