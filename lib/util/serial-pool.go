package util

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/uo"
)

// SerialPool manages a pool of objects that implement the Serialer interface.
type SerialPool struct {
	objects map[uo.Serial]Serialer
	sm      *uo.SerialManager
	name    string
}

// NewSerialPool returns a new SerializeablePool object ready for use.
func NewSerialPool(name string, rng uo.RandomSource) *SerialPool {
	return &SerialPool{
		objects: make(map[uo.Serial]Serialer),
		sm:      uo.NewSerialManager(rng),
		name:    name,
	}
}

// Add adds the object to the pool, assigning it a unique serial.
func (p *SerialPool) Add(o Serialer, stype uo.SerialType) {
	o.SetSerial(p.sm.New(stype))
	p.sm.Add(o.GetSerial())
	p.objects[o.GetSerial()] = o
}

// Insert adds the object to the pool without overwriting its serial. This will
// panic on duplicate insertion.
func (p *SerialPool) Insert(o Serialer) {
	if p.sm.Contains(o.GetSerial()) {
		panic(fmt.Sprintf("duplicate insertion into %s:0x%08X", p.name, o.GetSerial()))
	}
	p.sm.Add(o.GetSerial())
	p.objects[o.GetSerial()] = o
}

// Remove removes the object from the pool, assigning it uo.SerialZero.
func (p *SerialPool) Remove(o Serialer) {
	p.sm.Remove(o.GetSerial())
	delete(p.objects, o.GetSerial())
	o.SetSerial(uo.SerialZero)
}

// Get returns the identified object or nil
func (p *SerialPool) Get(id uo.Serial) Serialer {
	if o, found := p.objects[id]; found {
		return o
	}
	return nil
}

// Map executes fn on every object in the pool and returns a slice of all
// non-nil return values.
func (p *SerialPool) Map(fn func(Serialer) error) []error {
	var ret []error
	for _, o := range p.objects {
		if err := fn(o); err != nil {
			ret = append(ret, err)
		}
	}
	return ret
}

// Length returns the number of objects in the pool.
func (p *SerialPool) Length() int {
	return len(p.objects)
}
