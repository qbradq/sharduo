package game

import (
	"log"

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
	ms := uo.SerialZero
	if i.m != nil {
		ms = i.m.Serial()
	}
	i.BaseWearable.Marshal(s)
	s.PutTag(marshal.TagManagedObject, marshal.TagValueInt, uint32(ms))
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *MountItem) Unmarshal(s *marshal.TagFileSegment) *marshal.TagCollection {
	tags := i.BaseWearable.Unmarshal(s)
	ms := uo.Serial(tags.Int(marshal.TagManagedObject))
	if ms != 0 {
		o := world.Find(ms)
		if o == nil {
			log.Printf("warning: mount item %s references non-existent object %s", i.Serial().String(), ms.String())
			i.m = nil
		} else {
			m, ok := o.(Mobile)
			if !ok {
				log.Printf("warning: mount item %s references non-mobile object %s, it probably leaked", i.Serial().String(), ms.String())
				i.m = nil
			} else {
				i.m = m
			}
		}
	} else {
		i.m = nil
	}
	return tags
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
