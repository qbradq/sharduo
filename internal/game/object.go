package game

import (
	"github.com/qbradq/sharduo/internal/util"
	"github.com/qbradq/sharduo/lib/uo"
)

// Object is the interface every object in the game implements
type Object interface {
	// GetID returns the unique serial of dynamic objects, or SerialMobileNil
	// for static objects.
	GetID() uo.Serial
	// GetLocation returns the current location of the object
	GetLocation() Location
}

// BaseObject is the base of all game objects and implements the Object
// interface
type BaseObject struct {
	util.BaseSerializeable
	// Unique ID number of the object
	ID uo.Serial
	// Item graphic of the object, if any
	Item uo.Item
	// Animation body of the object, if any
	Body uo.Body
	// Display name of the object
	Name string
	// Location of the object
	Location Location
}

// GetID implements the Object interface
func (o *BaseObject) GetID() uo.Serial { return o.ID }

// GetLocation implements the Object interface
func (o *BaseObject) GetLocation() Location { return o.Location }
