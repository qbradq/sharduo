package game

import (
	"log"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// Container is the interface all objects implement that can contain other
// other objects within an inventory.
type Container interface {
	Item
	// GumpGraphic returns the gump graphic of the item
	GumpGraphic() uo.Gump
	// RemoveObject removes an object from this object. This is called when
	// changing parent objects. This function should return false if the object
	// could not be removed.
	RemoveObject(Object) bool
	// AddObject adds an object to this object. This is called when changing
	// parent objects. This function should return false if the object could not
	// be added.
	AddObject(Object) bool
	// Contains returns true if the object is a direct child of this container,
	// or any child containers.
	Contains(Object) bool
	// ContentWeight returns the total weight of all contained items excluding
	// the weight of the container itself. Use game.Item.Weight to get the
	// weight of the container without the contents.
	ContentWeight() int
	// ItemCount returns the total number of items within the container.
	ItemCount() int
	// RecalculateContentWeight refreshes the content weight cache, and calls
	// itself recursively for child containers
	RecalculateContentWeight()
	// MapContents executes the function over every item in the container and
	// returns the accumulated non-nil errors.
	MapContents(fn func(Item) error) []error
	// RemoveObserver removes an observer from this container's list of
	// containers.
	RemoveObserver(o ContainerObserver)
}

// BaseContainer implements the base implementation of the Container interface.
type BaseContainer struct {
	// Contents of the container
	contents util.Slice[Item]
	// Gump image to show for the container background
	gump uo.Gump
	// Max weight this container can hold
	maxContainerWeight int
	// Max items this container can hold
	maxContainerItems int
	// Cache of the current weight of all the items in the container
	contentWeight int
	// All of the observers of the container
	observers util.Slice[ContainerObserver]
	// Bounds of the container
	bounds uo.Bounds
}

// Serialize implements the util.Serializeable interface.
func (c *BaseContainer) Serialize(f *util.TagFileWriter) {
	f.WriteHex("Gump", uint32(c.gump))
	f.WriteNumber("MaxContainerWeight", c.maxContainerWeight)
	f.WriteNumber("MaxContainerItems", c.maxContainerItems)
	f.WriteBounds("Bounds", c.bounds)
	f.WriteObjectReferences("Contents", util.ToSerials(c.contents))
}

// Deserialize implements the util.Serializeable interface.
func (c *BaseContainer) Deserialize(f *util.TagFileObject) {
	c.gump = uo.Gump(f.GetHex("Gump", uint32(0x003C)))
	c.maxContainerWeight = f.GetNumber("MaxContainerWeight", uo.DefaultMaxContainerWeight)
	c.maxContainerItems = f.GetNumber("MaxContainerItems", uo.DefaultMaxContainerItems)
	c.bounds = f.GetBounds("Bounds", uo.Bounds{X: 44, Y: 65, W: 142, H: 94})
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (c *BaseContainer) OnAfterDeserialize(t *util.TagFileObject) {
	for _, serial := range t.GetObjectReferences("Contents") {
		o := world.Find(serial)
		if o == nil {
			log.Printf("failed to link object 0x%X into container", serial)
		}
		item, ok := o.(Item)
		if !ok {
			log.Printf("failed to link object 0x%X into container, it is not an item", serial)
		}
		c.contents = c.contents.Append(item)
	}
	c.RecalculateContentWeight()
}

// RecalculateContentWeight refreshes BaseContainer.contentWeight, and calls
// itself recursively for child containers
func (c *BaseContainer) RecalculateContentWeight() {
	c.contentWeight = 0
	for _, item := range c.contents {
		if container, ok := item.(Container); ok {
			container.RecalculateContentWeight()
		}
		c.contentWeight += item.Weight()
	}
}

// GumpGraphic implements the Container interface.
func (c *BaseContainer) GumpGraphic() uo.Gump { return c.gump }

// RemoveObject implements the Container interface.
func (c *BaseContainer) RemoveObject(o Object) bool {
	// Only items go into containers
	item, ok := o.(Item)
	if !ok {
		return false
	}
	// This avoids a duplicate call to IndexOf
	oldLength := len(c.contents)
	c.contents = c.contents.Remove(item)
	if len(c.contents) == oldLength {
		return false
	}
	c.contentWeight -= item.Weight()
	return true
}

// AddObject implements the Container interface.
func (c *BaseContainer) AddObject(o Object) bool {
	// Only items go into containers
	item, ok := o.(Item)
	if !ok {
		return false
	}
	// Something is very wrong
	if c.contents.IndexOf(item) >= 0 {
		return false
	}
	// Container weight check
	if c.maxContainerWeight > 0 && c.ContentWeight()+item.Weight() > c.maxContainerWeight {
		// TODO Send cliloc message 1080016
		return false
	}
	// Max items check
	if c.maxContainerItems > 0 && len(c.contents)+1 > c.maxContainerItems {
		// TODO Send cliloc message 1080017
		return false
	}
	// Location bounding
	l := item.Location()
	if l.X == uo.RandomX {
		l.X = world.Random().Random(c.bounds.X, c.bounds.X+c.bounds.W-1)
	}
	if l.Y == uo.RandomY {
		l.Y = world.Random().Random(c.bounds.Y, c.bounds.Y+c.bounds.H-1)
	}
	if l.X < c.bounds.X {
		l.X = c.bounds.X
	}
	if l.X >= c.bounds.X+c.bounds.W {
		l.X = c.bounds.X + c.bounds.W - 1
	}
	if l.Y < c.bounds.Y {
		l.Y = c.bounds.Y
	}
	if l.Y >= c.bounds.Y+c.bounds.H {
		l.Y = c.bounds.Y + c.bounds.H - 1
	}
	item.SetLocation(l)
	// Add the item to our contents
	c.contents = c.contents.Append(item)
	c.contentWeight += item.Weight()
	return true
}

// Contains implements the Container interface.
func (c *BaseContainer) Contains(o Object) bool {
	// Only items go into containers
	otherItem, ok := o.(Item)
	if !ok {
		return false
	}
	for _, item := range c.contents {
		if item == otherItem {
			return true
		}
		if container, ok := item.(Container); ok {
			if container.Contains(o) {
				return true
			}
		}
	}
	return false
}

// ContentWeight implements the Container interface.
func (c *BaseContainer) ContentWeight() int { return c.contentWeight }

// ItemCount implements the Container interface.
func (c *BaseContainer) ItemCount() int { return len(c.contents) }

// MapContents implements the Container interface.
func (c *BaseContainer) MapContents(fn func(Item) error) []error {
	var ret []error
	for _, item := range c.contents {
		if err := fn(item); err != nil {
			ret = append(ret, err)
		}
	}
	return ret
}

// RemoveObserver implements the Container interface.
func (c *BaseContainer) RemoveObserver(o ContainerObserver) {
	c.observers = c.observers.Remove(o)
}
