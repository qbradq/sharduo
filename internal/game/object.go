package game

import (
	"log"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

var objctors = make(map[string]func() Object)

func init() {
	objctors["BaseObject"] = func() Object { return &BaseObject{} }
	marshal.RegisterCtor(marshal.ObjectTypeObject, func() interface{} { return &BaseObject{} })
}

// Constructor returns the constructor for the given type string or nil
func Constructor(which string) func() Object {
	return objctors[which]
}

// Object is the interface every object in the game implements
type Object interface {
	Deserialize(*util.TagFileObject)
	marshal.Marshaler
	marshal.Unmarshaler

	// List of all events supported by all Objects
	// OnDoubleClick....................Player double-clicks on object

	//
	// Parent / child relationships
	//

	// Serial returns the serial of the object
	Serial() uo.Serial
	// SetSerial sets the serial of the object, only used during object creation
	SetSerial(uo.Serial)
	// SerialType returns what kind of serial should be generated for this type
	// of object.
	SerialType() uo.SerialType
	// Parent returns a pointer to the parent object of this object, or nil
	// if the object is attached directly to the world
	Parent() Object
	// HasParent returns true if the given object is this object's parent, or
	// the parent of any other object in the parent chain.
	HasParent(Object) bool
	// SetParent sets the parent pointer. Use nil to represent the world.
	SetParent(Object)
	// ObjectType returns the marshal.ObjectType associated with this struct.
	ObjectType() marshal.ObjectType
	// TemplateName returns the name of the template used to create this object.
	TemplateName() string

	//
	// Callbacks
	//

	// LinkEvent links the named handler to this object's event callbacks. Use
	// the global function ExecuteEventHandler.
	LinkEvent(string, string)
	// GetEventHandler returns the named link function or nil.
	GetEventHandler(string) *EventHandler
	// RecalculateStats is called after an object has been unmarshaled and
	// should be used to recalculate dynamic attributes.
	RecalculateStats()
	// RemoveObject removes an object from this object. This is called when
	// changing parent objects. This function should return false if the object
	// could not be removed.
	RemoveObject(Object) bool
	// AddObject adds an object to this object. This is called when changing
	// parent objects. This function should return false if the object could not
	// be added.
	AddObject(Object) bool
	// ForceAddObject is like AddObject but should not fail for any reason.
	ForceAddObject(Object)
	// ForceRemoveObject is like RemoveObject but should not fail for any
	// reason.
	ForceRemoveObject(Object)
	// AfterUnmarshalOntoMap is called only when 1) The object has just been
	// reconstructed from a save file 2) The object's parent is nil - the map
	// If these conditions are true, then this function will be called for all
	// objects after all objects have been placed into the map data structure.
	// This is used for recalculating dynamic values that require spatial
	// awareness, such as the surface a mobile is standing on.
	AfterUnmarshalOntoMap()
	// DropObject is called when another object is dropped onto / into this
	// object by a mobile. A nil mobile usually means a script is generating
	// items directly into a container. This returns false if the drop action
	// is rejected for any reason.
	DropObject(Object, uo.Location, Mobile) bool
	// SingleClick is called when a client single-clicks the object
	SingleClick(Mobile)

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
	Weight() float32
}

// BaseObject is the base of all game objects and implements the Object
// interface
type BaseObject struct {
	// Unique serial of the object
	serial uo.Serial
	// Name of the template this object was constructed from
	templateName string
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
	// Collection of all event handler names
	eventHandlers map[string]string
}

// ObjectType implements the Object interface.
func (o *BaseObject) ObjectType() marshal.ObjectType { return marshal.ObjectTypeObject }

// SerialType implements the util.Serializeable interface.
func (o *BaseObject) SerialType() uo.SerialType {
	return uo.SerialTypeItem
}

// Serial implements the Object interface.
func (o *BaseObject) Serial() uo.Serial { return o.serial }

// SetSerial implements the Object interface.
func (o *BaseObject) SetSerial(s uo.Serial) { o.serial = s }

// Marshal implements the marshal.Marshaler interface.
func (o *BaseObject) Marshal(s *marshal.TagFileSegment) {
	ps := uo.SerialSystem
	if o.parent != nil {
		ps = o.parent.Serial()
	}
	s.PutString(o.templateName)
	s.PutInt(uint32(ps))
	s.PutString(o.name)
	s.PutShort(uint16(o.hue))
	s.PutLocation(o.location)
	s.PutStringsMap(o.eventHandlers)
	s.PutTag(marshal.TagArticleA, marshal.TagValueBool, o.articleA)
	s.PutTag(marshal.TagArticleAn, marshal.TagValueBool, o.articleAn)
	s.PutTag(marshal.TagFacing, marshal.TagValueByte, byte(o.facing))
}

// deserializeEvent attempts to deserialize the named event handler
func (o *BaseObject) deserializeEvent(which string, tfo *util.TagFileObject) {
	eventName := tfo.GetString(which, "")
	if eventName == "" {
		return
	}
	o.LinkEvent(which, eventName)
}

// Deserialize implements the util.Serializeable interface.
func (o *BaseObject) Deserialize(f *util.TagFileObject) {
	o.templateName = f.GetString("TemplateName", "")
	if o.templateName == "" {
		// Something is very wrong
		log.Printf("warning: object %s has no TemplateName property", o.Serial().String())
		return
	}
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
	// Events
	o.deserializeEvent("OnDoubleClick", f)
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (o *BaseObject) Unmarshal(s *marshal.TagFileSegment) *marshal.TagCollection {
	o.templateName = s.String()
	ps := uo.Serial(s.Int())
	if ps == uo.SerialSystem {
		o.parent = nil
	} else if ps == uo.SerialZero {
		log.Printf("warning: object %s has no parent", o.Serial().String())
		o.parent = nil
	} else {
		o.parent = world.Find(ps)
	}
	o.name = s.String()
	o.hue = uo.Hue(s.Short())
	o.location = s.Location()
	o.eventHandlers = s.StringMap()
	tags := s.Tags()
	o.articleA = tags.Bool(marshal.TagArticleA)
	o.articleAn = tags.Bool(marshal.TagArticleAn)
	o.facing = uo.Direction(tags.Byte(marshal.TagFacing))
	return tags
}

// AfterUnmarshal implements the marshal.Unmarshaler interface.
func (o *BaseObject) AfterUnmarshal(tags *marshal.TagCollection) {}

// Parent implements the Object interface
func (o *BaseObject) Parent() Object { return o.parent }

// HasParent implements the Object interface
func (o *BaseObject) HasParent(t Object) bool {
	p := o.Parent()
	for {
		if p == nil {
			return false
		}
		if p.Serial() == t.Serial() {
			return true
		}
		p = p.Parent()
	}
}

// SetParent implements the Object interface
func (o *BaseObject) SetParent(p Object) { o.parent = p }

// TemplateName implements the Object interface
func (o *BaseObject) TemplateName() string { return o.templateName }

// RecalculateStats implements the Object interface
func (o *BaseObject) RecalculateStats() {}

// RemoveObject implements the Object interface
func (o *BaseObject) RemoveObject(c Object) bool {
	// BaseObject has no child references
	return true
}

// AddObject implements the Object interface
func (o *BaseObject) AddObject(c Object) bool {
	// BaseObject has no child references
	o.SetParent(c)
	return false
}

// ForceAddObject implements the Object interface. PLEASE NOTE that a call to
// BaseObject.ForceAddObject() will leak the object!
func (o *BaseObject) ForceAddObject(obj Object) {
	// BaseObject has no child references
	obj.SetParent(o)
}

// ForceRemoveObject implements the Object interface. PLEASE NOTE that a call to
// BaseObject.ForceRemoveObject() will leak the object!
func (o *BaseObject) ForceRemoveObject(obj Object) {
	// BaseObject has no child references
}

// AfterUnmarshalOntoMap implements the Object interface.
func (o *BaseObject) AfterUnmarshalOntoMap() {}

// DropObject implements the Object interface
func (o *BaseObject) DropObject(obj Object, l uo.Location, from Mobile) bool {
	// This makes no sense for a base object
	return false
}

func defaultSingleClickHandler(o Object, from Mobile) {
	// Default action is to send the name as over-head text
	if from.NetState() != nil {
		from.NetState().Speech(o, o.DisplayName())
	}
}

// SingleClick implements the Object interface
func (o *BaseObject) SingleClick(from Mobile) {
	defaultSingleClickHandler(o, from)
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
func (o *BaseObject) Weight() float32 {
	// This makes no sense for base objects
	return 0
}

// Facing implements the Object interface
func (o *BaseObject) Facing() uo.Direction { return o.facing }

// SetFacing implements the Object interface
func (o *BaseObject) SetFacing(f uo.Direction) {
	o.facing = f.Bound()
}

// LinkEvent implements the Object interface
func (o *BaseObject) LinkEvent(event, handler string) {
	if event == "" {
		return
	}
	if handler == "" {
		if o.eventHandlers == nil {
			return
		}
		delete(o.eventHandlers, event)
		return
	}
	if o.eventHandlers == nil {
		o.eventHandlers = make(map[string]string)
	}
	o.eventHandlers[event] = handler
}

// GetEventHandler implements the Object interface
func (o *BaseObject) GetEventHandler(which string) *EventHandler {
	if o.eventHandlers == nil {
		return nil
	}
	eventHandler := o.eventHandlers[which]
	return eventHandlerGetter(eventHandler)
}
