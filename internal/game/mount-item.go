package game

import (
	"log"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &MountItem{} })
	marshal.RegisterCtor(marshal.ObjectTypeMountItem, func() interface{} { return &MountItem{} })
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

// ObjectType implements the Object interface.
func (i *MountItem) ObjectType() marshal.ObjectType {
	return marshal.ObjectTypeMountItem
}

// Serialize implements the util.Serializeable interface.
func (i *MountItem) Serialize(f *util.TagFileWriter) {
	i.BaseWearable.Serialize(f)
	f.WriteHex("Mount", uint32(i.m.Serial()))
}

// Marshal implements the marshal.Marshaler interface.
func (i *MountItem) Marshal(s *marshal.TagFileSegment) {
	i.BaseWearable.Marshal(s)
	ms := uo.SerialZero
	if i.m != nil {
		ms = i.m.Serial()
	}
	i.BaseWearable.Marshal(s)
	s.PutTag(marshal.TagManagedObject, marshal.TagValueInt, uint32(ms))
}

// Deserialize implements the util.Serializeable interface.
func (i *MountItem) Deserialize(f *util.TagFileObject) {
	i.BaseWearable.Deserialize(f)
	ms := uo.Serial(f.GetHex("Mount", uint32(uo.SerialSystem)))
	if ms != uo.SerialSystem {
		o := world.Find(ms)
		if o == nil {
			log.Printf("error: mount item %s referenced non-existent mobile %s", i.Serial().String(), ms.String())
		} else if m, ok := o.(Mobile); ok {
			i.m = m
		}
	}
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *MountItem) Unmarshal(to *marshal.TagObject) {
	i.BaseWearable.Unmarshal(to)
	ms := uo.Serial(to.Tags.Int(marshal.TagManagedObject))
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
