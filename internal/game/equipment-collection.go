package game

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// EquipmentCollection is a collection of Items associated to equipment layers.
// The zero value of EquipmentCollection is valid for all operations.
type EquipmentCollection struct {
	// Collection of currently equipped items
	equipment map[uo.Layer]Wearable
}

// Write writes the collection to the given tag file.
func (c *EquipmentCollection) Write(name string, f *util.TagFileWriter) {
	f.WriteObjectReferences(name, util.ValuesAsSerials(c.equipment))
}

// Read reads the collection from the given tag file object and rebuilds the
// collection's pointers.
func (c *EquipmentCollection) Read(name string, f *util.TagFileObject) {
	c.equipment = make(map[uo.Layer]Wearable)
	for _, id := range f.GetObjectReferences(name) {
		o := world.Find(id)
		w, ok := o.(Wearable)
		if !ok {
			f.InjectError(fmt.Errorf("object %s does not implement the wearable interface", id.String()))
			continue
		}
		if _, duplicate := c.equipment[w.Layer()]; duplicate {
			f.InjectError(fmt.Errorf("object %s duplicate layer %d", id.String(), w.Layer()))
			continue
		}
		c.equipment[w.Layer()] = w
	}
}

// Map executes a function for every item in the collection.
func (c *EquipmentCollection) Map(fn func(Wearable) error) []error {
	var errs []error
	for _, w := range c.equipment {
		if err := fn(w); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Equip adds an item to the collection as long as the layer is not already
// taken.
func (c *EquipmentCollection) Equip(o Wearable) bool {
	if c.equipment == nil {
		c.equipment = make(map[uo.Layer]Wearable)
	}
	if _, duplicate := c.equipment[o.Layer()]; duplicate {
		return false
	}
	c.equipment[o.Layer()] = o
	return true
}
