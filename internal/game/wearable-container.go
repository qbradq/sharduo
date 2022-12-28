package game

import (
	"fmt"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &WearableContainer{} })
}

// WearableContainer is a wearable item with the properties of a container, such
// as inventory backpacks and the player's bank box.
type WearableContainer struct {
	BaseWearable
	BaseContainer
}

// TypeName implements the util.Serializeable interface.
func (o *WearableContainer) TypeName() string {
	return "WearableContainer"
}

// Serialize implements the util.Serializeable interface.
func (s *WearableContainer) Serialize(f *util.TagFileWriter) {
	s.BaseWearable.Serialize(f) // This calls BaseObject.Serialize for us
	s.BaseContainer.Serialize(f)
}

// Deserialize implements the util.Serializeable interface.
func (s *WearableContainer) Deserialize(f *util.TagFileObject) {
	s.BaseWearable.Deserialize(f) // This calls BaseItem.Deserialize for us
	s.BaseContainer.Deserialize(f)
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (s *WearableContainer) OnAfterDeserialize(f *util.TagFileObject) {
	s.BaseWearable.OnAfterDeserialize(f) // This calls BaseObject.OnAfterDeserialize for us
	s.BaseContainer.OnAfterDeserialize(f)
}

// SingleClick implements the Object interface
func (c *WearableContainer) SingleClick(from Mobile) {
	// Default action is to send the name as over-head text
	if from.NetState() != nil {
		str := fmt.Sprintf("%s\n%d/%d items, %d/%d stones", c.DisplayName(),
			c.ItemCount(), c.maxContainerItems,
			c.contentWeight, c.maxContainerWeight)
		from.NetState().Speech(c, str)
	}
}

// Doubleclick implements the object interface.
func (c *WearableContainer) DoubleClick(from Mobile) {
	// TODO access calculations
	if from.NetState() != nil {
		if c.observers.IndexOf(from) < 0 {
			c.observers = c.observers.Append(from)
		}
		from.NetState().OpenContainer(c)
	}
}

// DropObject implements the Object interface.
func (c *WearableContainer) DropObject(o Object, l uo.Location, from Mobile) bool {
	if item, ok := o.(Item); ok {
		item.SetDropLocation(l)
		return world.Map().SetNewParent(o, c)
	}
	return false
}

// ForceAddObject implements the Object interface.
func (c *WearableContainer) ForceAddObject(o Object) {
	o.SetParent(c)
	c.BaseContainer.ForceAddObject(o)
}

// RemoveObject implements the Container interface.
func (c *WearableContainer) RemoveObject(o Object) bool {
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

// AddObject implements the Container interface.
func (c *WearableContainer) AddObject(o Object) bool {
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
func (c *WearableContainer) Weight() int {
	return c.BaseItem.Weight() + c.ContentWeight()
}
