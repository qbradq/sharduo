package game

import (
	"image/color"
	"log"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("BaseObject", marshal.ObjectTypeObject, func() any { return &BaseObject{} })
}

// reg registers a constructor with the template and marshal packages.
func reg(name string, ot marshal.ObjectType, fn func() any) {
	template.RegisterConstructor(name, fn)
	marshal.RegisterCtor(ot, fn)
}

// Object is the interface every object in the game implements
type Object interface {
	Deserialize(*template.Template, bool)
	marshal.Marshaler
	marshal.Unmarshaler

	// List of all events supported by all Objects
	// DoubleClick......................Player double-clicks on object

	//
	// Lifecycle management
	//

	// Removed returns true if this object has been removed from the data store
	Removed() bool
	// Remove flags the object as having been removed from the data store
	Remove()
	// NoRent returns true if this object should not persist through server
	// restart.
	NoRent() bool
	// SpawnerRegion returns a pointer to the region that spawned this object
	// if any.
	SpawnerRegion() *Region
	// SetSpawnerRegion sets the spawner region.
	SetSpawnerRegion(*Region)

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
	// RemoveChildren is responsible for calling Remove() for all child objects.
	RemoveChildren()
	// TemplateName returns the name of the template used to create this object.
	TemplateName() string
	// SetTemplateName sets the name of the template used to create this object.
	// For internal use only.
	SetTemplateName(string)
	// BaseTemplate returns the name of the direct base template used when
	// creating this object.
	BaseTemplate() string

	//
	// Callbacks
	//

	// Update handles updates for objects. Mobiles should receive an Update()
	// call every tick, while Items (on the ground) should receive one every
	// minute.
	Update(uo.Time)
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
	// InsertObject adds an object as a child of this object through an empty
	// interface.
	InsertObject(any)
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
	// AppendTemplateContextMenuEntry is used to add a context menu entry to the
	// list of context menu entries from the template's Events line.
	AppendTemplateContextMenuEntry(string, uo.Cliloc)
	// AppendContextMenuEntries is called when building the context menu of an
	// object. The function may append any entries it needs with the
	// ContextMenu.Append function. The second parameter represents the mobile
	// requesting the context menu, which should always be a player's mobile
	// with attached NetState.
	AppendContextMenuEntries(*ContextMenu, Mobile)

	//
	// Object Property List functionality
	//

	// OPLPackets returns the current OPL packets for this object. The Object
	// parameter must be a self-reference to this object.
	OPLPackets(Object) (*serverpacket.OPLPacket, *serverpacket.OPLInfo)
	// InvalidateOPL flags the OPLPacket for this object as stale and in need of
	// rebuilding.
	InvalidateOPL()
	// AppendOPLEntires is called when building the object property list of an
	// object. The function may append any entries it needs with the
	// OPLPacket.Append and OPLPacket.AppendColor functions. The first parameter
	// must be the object self-reference.
	AppendOPLEntires(Object, *serverpacket.OPLPacket)

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
	// SetHue sets the hue of the item. Remember to use hue.SetPartial() if a
	// partial hue is desired.
	SetHue(uo.Hue)
	// Facing returns the direction the object is currently facing. 8-way for
	// mobiles, 2-way for most items, and 4-way for a few items.
	Facing() uo.Direction
	// SetFacing sets the direction the object is currently facing.
	SetFacing(uo.Direction)
	// Visibility returns the current visibility level of the object.
	Visibility() uo.Visibility
	// SetAmount sets the amount of the stack, if any. If this is out of range
	// it will be bounded to a sane value. Generally this only applies to items.
	SetAmount(int)

	//
	// Complex accessors
	//

	// DisplayName returns the name of the object with any articles attached.
	DisplayName() string
	// SetName sets the name of the object and clears all article flags.
	SetName(string)
	// Name returns the name of the object without any articles attached.
	Name() string
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
	//
	// Persistent values
	//

	// Unique serial of the object
	serial uo.Serial
	// Name of the template this object was constructed from
	templateName string
	// Name of the base template used when constructing the object
	baseTemplate string
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
	// Owner of the object if any
	owner Object

	//
	// Non-persistent values
	//

	// Collection of all event handler names
	eventHandlers map[string]string
	// List of all the context menu entries from the template
	templateContextMenuEntries []ContextMenuEntry
	// If true this object has already been removed from the datastore
	removed bool
	// OPLPacket cache
	opl *serverpacket.OPLPacket
	// OPLInfo cache
	oplinfo *serverpacket.OPLInfo
	// Spawner region pointer
	spawnerRegion *Region
}

