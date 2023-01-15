package marshal

import "fmt"

// Collection of object constructors
var ctors = make(map[ObjectType]func() interface{})

// Registers a constructor for an object.
func RegisterCtor(ot ObjectType, ctor func() interface{}) {
	if _, duplicate := ctors[ot]; duplicate {
		panic(fmt.Sprintf("duplicate marshal object ctor for %d", ot))
	}
	ctors[ot] = ctor
}

// Constructor returns the constructor for an object type.
func Constructor(ot ObjectType) func() interface{} {
	return ctors[ot]
}
