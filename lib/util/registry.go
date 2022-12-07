package util

// Registry is a managed map that only supports Add and Get operations. This is
// used in the factory pattern as well as a few other things.
type Registery[K comparable, V any] struct {
	values map[K]V
	name   string
}

// NewRegistry returns a new Registry ready for use.
func NewRegistry[K comparable, V any](name string) *Registery[K, V] {
	return &Registery[K, V]{
		values: make(map[K]V),
		name:   name,
	}
}

// Add adds a value to the registry
func (r *Registery[K, V]) Add(k K, v V) {
	r.values[k] = v
}

// Get returns the value associated with key and has the same semantics as
// looking up a map key
func (r *Registery[K, V]) Get(k K) (V, bool) {
	v, ok := r.values[k]
	return v, ok
}

// GetName returns the debug name of the registry
func (r *Registery[K, V]) GetName() string {
	return r.name
}
