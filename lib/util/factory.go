package util

import "fmt"

// Factory manages a Registry of constructors for creating new objects based
// on registration ID.
type Factory[K comparable, I any, R any] struct {
	Registry[K, func(I) R]
}

// NewFactory returns a new Factory[K, I] ready for use
func NewFactory[K comparable, I any, R any](name string) *Factory[K, I, R] {
	return &Factory[K, I, R]{
		*NewRegistry[K, func(I) R](name),
	}
}

// New creates a new object from the factory, or nil if no constructor was
// found
func (f *Factory[K, I, R]) New(id K, in I) R {
	var defaultValue R
	if ctor, ok := f.Get(id); ok {
		return ctor(in)
	}
	return defaultValue
}

// Register registers a constructor with this registry
func (f *Factory[K, I, R]) Register(id K, ctor func(I) R) {
	if f.Contains(id) {
		panic(fmt.Sprintf("duplicate type constructor registered for %v", id))
	}
	f.Add(id, ctor)
}
