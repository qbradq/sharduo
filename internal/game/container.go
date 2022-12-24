package game

import (
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
func (s *ContainerItem) OnAfterDeserialize(f *util.TagFileObject) {
	s.BaseItem.OnAfterDeserialize(f) // This calls BaseObject.OnAfterDeserialize for us
	// BaseContainer has nothing to do OnAfterDeserialize
}

// DoubleClick implements the Object interface.
func (c *ContainerItem) DoubleClick(from Mobile) {
	// TODO Debug code
	if from.NetState() != nil {
		from.NetState().OpenContainer(c)
	}
}

// Container is the interface all objects implement that can contain other
// other objects within an inventory.
type Container interface {
	util.Serialer
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
}

// BaseContainer implements the base implementation of the Container interface.
type BaseContainer struct {
	contents           util.Slice[Item]
	gump               uo.Gump
	maxContainerWeight int
	maxContainerItems  int
}

// Serialize implements the util.Serializeable interface.
func (c *BaseContainer) Serialize(f *util.TagFileWriter) {
	f.WriteHex("Gump", uint32(c.gump))
	f.WriteNumber("MaxContainerWeight", c.maxContainerWeight)
	f.WriteNumber("MaxContainerItems", c.maxContainerItems)
}

// Deserialize implements the util.Serializeable interface.
func (c *BaseContainer) Deserialize(f *util.TagFileObject) {
	c.gump = uo.Gump(f.GetHex("Gump", uint32(uo.GumpNone)))
	c.maxContainerWeight = f.GetNumber("MaxContainerWeight", uo.DefaultMaxContainerWeight)
	c.maxContainerItems = f.GetNumber("MaxContainerItems", uo.DefaultMaxContainerItems)
}

// GumpGraphic implements the Container interface.
func (c *BaseContainer) GumpGraphic() uo.Gump { return c.gump }

// RemoveObject implements the Container interface.
func (c *BaseContainer) RemoveObject(o Object) bool {
	item, ok := o.(Item)
	if !ok {
		return false
	}
	// This avoids a duplicate call to IndexOf
	oldLength := len(c.contents)
	c.contents = c.contents.Remove(item)
	return len(c.contents) != oldLength
}

// AddObject implements the Container interface.
func (c *BaseContainer) AddObject(o Object) bool {
	item, ok := o.(Item)
	if !ok {
		return false
	}
	if c.contents.IndexOf(item) >= 0 {
		return false
	}
	c.contents = c.contents.Append(item)
	return true
}

// Contains implements the Container interface.
func (c *BaseContainer) Contains(o Object) bool {
	for _, item := range c.contents {
		if item == o {
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
