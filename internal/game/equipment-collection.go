package game

import (
	"log"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// EquipmentCollection is a collection of Items associated to equipment layers.
// The zero value of EquipmentCollection is valid for all operations.
type EquipmentCollection struct {
	// Collection of currently equipped items
	equipment map[uo.Layer]Wearable
}

// NewEquipmentCollection creates a new, empty EquipmentCollection and returns
// it.
func NewEquipmentCollection() *EquipmentCollection {
	return &EquipmentCollection{
		equipment: make(map[uo.Layer]Wearable),
	}
}

// NewEquipmentCollectionWith reads the collection references from the given
// slice of object IDs and rebuilds the pointers.
func NewEquipmentCollectionWith(ids []uo.Serial) *EquipmentCollection {
	c := NewEquipmentCollection()
	for _, id := range ids {
		o := world.Find(id)
		if o == nil {
			log.Printf("object %s does not exist", id.String())
			continue
		}
		w, ok := o.(Wearable)
		if !ok {
			log.Printf("object %s does not implement the wearable interface", id.String())
			continue
		}
		if _, duplicate := c.equipment[w.Layer()]; duplicate {
			log.Printf("object %s duplicate layer %d", id.String(), w.Layer())
			continue
		}
		c.equipment[w.Layer()] = w
	}
	return c
}

// Write writes the collection to the given tag file.
func (c *EquipmentCollection) Write(name string, f *util.TagFileWriter) {
	f.WriteObjectReferences(name, util.ValuesAsSerials(c.equipment))
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

// Unequip removes an item from the collection. Returns true if successful.
func (c *EquipmentCollection) Unequip(o Wearable) bool {
	if c.equipment == nil {
		c.equipment = make(map[uo.Layer]Wearable)
	}
	// Only remove if the item is what is equipped in that slot
	if equipped, ok := c.equipment[o.Layer()]; ok {
		if equipped == o {
			delete(c.equipment, o.Layer())
			return true
		}
	}
	return false
}

// Contains returns true if the equipment collection contains the item
func (c *EquipmentCollection) Contains(o Wearable) bool {
	if c.equipment == nil {
		return false
	}
	_, found := c.equipment[o.Layer()]
	return found
}
