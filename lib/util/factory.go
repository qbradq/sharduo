package util

import "fmt"

// Factory manages a Registry of constructors for creating new objects based
// on registration ID.
type Factory[K comparable, R any] struct {
	Registry[K, func() R]
}

// NewFactory returns a new Factory[K, R] ready for use
func NewFactory[K comparable, R any](name string) *Factory[K, R] {
	return &Factory[K, R]{
		*NewRegistry[K, func() R](name),
	}
}

// New creates a new object from the factory, or nil if no constructor was
// found
func (f *Factory[K, R]) New(id K) R {
	var defaultValue R
	if ctor, ok := f.Get(id); ok {
		return ctor()
	}
	return defaultValue
}

// Register registers a constructor with this registry
func (f *Factory[K, R]) Register(id K, ctor func() R) {
	if f.Contains(id) {
		panic(fmt.Sprintf("duplicate type constructor registered for %v", id))
	}
	f.Add(id, ctor)
}
