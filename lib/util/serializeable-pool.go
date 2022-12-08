package util

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/uo"
)

// SerializeablePool manages a pool of Serializeable objects
type SerializeablePool struct {
	objects map[uo.Serial]Serializeable
	sm      *uo.SerialManager
	name    string
}

// NewSerializeablePool returns a new SerializeablePool object ready for use.
func NewSerializeablePool(name string, rng uo.RandomSource) *SerializeablePool {
	return &SerializeablePool{
		objects: make(map[uo.Serial]Serializeable),
		sm:      uo.NewSerialManager(rng),
		name:    name,
	}
}

// Add adds the serializeable object to the pool, assigning it a unique serial.
func (p *SerializeablePool) Add(o Serializeable, stype uo.SerialType) {
	o.SetSerial(p.sm.New(stype))
	p.sm.Add(o.GetSerial())
	p.objects[o.GetSerial()] = o
}

// Insert adds the serializeable object to the pool without overwriting its
// serial. This will panic on duplicate insertion.
func (p *SerializeablePool) Insert(o Serializeable) {
	if p.sm.Contains(o.GetSerial()) {
		panic(fmt.Sprintf("duplicate insertion into %s:0x%08X", p.name, o.GetSerial()))
	}
	p.sm.Add(o.GetSerial())
	p.objects[o.GetSerial()] = o
}

// Remove removes the serializeable object from the pool, assigning it
// uo.SerialZero.
func (p *SerializeablePool) Remove(o Serializeable) {
	p.sm.Remove(o.GetSerial())
	delete(p.objects, o.GetSerial())
	o.SetSerial(uo.SerialZero)
}

// Get returns the identified object or nil
func (p *SerializeablePool) Get(id uo.Serial) Serializeable {
	if o, found := p.objects[id]; found {
		return o
	}
	return nil
}

// Map executes fn on every object in the pool and returns a slice of all
// non-nil return values.
func (p *SerializeablePool) Map(fn func(Serializeable) error) []error {
	var ret []error
	for _, o := range p.objects {
		if err := fn(o); err != nil {
			ret = append(ret, err)
		}
	}
	return ret
}

// Length returns the number of objects in the pool.
func (p *SerializeablePool) Length() int {
	return len(p.objects)
}
