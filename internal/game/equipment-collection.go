package game

import (
	"github.com/qbradq/sharduo/internal/util"
	"github.com/qbradq/sharduo/lib/uo"
)

// EquipmentCollection is a collection of Items associated to equipment layers.
// The zero value of EquipmentCollection is valid for all operations.
type EquipmentCollection struct {
	// Collection of currently equipped items
	equipment map[uo.Layer]Item
}

// Write writes the the collection to the given tag file.
func (c *EquipmentCollection) Write(name string, f *util.TagFileWriter) {
	var serials []uo.Serial
	for _, item := range c.equipment {
		serials = append(serials, item.GetSerial())
	}
	f.WriteSerialSlice(name, serials)
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
	if _, duplicate := c.equipment[o.GetLayer()]; duplicate {
		return false
	}
	c.equipment[o.GetLayer()] = o
	return true
}
