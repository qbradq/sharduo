package game

import (
	"fmt"
	"log"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseContainer{} })
}

// Container is the interface all objects implement that can contain other
// other objects within an inventory.
type Container interface {
	Item
	// GumpGraphic returns the gump graphic of the item
	GumpGraphic() uo.Gump
	// Contains returns true if the object is a direct child of this container,
	// or any child containers.
	Contains(Object) bool
	// ContentWeight returns the total weight of all contained items excluding
	// the weight of the container itself. Use game.Item.Weight to get the
	// weight of the container without the contents.
	ContentWeight() float32
	// ItemCount returns the total number of items within the container.
	ItemCount() int
	// MapContents executes the function over every item in the container and
	// returns the accumulated non-nil errors.
	MapContents(fn func(Item) error) []error
	// RemoveObserver removes an observer from this container's list of
	// containers.
	RemoveObserver(o ContainerObserver)
	// AdjustWeightAndCount adds to the cached total weight and item count of
	// the container.
	AdjustWeightAndCount(float32, int)
}

// BaseContainer implements the base implementation of the Container interface.
type BaseContainer struct {
	BaseItem
	// Contents of the container
	contents util.Slice[Item]
	// Gump image to show for the container background
	gump uo.Gump
	// Max weight this container can hold
	maxContainerWeight float32
	// Max items this container can hold
	maxContainerItems int
	// Cache of the current weight of all the items in the container
	contentWeight float32
	// Cache of the number of items in the container and all sub containers
	contentItems int
	// All of the observers of the container
	observers util.Slice[ContainerObserver]
	// Bounds of the container
	bounds uo.Bounds
}

// TypeName implements the util.Serializeable interface.
func (o *BaseContainer) TypeName() string {
	return "BaseContainer"
}

// Serialize implements the util.Serializeable interface.
func (c *BaseContainer) Serialize(f *util.TagFileWriter) {
	c.BaseItem.Serialize(f)
	f.WriteHex("Gump", uint32(c.gump))
	f.WriteFloat("MaxContainerWeight", c.maxContainerWeight)
	f.WriteNumber("MaxContainerItems", c.maxContainerItems)
	f.WriteBounds("Bounds", c.bounds)
	f.WriteObjectReferences("Contents", util.ToSerials(c.contents))
}

