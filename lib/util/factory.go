package util

import (
	"fmt"
)

// Factory creates Packet implementations based on a data slice.
type Factory[K comparable, I any] struct {
	// Map of packet constructors
	ctors map[K]func(I) any
	// Debug name of this factory
	name string
}

func NewFactory[K comparable, I any](name string) *Factory[K, I] {
	return &Factory[K, I]{
		ctors: make(map[K]func(I) any),
		name:  name,
	}
}

// GetName returns the debug name of the factory
func (f *Factory[K, I]) GetName() string {
	return f.name
}

// Add adds a constructor function to the factory
func (f *Factory[K, I]) Add(id K, ctor func(I) any) {
	if _, duplicate := f.ctors[id]; duplicate {
		panic(fmt.Sprintf("in packet factory %s duplicate ctor %v", f.name, id))
	}
	f.ctors[id] = ctor
}

// New creates a new object from the factory, or nil if no constructor was
// found
func (f *Factory[K, I]) New(id K, in I) any {
	if ctor, ok := f.ctors[id]; ok {
		return ctor(in)
	}
	return nil
}
