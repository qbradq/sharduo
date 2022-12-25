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

	//
	// Parent / child relationships
	//

	// Parent returns a pointer to the parent object of this object, or nil
	// if the object is attached directly to the world
	Parent() Object
	// RootParent returns the top-most parent of the object who's parent is the
	// map. If this object's parent is the map this object is returned.
	RootParent() Object
	// HasParent returns true if the given object is this object's parent, or
	// the parent of any other object in the parent chain.
	HasParent(Object) bool
	// SetParent sets the parent pointer. Use nil to represent the world.
	SetParent(Object)

	//
	// Callbacks
	//

	// RemoveObject removes an object from this object. This is called when
	// changing parent objects. This function should return false if the object
	// could not be removed.
	RemoveObject(Object) bool
	// AddObject adds an object to this object. This is called when changing
	// parent objects. This function should return false if the object could not
	// be added.
	AddObject(Object) bool
	// DropObject is called when another object is dropped onto / into this
	// object by a mobile. A nil mobile usually means a script is generating
	// items directly into a container. This returns false if the drop action
	// is rejected for any reason.
	DropObject(Object, Mobile) bool

	//
	// Player / client interaction callbacks
	//

	// SingleClick is called when a client single-clicks the object
	SingleClick(Mobile)
	// DoubleClick is called when a client double-clicks the object
	DoubleClick(Mobile)

	//
	// Generic accessors
	//

	// Location returns the current location of the object
	Location() uo.Location
	// SetLocation sets the absolute location of the object without regard to
	// the map.
	SetLocation(uo.Location)
	// Hue returns the hue of the item
	Hue() uo.Hue
	// Facing returns the direction the object is currently facing. 8-way for
	// mobiles, 2-way for most items, and 4-way for a few items.
	Facing() uo.Direction
	// SetFacing sets the direction the object is currently facing.
	SetFacing(uo.Direction)

	//
	// Complex accessors
	//

	// DisplayName returns the name of the object with any articles attached
	DisplayName() string
	// Weight returns the total weight of the object. For an item, this is the
	// base weight of the item times the amount. For container items this is the
	// base weight of the item plus the weight of the contents. For mobiles this
	// is the total weight of all equipment including containers, but excluding
	// the bank box if any.
	Weight() int
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
	f.WriteLocation("Location", o.location)
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
	o.hue = uo.Hue(f.GetNumber("Hue", int(uo.HueDefault)))
	o.location = f.GetLocation("Location", uo.Location{
		X: 1324,
		Y: 1624,
		Z: 55,
	})
	o.facing = uo.Direction(f.GetNumber("Facing", int(uo.DirectionSouth)))
}

// Parent implements the Object interface
func (o *BaseObject) Parent() Object { return o.parent }

// RootParent implements the Object interface
func (o *BaseObject) RootParent() Object {
	if o.parent == nil {
		return o
	}
	topmost := o.Parent()
	for {
		if topmost.Parent() == nil {
			return topmost
		}
		topmost = topmost.Parent()
	}
}

// HasParent implements the Object interface
func (o *BaseObject) HasParent(t Object) bool {
	p := o.Parent()
	for {
		if p == nil {
			return false
		}
		if p == t {
			return true
		}
		p = p.Parent()
	}
}

// SetParent implements the Object interface
func (o *BaseObject) SetParent(p Object) { o.parent = p }

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

// DropObject implements the Object interface
func (o *BaseObject) DropObject(obj Object, from Mobile) bool {
	// This makes no sense for a base object
	return false
}

// SingleClick implements the Object interface
func (o *BaseObject) SingleClick(from Mobile) {
	// Default action is to send the name as over-head text
	if from.NetState() != nil {
		from.NetState().Speech(o, o.DisplayName())
	}
}

// DoubleClick implements the Object interface
func (o *BaseObject) DoubleClick(from Mobile) {
	// No default action
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

// Weight implements the Object interface
func (o *BaseObject) Weight() int {
	// This makes no sense for base objects
	return 0
}

// Facing implements the Object interface
func (o *BaseObject) Facing() uo.Direction { return o.facing }

// SetFacing implements the Object interface
func (o *BaseObject) SetFacing(f uo.Direction) {
	o.facing = f.Bound()
}
