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
	// GetLocation returns the current location of the object
	GetLocation() Location
	// GetHue returns the hue of the item
	GetHue() uo.Hue
	// GetDisplayName returns the name of the object with any articles attached
	GetDisplayName() string
}

// BaseObject is the base of all game objects and implements the Object
// interface
type BaseObject struct {
	util.BaseSerializeable
	// Display name of the object
	Name string
	// If true, the article "a" is used to refer to the object. If no article
	// is specified none will be used.
	ArticleA bool
	// If true, the article "an" is used to refer to the object. If no article
	// is specified none will be used.
	ArticleAn bool
	// The hue of the object
	Hue uo.Hue
	// Location of the object
	Location Location
	// Facing is the direction the object is facing
	Facing uo.Direction
	// Contents is the collection of all the items contained within this object
	Inventory Inventory
}

// GetTypeName implements the util.Serializeable interface.
func (o *BaseObject) GetTypeName() string {
	return "BaseObject"
}

// GetSerialType implements the util.Serializeable interface.
func (o *BaseObject) GetSerialType() uo.SerialType {
	return uo.SerialTypeItem
}

// Serialize implements the util.Serializeable interface.
func (o *BaseObject) Serialize(f *util.TagFileWriter) {
	o.BaseSerializeable.Serialize(f)
	f.WriteString("Name", o.Name)
	f.WriteBool("ArticleA", o.ArticleA)
	f.WriteBool("ArticleAn", o.ArticleA)
	f.WriteNumber("Hue", int(o.Hue))
	f.WriteNumber("X", o.Location.X)
	f.WriteNumber("Y", o.Location.Y)
	f.WriteNumber("Z", o.Location.Z)
	f.WriteNumber("Facing", int(o.Facing))
}

// Deserialize implements the util.Serializeable interface.
func (o *BaseObject) Deserialize(f *util.TagFileObject) {
	o.BaseSerializeable.Deserialize(f)
	o.Name = f.GetString("Name", "unknown entity")
	o.ArticleA = f.GetBool("ArticleA", false)
	o.ArticleAn = f.GetBool("ArticleAn", false)
	o.Hue = uo.Hue(f.GetNumber("Hue", int(uo.HueIce1)))
	o.Location.X = f.GetNumber("X", 1607)
	o.Location.Y = f.GetNumber("Y", 1595)
	o.Location.Z = f.GetNumber("Z", 13)
	o.Facing = uo.Direction(f.GetNumber("Facing", int(uo.DirectionSouth)))
}

// GetLocation implements the Object interface
func (o *BaseObject) GetLocation() Location { return o.Location }

// GetHue implements the Object interface
func (o *BaseObject) GetHue() uo.Hue { return o.Hue }

// GetDisplayName implements the Object interface
func (o *BaseObject) GetDisplayName() string {
	if o.ArticleA {
		return "a " + o.Name
	}
	if o.ArticleAn {
		return "an " + o.Name
	}
	return o.Name
}
