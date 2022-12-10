package game

import (
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// EquipmentCollection is a collection of Items associated to equipment layers.
// The zero value of EquipmentCollection is valid for all operations.
type EquipmentCollection struct {
	// Collection of currently equipped items
	equipment map[uo.Layer]Item
}

// Write writes the the collection to the given tag file.
func (c *EquipmentCollection) Write(name string, f *util.TagFileWriter) {
	f.WriteObjectReferences(name, util.ValuesAsSerials(c.equipment))
}

// Map executes a function for every item in the collection.
func (c *EquipmentCollection) Map(fn func(Item) error) []error {
	var errs []error
	for _, item := range c.equipment {
		if err := fn(item); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Equip adds an item to the collection as long as the layer is not already
// taken.
func (c *EquipmentCollection) Equip(o Item) bool {
	if c.equipment == nil {
		c.equipment = make(map[uo.Layer]Item)
	}
	if _, duplicate := c.equipment[o.Layer()]; duplicate {
		return false
	}
	c.equipment[o.Layer()] = o
	return true
}
