package game

import (
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &MountItem{} })
}

// MountItem is a special wearable on uo.LayerMount that represents a mount
// that a mobile is riding, if any.
type MountItem struct {
	BaseWearable
	// The mobile we are managing
	m Mobile
}

// TypeName implements the util.Serializeable interface.
func (i *MountItem) TypeName() string {
	return "MountItem"
}

// Serialize implements the util.Serializeable interface.
func (i *MountItem) Serialize(f *util.TagFileWriter) {
	i.BaseItem.Serialize(f)
}

// Deserialize implements the util.Serializeable interface.
func (i *MountItem) Deserialize(f *util.TagFileObject) {
	i.BaseItem.Deserialize(f)
}

// RemoveObject implements the Object interface
func (i *MountItem) RemoveObject(o Object) bool {
	if i.m == nil {
		return false
	}
	if i.m.Serial() != o.Serial() {
		return false
	}
	i.ForceRemoveObject(o)
	return true
}

// doAdd adds this mobile to the item forcefully if requested
func (i *MountItem) doAdd(o Object, force bool) bool {
	if o == nil {
		return force
	}
	o.SetParent(i)
	m, ok := o.(Mobile)
	if !ok {
		return force
	}
	if i.m != nil {
		if force {
			i.m = m
		}
		return force
	}
	i.m = m
	return true
}

// AddObject implements the Object interface
func (i *MountItem) AddObject(o Object) bool {
	return i.doAdd(o, false)
}

// ForceAddObject implements the Object interface. PLEASE NOTE that a call to
// BaseObject.ForceAddObject() will leak the object!
func (i *MountItem) ForceAddObject(o Object) {
	i.doAdd(o, true)
}

// ForceRemoveObject implements the Object interface. PLEASE NOTE that a call to
// BaseObject.ForceRemoveObject() will leak the object!
func (i *MountItem) ForceRemoveObject(o Object) {
	i.m = nil
}
