package util

// Factory manages a Registry of constructors and creating new objects based
// on registration ID.
type Factory[K comparable, I any] struct {
	Registery[K, func(I) any]
}

func NewFactory[K comparable, I any](name string) *Factory[K, I] {
	return &Factory[K, I]{
		*NewRegistry[K, func(I) any](name),
	}
}

// New creates a new object from the factory, or nil if no constructor was
// found
func (f *Factory[K, I]) New(id K, in I) any {
	if ctor, ok := f.Get(id); ok {
		return ctor(in)
	}
	return nil
}
