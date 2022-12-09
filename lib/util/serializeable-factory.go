package util

import "fmt"

// SerializeableFactory is a special-case Factory for implementations of
// Serializeable
type SerializeableFactory struct {
	Factory[string, any, Serializeable]
}

// NewSerializeableFactory returns a new Serializeable factory ready for use
func NewSerializeableFactory(name string) *SerializeableFactory {
	return &SerializeableFactory{
		*NewFactory[string, any, Serializeable](name),
	}
}

// RegisterCtor registers a constructor and infers the type name
func (f *SerializeableFactory) RegisterCtor(ctor func(any) Serializeable) {
	var d any
	o := ctor(d)
	if o == nil {
		panic(fmt.Sprintf("serializeable factory %s got nil result from ctor", f.name))
	}
	f.Register(o.TypeName(), ctor)
}
