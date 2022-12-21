package game

import (
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

var ObjectFactory = util.NewSerializeableFactory("objects")

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseObject{} })
}

// Object is the interface every object in the game implements
type Object interface {
	util.Serializeable
	// Parent returns a pointer to the parent object of this object, or nil
	// if the object is attached directly to the world
	Parent() Object
	// SetParent sets the parent pointer. Use nil to represent the world.
	SetParent(Object)
	// RemoveObject removes an object from this object. This is called when
	// changing parent objects. This function should return false if the object
	// could not be removed.
	RemoveObject(Object) bool
	// AddObject adds an object to this object. This is called when changing
	// parent objects. This function should return false if the object could not
	// be added.
	AddObject(Object) bool
	// Location returns the current location of the object
	Location() uo.Location
	// SetLocation sets the absolute location of the object without regard to
	// the map.
	SetLocation(uo.Location)
	// Hue returns the hue of the item
	Hue() uo.Hue
	// DisplayName returns the name of the object with any articles attached
	DisplayName() string
	// Facing returns the direction the object is currently facing. 8-way for
	// mobiles, 2-way for most items, and 4-way for a few items.
	Facing() uo.Direction
	// SetFacing sets the direction the object is currently facing.
	SetFacing(uo.Direction)
}

// BaseObject is the base of all game objects and implements the Object
// interface
type BaseObject struct {
	util.BaseSerializeable
	// Parent object
	parent Object
	// Display name of the object
	name string
	// If true, the article "a" is used to refer to the object. If no article
	// is specified none will be used.
	articleA bool
	// If true, the article "an" is used to refer to the object. If no article
	// is specified none will be used.
	articleAn bool
	// The hue of the object
	hue uo.Hue
	// Location of the object
	location uo.Location
	// Facing is the direction the object is facing
	facing uo.Direction
}

// TypeName implements the util.Serializeable interface.
func (o *BaseObject) TypeName() string {
	return "BaseObject"
}

// SerialType implements the util.Serializeable interface.
func (o *BaseObject) SerialType() uo.SerialType {
	return uo.SerialTypeItem
}

// Serialize implements the util.Serializeable interface.
func (o *BaseObject) Serialize(f *util.TagFileWriter) {
	o.BaseSerializeable.Serialize(f)
	if o.parent != nil {
		f.WriteHex("Parent", uint32(o.parent.Serial()))
	} else {
		f.WriteHex("Parent", uint32(uo.SerialSystem))
	}
	f.WriteString("Name", o.name)
	f.WriteBool("ArticleA", o.articleA)
	f.WriteBool("ArticleAn", o.articleAn)
	f.WriteNumber("Hue", int(o.hue))
	f.WriteNumber("X", o.location.X)
	f.WriteNumber("Y", o.location.Y)
	f.WriteNumber("Z", o.location.Z)
	f.WriteNumber("Facing", int(o.facing))
}

// Deserialize implements the util.Serializeable interface.
func (o *BaseObject) Deserialize(f *util.TagFileObject) {
	o.BaseSerializeable.Deserialize(f)
	ps := uo.Serial(f.GetHex("Parent", uint32(uo.SerialSystem)))
	if ps == uo.SerialSystem {
		o.parent = nil
	} else {
		o.parent = world.Find(ps)
	}
	o.name = f.GetString("Name", "unknown entity")
	o.articleA = f.GetBool("ArticleA", false)
	o.articleAn = f.GetBool("ArticleAn", false)
	o.hue = uo.Hue(f.GetNumber("Hue", int(uo.HueIce1)))
	o.location.X = f.GetNumber("X", 1607)
	o.location.Y = f.GetNumber("Y", 1595)
	o.location.Z = f.GetNumber("Z", 13)
	o.facing = uo.Direction(f.GetNumber("Facing", int(uo.DirectionSouth)))
}

// Parent implements the Object interface
func (o *BaseObject) Parent() Object { return o.parent }

// SetParent implements the Object interface
func (o *BaseObject) SetParent(p Object) { o.parent = p }

// SetNewParent implements the Object interface
func (o *BaseObject) SetNewParent(p Object) bool {
	oldParent := o.parent
	if o.parent == nil {
		if !world.Map().RemoveObject(o) {
			return false
		}
	} else {
		if !o.parent.RemoveObject(o) {
			return false
		}
	}
	o.parent = p
	addFailed := false
	if o.parent == nil {
		if !world.Map().AddObject(o) {
			addFailed = true
		}
	} else {
		if !o.parent.AddObject(o) {
			addFailed = true
		}
	}
	if addFailed {
		// Make our best effort to not leak the object
		o.parent = oldParent
		if oldParent == nil {
			world.Map().AddObject(o)
		} else {
			oldParent.AddObject(o)
		}
	}
	return true
}

// RemoveObject implements the Object interface
func (o *BaseObject) RemoveObject(c Object) bool {
	// BaseObject has no child references
	return false
}

// AddObject implements the Object interface
func (o *BaseObject) AddObject(c Object) bool {
	// BaseObject has no child references
	return false
}

// Location implements the Object interface
func (o *BaseObject) Location() uo.Location { return o.location }

// SetLocation implements the Object interface
func (o *BaseObject) SetLocation(l uo.Location) {
	o.location = l
}

// Hue implements the Object interface
func (o *BaseObject) Hue() uo.Hue { return o.hue }

// DisplayName implements the Object interface
func (o *BaseObject) DisplayName() string {
	if o.articleA {
		return "a " + o.name
	}
	if o.articleAn {
		return "an " + o.name
	}
	return o.name
}

// Facing implements the Object interface
func (o *BaseObject) Facing() uo.Direction { return o.facing }

// SetFacing implements the Object interface
func (o *BaseObject) SetFacing(f uo.Direction) {
	o.facing = f.Bound()
}
