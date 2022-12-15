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
	// Location returns the current location of the object
	Location() Location
	// SetLocation sets the absolute location of the object without regard to
	// the map.
	SetLocation(Location)
	// Hue returns the hue of the item
	Hue() uo.Hue
	// DisplayName returns the name of the object with any articles attached
	DisplayName() string
	// Facing returns the direction the object is currently facing. 8-way for
	// mobiles, 2-way for most items, and 4-way for a few items.
	Facing() uo.Direction
	// SetFacing sets the direction the object is currently facing.
	SetFacing(uo.Direction)

// BaseObject is the base of all game objects and implements the Object
// interface
type BaseObject struct {
	util.BaseSerializeable
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
	location Location
	// Facing is the direction the object is facing
	facing uo.Direction
	// Contents is the collection of all the items contained within this object
	inventory Inventory
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
	o.name = f.GetString("Name", "unknown entity")
	o.articleA = f.GetBool("ArticleA", false)
	o.articleAn = f.GetBool("ArticleAn", false)
	o.hue = uo.Hue(f.GetNumber("Hue", int(uo.HueIce1)))
	o.location.X = f.GetNumber("X", 1607)
	o.location.Y = f.GetNumber("Y", 1595)
	o.location.Z = f.GetNumber("Z", 13)
	o.facing = uo.Direction(f.GetNumber("Facing", int(uo.DirectionSouth)))
}

// Location implements the Object interface
func (o *BaseObject) Location() Location { return o.location }

// SetLocation implements the Object interface
func (o *BaseObject) SetLocation(l Location) {
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
func (o *BaseObject) Facing() uo.Direction { return o.facing; }

// SetFacing implements the Object interface
func (o *BaseObject) SetFacing(f uo.Direction) {
	
}
