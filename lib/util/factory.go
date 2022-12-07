package util

// Constructor represents a serializeable constructor function that returns
// concrete implementations of the interface.
type Constructor func() Serializeable

// Factory manages a Registry of constructors and creating new objects based
// on registration ID.
type Factory[K comparable, I Serializeable] struct {
	Registery[K, func(I) Serializeable]
}

// NewFactory returns a new Factory[K, I] ready for use
func NewFactory[K comparable, I any](name string) *Factory[K, I] {
	return &Factory[K, I]{
		*NewRegistry[K, func(I) any](name),
	}
}

// New creates a new object from the factory, or nil if no constructor was
// found
func (f *Factory[K, I]) New(id K, in I) Serializeable {
	if ctor, ok := f.Get(id); ok {
		return ctor(in)
	}
	return nil
}

// Register registers a constructor with this registry
func (f *Factory[K, I]) Register(id K, ctor func(I) Serializeable) {
	s := ctor()
	if s == nil {
		panic("nil object returned from constructor", id)
	}
	name := s.GetTypeName()
	if f.Contains(name) {
		panic("duplicate type constructor registered for", name))
	}
	f.Add(id, ctor)
}
