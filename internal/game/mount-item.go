package game

import (
	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("MountItem", marshal.ObjectTypeMountItem, func() any { return &MountItem{} })
}

// MountItem is a special wearable on uo.LayerMount that represents a mount
// that a mobile is riding, if any.
type MountItem struct {
	BaseWearable
	// The mobile we are managing
	m Mobile
}

// ObjectType implements the Object interface.
func (i *MountItem) ObjectType() marshal.ObjectType {
	return marshal.ObjectTypeMountItem
}

// Marshal implements the marshal.Marshaler interface.
func (i *MountItem) Marshal(s *marshal.TagFileSegment) {
	i.BaseWearable.Marshal(s)
	i.m.Marshal(s)
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *MountItem) Unmarshal(s *marshal.TagFileSegment) {
	i.BaseWearable.Unmarshal(s)
	mum := s.Object()
	m, ok := mum.(Mobile)
	if !ok {
		panic("mount item's mount did not implement Mobile")
	}
	i.m = m
	i.SetBaseGraphicForBody(i.m.Body())
}

// SetBaseGraphicForBody sets the base graphic of the item correctly for the
// given mount body.
func (i *MountItem) SetBaseGraphicForBody(body uo.Body) {
	switch body {
	case 0xC8:
		i.SetBaseGraphic(0x3E9F)
	case 0xCC:
		i.SetBaseGraphic(0x3EA2)
	case 0xDC:
		i.SetBaseGraphic(0x3EA6)
	case 0xE2:
		i.SetBaseGraphic(0x3EA0)
	case 0xE4:
		i.SetBaseGraphic(0x3EA1)
	}
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

// Mount returns the current mount mobile associated to the item. Might be nil.
func (i *MountItem) Mount() Mobile { return i.m }
