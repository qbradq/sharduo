package game

import "github.com/qbradq/sharduo/lib/util"

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
func (c *WearableContainer) DropObject(o Object, from Mobile) bool {
	// TODO Access calculations
	return world.Map().SetNewParent(o, c)
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
