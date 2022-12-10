package util

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/uo"
)

// SerialPool manages a pool of objects that implement the Serialer interface.
type SerialPool[K Serialer] struct {
	objects map[uo.Serial]K
	sm      *uo.SerialManager
	name    string
}

// NewSerialPool returns a new SerializeablePool object ready for use.
func NewSerialPool[K Serialer](name string, rng uo.RandomSource) *SerialPool[K] {
	return &SerialPool[K]{
		objects: make(map[uo.Serial]K),
		sm:      uo.NewSerialManager(rng),
		name:    name,
	}
}

// Add adds the object to the pool, assigning it a unique serial.
func (p *SerialPool[K]) Add(o K, stype uo.SerialType) {
	o.SetSerial(p.sm.New(stype))
	p.sm.Add(o.Serial())
	p.objects[o.Serial()] = o
}

// Insert adds the object to the pool using the given serial. This function will
// panic on duplicate insertion.
func (p *SerialPool[K]) Insert(o K, s uo.Serial) {
	if p.sm.Contains(s) {
		panic(fmt.Sprintf("duplicate insertion into %s:0x%08X", p.name, s))
	}
	p.sm.Add(s)
	p.objects[s] = o
}

// Remove removes the object from the pool, assigning it uo.SerialZero.
func (p *SerialPool[K]) Remove(o K) {
	p.sm.Remove(o.Serial())
	delete(p.objects, o.Serial())
	o.SetSerial(uo.SerialZero)
}

// Get returns the identified object or the zero value
func (p *SerialPool[K]) Get(id uo.Serial) K {
	var zero K
	if o, found := p.objects[id]; found {
		return o
	}
	return zero
}

// Map executes fn on every object in the pool and returns a slice of all
// non-nil return values.
func (p *SerialPool[K]) Map(fn func(K) error) []error {
	var ret []error
	for _, o := range p.objects {
		if err := fn(o); err != nil {
			ret = append(ret, err)
		}
	}
	return ret
}
