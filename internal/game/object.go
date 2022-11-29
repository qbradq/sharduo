package game

import (
	"github.com/qbradq/sharduo/internal/util"
	"github.com/qbradq/sharduo/lib/uo"
	orderedmap "github.com/wk8/go-ordered-map"
)

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
	Contents *orderedmap.OrderedMap
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

// AddObject adds an object to the contents of this one
func (o *BaseObject) AddObject(newObj Object) {
	if o.Contents == nil {
		o.Contents = orderedmap.New()
	}
	if _, duplicate := o.Contents.Get(newObj.GetSerial()); duplicate {
		return
	}
	o.Contents.Set(newObj.GetSerial(), newObj)
}