// Deserialize implements the util.Serializeable interface.
func (c *BaseContainer) Deserialize(f *util.TagFileObject) {
	c.BaseItem.Deserialize(f)
	c.gump = uo.Gump(f.GetHex("Gump", uint32(0x003C)))
	c.maxContainerWeight = f.GetFloat("MaxContainerWeight", float32(uo.DefaultMaxContainerWeight))
	c.maxContainerItems = f.GetNumber("MaxContainerItems", uo.DefaultMaxContainerItems)
	c.bounds = f.GetBounds("Bounds", uo.Bounds{X: 44, Y: 65, W: 142, H: 94})
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (c *BaseContainer) OnAfterDeserialize(t *util.TagFileObject) {
	c.BaseItem.OnAfterDeserialize(t)
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
}

// RecalculateStats implements the Object interface
func (c *BaseContainer) RecalculateStats() {
	c.contentWeight = 0
	c.contentItems = len(c.contents)
	for _, item := range c.contents {
		c.contentWeight += item.Weight()
		if container, ok := item.(Container); ok {
			container.RecalculateStats()
			c.contentItems += container.ItemCount()
			c.contentWeight += container.ContentWeight()
		}
	}
}

// GumpGraphic implements the Container interface.
func (c *BaseContainer) GumpGraphic() uo.Gump { return c.gump }

// SingleClick implements the Object interface
func (c *BaseContainer) SingleClick(from Mobile) {
	// Default action is to send the name as over-head text
	if from.NetState() != nil {
		// TODO send cliloc 1050044
		str := fmt.Sprintf("%s\n%d/%d items, %d/%d stones", c.DisplayName(),
			c.ItemCount(), c.maxContainerItems,
			int(c.contentWeight), int(c.maxContainerWeight))
		from.NetState().Speech(c, str)
	}
}

// Doubleclick implements the object interface.
func (c *BaseContainer) DoubleClick(from Mobile) {
	// TODO access calculations
	if from.NetState() != nil {
		if c.observers.IndexOf(from) < 0 {
			c.observers = c.observers.Append(from)
		}
		from.NetState().OpenContainer(c)
	}
}

// DropObject implements the Object interface.
func (c *BaseContainer) DropObject(o Object, l uo.Location, from Mobile) bool {
	// TODO Access calculations
	if item, ok := o.(Item); ok {
		item.SetDropLocation(l)
		return world.Map().SetNewParent(o, c)
	}
	return false
}

// doRemove removes an object forcefully if requested
func (c *BaseContainer) doRemove(o Object, force bool) bool {
	// Only items go into containers
	item, ok := o.(Item)
	if !ok {
		return force
	}
	// This avoids a duplicate call to IndexOf
	oldLength := len(c.contents)
	c.contents = c.contents.Remove(item)
	if len(c.contents) == oldLength {
		return force
	}
	itemsRemoved := 1
	if container, ok := o.(Container); ok {
		itemsRemoved += container.ItemCount()
	}
	c.AdjustWeightAndCount(-item.Weight(), -itemsRemoved)
	// Broadcast the item removal
	for _, observer := range c.observers {
		observer.ContainerItemRemoved(c, item)
	}
	// Gold calculations
	if mobile, ok := c.RootParent().(Mobile); ok && item.TemplateName() == "GoldCoin" {
		mobile.AdjustGold(-item.Amount())
	}
	return true
}

// RemoveObject implements the Object interface.
func (c *BaseContainer) RemoveObject(o Object) bool {
	return c.doRemove(o, false)
}

// ForceRemoveObject implements the Object interface.
func (c *BaseContainer) ForceRemoveObject(o Object) {
	c.doRemove(o, true)
}

// AddObject implements the Object interface.
func (c *BaseContainer) AddObject(o Object) bool {
	// Only items go into containers
	item, ok := o.(Item)
	if !ok {
		return false
	}
	// Something is very wrong
	if c.contents.Contains(item) {
		item.SetDropLocation(o.Location())
		return false
	}
	addedItems := 1
	addedWeight := item.Weight()
	if container, ok := item.(Container); ok {
		addedItems += container.ItemCount()
	}
	// Container weight check
	if c.maxContainerWeight > 0 && c.contentWeight+addedWeight > c.maxContainerWeight {
		// TODO Send cliloc message 1080016
		item.SetDropLocation(o.Location())
		return false
	}
	// Max items check
	if c.maxContainerItems > 0 && c.contentItems+addedItems > c.maxContainerItems {
		// TODO Send cliloc message 1080017
		item.SetDropLocation(o.Location())
		return false
	}
	c.ForceAddObject(o)
	return true
}

// ForceAddObject implements the Container interface.
func (c *BaseContainer) ForceAddObject(o Object) {
	if o == nil {
		return
	}
	o.SetParent(c)
	item, ok := o.(Item)
	if !ok {
		return
	}
	if c.contents.Contains(item) {
		return
	}
	// Determine if we should try to auto-stack the item
	l := item.DropLocation()
	if item.Stackable() && l.X == uo.RandomX && l.Y == uo.RandomY {
		for _, i := range c.contents {
			if i.CanCombineWith(item) && i.Combine(item) {
				return
			}
		}
		// Never successfully combined with a stack, continue to normal
		// placement.
	}
	addedItems := 1
	addedWeight := item.Weight()
	if container, ok := item.(Container); ok {
		addedItems += container.ItemCount()
	}
	// Location bounding
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
	// Add the item to our contents
	item.SetLocation(l)
	c.contents = c.contents.Append(item)
	c.AdjustWeightAndCount(addedWeight, addedItems)
	// Let all the observers know about the new item
	for _, observer := range c.observers {
		observer.ContainerItemAdded(c, item)
	}
	// Gold calculations
	if mobile, ok := c.RootParent().(Mobile); ok && item.TemplateName() == "GoldCoin" {
		mobile.AdjustGold(item.Amount())
	}
}

// Contains implements the Container interface.
func (c *BaseContainer) Contains(o Object) bool {
	// Only items go into containers
	otherItem, ok := o.(Item)
	if !ok {
		return false
	}
	for _, item := range c.contents {
		if item.Serial() == otherItem.Serial() {
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
func (c *BaseContainer) ContentWeight() float32 { return c.contentWeight }

// ItemCount implements the Container interface.
func (c *BaseContainer) ItemCount() int { return c.contentItems }

// Weight implements the Object interface.
func (c *BaseContainer) Weight() float32 {
	return c.BaseItem.Weight() + c.ContentWeight()
}

// AdjustWeightAndCount implements the Container interface.
func (c *BaseContainer) AdjustWeightAndCount(w float32, n int) {
	c.contentWeight += w
	c.contentItems += n
	if container, ok := c.parent.(Container); ok {
		// We are a sub-container, propagate the adjustment up
		container.AdjustWeightAndCount(w, n)
	} else if mobile, ok := c.parent.(Mobile); ok {
		if mobile.IsItemOnCursor() && mobile.ItemInCursor().Serial() == c.Serial() {
			// We are being held by a mobile's cursor, don't need to do anything
			return
		}
		// We are a mobile's backpack, send the weight adjustment up
		mobile.AdjustWeight(w)
	}
}

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