// ObjectType implements the Object interface.
func (o *BaseObject) ObjectType() marshal.ObjectType { return marshal.ObjectTypeObject }

// SerialType implements the util.Serializeable interface.
func (o *BaseObject) SerialType() uo.SerialType {
	return uo.SerialTypeItem
}

// Removed implements the Object interface.
func (o *BaseObject) Removed() bool { return o.removed }

// Remove implements the Object interface.
func (o *BaseObject) Remove() { o.removed = true }

// NoRent implements the Object interface.
func (o *BaseObject) NoRent() bool { return false }

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
	os := uo.SerialSystem
	if o.owner != nil {
		os = o.owner.Serial()
	}
	s.PutInt(0) // version
	s.PutInt(uint32(ps))
	s.PutInt(uint32(os))
	s.PutString(o.name)
	s.PutShort(uint16(o.hue))
	s.PutLocation(o.location)
	s.PutByte(byte(o.facing))
}

// Deserialize implements the util.Serializeable interface.
func (o *BaseObject) Deserialize(t *template.Template, create bool) {
	o.templateName = t.TemplateName
	if o.templateName == "" {
		// Something is very wrong
		log.Printf("warning: object %s has no TemplateName property", o.Serial().String())
		return
	}
	o.baseTemplate = t.BaseTemplate
	o.name = t.GetString("Name", "unknown entity")
	o.articleA = t.GetBool("ArticleA", false)
	o.articleAn = t.GetBool("ArticleAn", false)
	o.hue = uo.Hue(t.GetNumber("Hue", int(uo.HueDefault)))
	// Event handling
	events := t.GetString("Events", "")
	for idx, pair := range strings.Split(events, ",") {
		if pair == "" {
			continue
		}
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			log.Printf("warning: object %s event %d malformed pair", o.Serial().String(), idx)
			continue
		}
		o.LinkEvent(parts[0], parts[1])
	}
	// Context menu handling
	contextMenu := t.GetString("ContextMenu", "")
	for idx, pair := range strings.Split(contextMenu, ",") {
		if pair == "" {
			continue
		}
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			log.Printf("warning: object %s menu entry %d malformed pair", o.Serial().String(), idx)
			continue
		}
		cliloc, err := strconv.ParseUint(parts[0], 0, 32)
		if err != nil {
			log.Printf("warning: object %s menu entry %d malformed cliloc number", o.Serial().String(), idx)
			continue
		}
		o.AppendTemplateContextMenuEntry(parts[1], uo.Cliloc(cliloc))
	}
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (o *BaseObject) Unmarshal(s *marshal.TagFileSegment) {
	_ = s.Int() // version
	// Parent object resolution
	ps := uo.Serial(s.Int())
	if ps == uo.SerialSystem {
		o.parent = nil
	} else if ps == uo.SerialTheVoid {
		// This is the case for logged out player characters
		o.parent = TheVoid
	} else if ps == uo.SerialZero {
		log.Printf("warning: object %s has no parent", o.Serial().String())
		o.parent = nil
	} else {
		o.parent = world.Find(ps)
	}
	// Owner object resolution
	os := uo.Serial(s.Int())
	if os == uo.SerialSystem {
		o.owner = nil
	} else if ps == uo.SerialTheVoid {
		o.owner = TheVoid
	} else if ps == uo.SerialZero {
		log.Printf("warning: object %s has no owner", o.Serial().String())
		o.owner = nil
	} else {
		o.owner = world.Find(os)
	}
	// Unmarshal all dynamic values
	o.name = s.String()
	o.hue = uo.Hue(s.Short())
	o.location = s.Location()
	o.facing = uo.Direction(s.Byte())
}

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

// SetTemplateName implements the Object interface
func (o *BaseObject) SetTemplateName(name string) { o.templateName = name }

// RecalculateStats implements the Object interface
func (o *BaseObject) RecalculateStats() {}

// RemoveChildren implements  the Object interface
func (o *BaseObject) RemoveChildren() {}

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

