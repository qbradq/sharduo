package game

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &ContainerItem{} })
}

// ContainerItem is an item with the properties of a container
type ContainerItem struct {
	BaseItem
	BaseContainer
}

// TypeName implements the util.Serializeable interface.
func (o *ContainerItem) TypeName() string {
	return "ContainerItem"
}

// Serialize implements the util.Serializeable interface.
func (s *ContainerItem) Serialize(f *util.TagFileWriter) {
	s.BaseItem.Serialize(f) // This calls BaseObject.Serialize for us
	s.BaseContainer.Serialize(f)
}

// Deserialize implements the util.Serializeable interface.
func (s *ContainerItem) Deserialize(f *util.TagFileObject) {
	s.BaseItem.Deserialize(f) // This calls BaseObject.Deserialize for us
	s.BaseContainer.Deserialize(f)
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (c *ContainerItem) OnAfterDeserialize(f *util.TagFileObject) {
	c.BaseItem.OnAfterDeserialize(f) // This calls BaseObject.OnAfterDeserialize for us
	c.BaseContainer.OnAfterDeserialize(f)
}

// SingleClick implements the Object interface
func (c *ContainerItem) SingleClick(from Mobile) {
	// Default action is to send the name as over-head text
	if from.NetState() != nil {
		str := fmt.Sprintf("%s\n%d/%d items, %d/%d stones", c.DisplayName(),
			c.ItemCount(), c.maxContainerItems,
			c.contentWeight, c.maxContainerWeight)
		from.NetState().Speech(c, str)
	}
}

// Doubleclick implements the object interface.
func (c *ContainerItem) DoubleClick(from Mobile) {
	// TODO access calculations
	if from.NetState() != nil {
		if c.observers.IndexOf(from) < 0 {
			c.observers = c.observers.Append(from)
		}
		from.NetState().OpenContainer(c)
	}
}

// DropObject implements the Object interface.
func (c *ContainerItem) DropObject(o Object, l uo.Location, from Mobile) bool {
	// TODO Access calculations
	if item, ok := o.(Item); ok {
		item.SetDropLocation(l)
		return world.Map().SetNewParent(o, c)
	}
	return false
}

// RemoveObject implements the Container interface.
func (c *ContainerItem) RemoveObject(o Object) bool {
	if !c.BaseContainer.RemoveObject(o) {
		return false
	}
	item, ok := o.(Item)
	if !ok {
		return true
	}
	for _, observer := range c.observers {
		observer.ContainerItemRemoved(c, item)
	}
	return true
}

// ForceAddObject implements the Object interface.
func (c *ContainerItem) ForceAddObject(o Object) {
	o.SetParent(c)
	c.BaseContainer.ForceAddObject(o)
}

// AddObject implements the Container interface.
func (c *ContainerItem) AddObject(o Object) bool {
	if !c.BaseContainer.AddObject(o) {
		return false
	}
	item, ok := o.(Item)
	if !ok {
		return true
	}
	for _, observer := range c.observers {
		observer.ContainerItemAdded(c, item)
	}
	return true
}

// Weight implements the Object interface.
func (c *ContainerItem) Weight() int {
	return c.BaseItem.Weight() + c.ContentWeight()
}
