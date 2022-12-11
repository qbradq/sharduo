package util

import "fmt"

// Registry is a managed map that only supports Add and Get operations. This is
// used in the factory pattern as well as a few other things.
type Registry[K comparable, V any] struct {
	values map[K]V
	name   string
}

// NewRegistry returns a new Registry ready for use.
func NewRegistry[K comparable, V any](name string) *Registry[K, V] {
	return &Registry[K, V]{
		values: make(map[K]V),
		name:   name,
	}
}

// Add adds a value to the registry
func (r *Registry[K, V]) Add(k K, v V) {
	if _, duplicate := r.values[k]; duplicate {
		panic(fmt.Sprintf("duplicate key in registry %s", r.name))
	}
	r.values[k] = v
}

// Get returns the value associated with key and has the same semantics as
// looking up a map key
func (r *Registry[K, V]) Get(k K) (V, bool) {
	v, ok := r.values[k]
	return v, ok
}

// GetName returns the debug name of the registry
func (r *Registry[K, V]) GetName() string {
	return r.name
}

// Contains returns true if the key is contained within the registry
func (r *Registry[K, V]) Contains(k K) bool {
	_, found := r.Get(k)
	return found
}

// Map executes fn on every object in the pool and returns a slice of all
// non-nil return values.
func (r *Registry[K, V]) Map(fn func(K, V) error) []error {
	var ret []error
	for k, v := range r.values {
		if err := fn(k, v); err != nil {
			ret = append(ret, err)
		}
	}
	return ret
}