// InsertObject implements the Object interface. PLEASE NOTE that a call to
// BaseObject.InsertObject() will leak the object!
func (o *BaseObject) InsertObject(obj any) {
	other, ok := obj.(Object)
	if !ok {
		return
	}
	o.ForceAddObject(other)
}

// ForceRemoveObject implements the Object interface. PLEASE NOTE that a call to
// BaseObject.ForceRemoveObject() will leak the object!
func (o *BaseObject) ForceRemoveObject(obj Object) {
	// BaseObject has no child references
}

// AfterUnmarshalOntoMap implements the Object interface.
func (o *BaseObject) AfterUnmarshalOntoMap() {}

// AppendTemplateContextMenuEntry implements the Object interface
func (o *BaseObject) AppendTemplateContextMenuEntry(event string, cl uo.Cliloc) {
	o.templateContextMenuEntries = append(o.templateContextMenuEntries, ContextMenuEntry{
		Event:  event,
		Cliloc: cl,
	})
}

// AppendContextMenuEntries implements the Object interface
func (o *BaseObject) AppendContextMenuEntries(m *ContextMenu, src Mobile) {
	for _, e := range o.templateContextMenuEntries {
		m.Append(e.Event, e.Cliloc)
	}
}

// OPLPackets implements the Object interface.
func (o *BaseObject) OPLPackets(self Object) (*serverpacket.OPLPacket, *serverpacket.OPLInfo) {
	if o.opl == nil {
		o.opl = &serverpacket.OPLPacket{
			Serial: o.serial,
		}
		self.AppendOPLEntires(self, o.opl)
		o.opl.Compile()
		o.oplinfo = &serverpacket.OPLInfo{
			Serial: o.serial,
			Hash:   o.opl.Hash,
		}
	}
	return o.opl, o.oplinfo
}

// InvalidateOPL implements the Object interface.
func (o *BaseObject) InvalidateOPL() {
	o.opl = nil
	o.oplinfo = nil
	world.UpdateOPLInfo(o)
}

// AppendOPLEntries implements the Object interface.
func (o *BaseObject) AppendOPLEntires(r Object, p *serverpacket.OPLPacket) {
	p.AppendColor(color.White, r.DisplayName(), false)
}

// Location implements the Object interface
func (o *BaseObject) Location() uo.Location { return o.location }

// SetLocation implements the Object interface
func (o *BaseObject) SetLocation(l uo.Location) {
	o.location = l
}

// Hue implements the Object interface
func (o *BaseObject) Hue() uo.Hue { return o.hue }

// SetHue implements the Object interface
func (o *BaseObject) SetHue(hue uo.Hue) {
	o.hue = hue
	world.Update(o)
}

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
	world.Update(o)
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
	eventHandler, found := o.eventHandlers[which]
	if !found {
		return nil
	}
	h := eventHandlerGetter(eventHandler)
	if h == nil {
		log.Printf("error: reference to unregistered event handler \"%s\"", eventHandler)
	}
	return h
}

// Visibility implements the Object interface.
func (o *BaseObject) Visibility() uo.Visibility {
	return uo.VisibilityVisible
}

// Update implements the Object interface.
func (o *BaseObject) Update(t uo.Time) {}

// Owner implements the Object interface.
func (o *BaseObject) Owner() Object { return o.owner }

// SetOwner implements the Object interface.
func (o *BaseObject) SetOwner(owner Object) { o.owner = owner }

// SetName implements the Object interface.
func (o *BaseObject) SetName(name string) {
	o.name = name
	o.articleA = false
	o.articleAn = false
	o.InvalidateOPL()
	world.Update(o)
}

// Name implements the Object interface.
func (o *BaseObject) Name() string { return o.name }

// BaseTemplate implements the Object interface.
func (o *BaseObject) BaseTemplate() string { return o.baseTemplate }

// SpawnerRegion implements the Object interface.
func (o *BaseObject) SpawnerRegion() *Region { return o.spawnerRegion }

// SetSpawnerRegion implements the Object interface.
func (o *BaseObject) SetSpawnerRegion(r *Region) { o.spawnerRegion = r }

// SetAmount implements the Object interface.
func (o *BaseObject) SetAmount(n int) {}
