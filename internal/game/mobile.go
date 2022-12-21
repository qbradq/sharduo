package game

import (
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseMobile{} })
}

// Mobile is the interface all mobiles implement
type Mobile interface {
	Object
	//
	// NetState
	//

	// NetState returns the NetState implementation currently bound to this
	// mobile, or nil if there is none.
	NetState() NetState
	// SetNetState sets the currently bound NetState. Use SetNetState(nil) to
	// disconnect the mobile.
	SetNetState(NetState)
	// EquippedMobilePacket returns a serverpacket.EquippedMobile packet for
	// this mobile.
	EquippedMobilePacket() *serverpacket.EquippedMobile

	//
	// Stats, attributes, and skills
	//

	// MobileFlags returns the MobileFlags value for this mobile
	MobileFlags() uo.MobileFlags
	// Strength returns the current effective strength
	Strength() int
	// Dexterity returns the current effective dexterity
	Dexterity() int
	// Intelligence returns the current effective intelligence
	Intelligence() int
	// HitPoints returns the current hit points
	HitPoints() int
	// MaxHitPoints returns the current effective max hit points
	MaxHitPoints() int
	// Mana returns the current mana
	Mana() int
	// MaxMana returns the current effective max mana
	MaxMana() int
	// Stamina returns the current stamina
	Stamina() int
	// MaxStamina returns the current effective max stamina
	MaxStamina() int

	//
	// AI-related values
	//

	// ViewRange returns the number of tiles this mobile can see and visually
	// observe objects in the world. If this mobile has an attached NetState,
	// this value can change at any time at the request of the player.
	ViewRange() int
	// SetViewRange sets the view range of the mobile, bounding it to sane
	// values.
	SetViewRange(int)
	// IsRunning returns true if the mobile is running.
	IsRunning() bool
	// SetRunning sets the running flag of the mobile.
	SetRunning(bool)

	//
	// Graphics and display
	//

	// GetBody returns the animation body of the mobile.
	Body() uo.Body
	// IsFemale returns true if the mobile is female.
	IsFemale() bool
	// IsHumanBody returns true if the body value is humanoid.
	IsHumanBody() bool

	//
	// Equipment and inventory
	//

	// ItemInHand returns a pointer to the item held in the mobile's cursor.
	// This is usually only used by mobiles being controlled by a client.
	ItemInCursor() Item
	// Returns true if the mobile's cursor has an item on it.
	IsItemOnCursor() bool
	// SetItemInCursor sets the item held in the mobile's cursor. Returns true
	// if successful.
	SetItemInCursor(Item) bool
	// Equip equips the given item in the item's layer, returns false if the
	// equip operation failed for any reason.
	Equip(Wearable) bool
}

// BaseMobile provides the base implementation for Mobile
type BaseMobile struct {
	BaseObject
	// Attached NetState implementation
	n NetState
	// Current view range of the mobile. Please note that the zero value IS NOT
	// SANE for this variable!
	viewRange int
	// isFemale is true if the mobile is female
	isFemale bool
	// Animation body of the object
	body uo.Body
	// Running flag
	isRunning bool
	// Notoriety of the mobile
	notoriety uo.Notoriety
	// Pointer to the item held in the cursor
	itemInCursor Item
	// The collection of equipment this mobile is wearing, if any
	equipment *EquipmentCollection
	// Mobile flags
	flags uo.MobileFlags
	// Base strength
	baseStrength int
	// Base dexterity
	baseDexterity int
	// Base intelligence
	baseIntelligence int
	// Current HP
	hitPoints int
	// Current mana
	mana int
	// Current stamina
	stamina int
}

// GetTypeName implements the util.Serializeable interface.
func (m *BaseMobile) TypeName() string {
	return "BaseMobile"
}

// SerialType implements the util.Serializeable interface.
func (o *BaseMobile) SerialType() uo.SerialType {
	return uo.SerialTypeMobile
}

// Serialize implements the util.Serializeable interface.
func (m *BaseMobile) Serialize(f *util.TagFileWriter) {
	m.BaseObject.Serialize(f)
	f.WriteNumber("ViewRange", m.viewRange)
	f.WriteBool("IsFemale", m.isFemale)
	f.WriteNumber("Body", int(m.body))
	f.WriteNumber("Notoriety", int(m.notoriety))
	f.WriteNumber("Strength", m.baseStrength)
	f.WriteNumber("Dexterity", m.baseDexterity)
	f.WriteNumber("Intelligence", m.baseIntelligence)
	f.WriteNumber("HitPoints", m.hitPoints)
	f.WriteNumber("Stamina", m.stamina)
	f.WriteNumber("Mana", m.mana)
	if m.equipment != nil {
		m.equipment.Write("Equipment", f)
	}
}

// Deserialize implements the util.Serializeable interface.
func (m *BaseMobile) Deserialize(f *util.TagFileObject) {
	m.BaseObject.Deserialize(f)
	m.viewRange = f.GetNumber("ViewRange", uo.MaxViewRange)
	m.isFemale = f.GetBool("IsFemale", false)
	m.body = uo.Body(f.GetNumber("Body", int(uo.BodyDefault)))
	// Special case for human bodies to select between male and female models
	if m.body == uo.BodyHuman && m.isFemale {
		m.body += 1
	}
	m.baseStrength = f.GetNumber("Strength", 10)
	m.baseDexterity = f.GetNumber("Dexterity", 10)
	m.baseIntelligence = f.GetNumber("Intelligence", 10)
	m.hitPoints = f.GetNumber("HitPoints", 1)
	m.mana = f.GetNumber("Mana", 1)
	m.stamina = f.GetNumber("Stamina", 1)
	m.notoriety = uo.Notoriety(f.GetNumber("Notoriety", int(uo.NotorietyInnocent)))
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (m *BaseMobile) OnAfterDeserialize(f *util.TagFileObject) {
	m.equipment = NewEquipmentCollectionWith(f.GetObjectReferences("Equipment"))
}

// NetState implements the Mobile interface.
func (m *BaseMobile) NetState() NetState { return m.n }

// SetNetState implements the Mobile interface.
func (m *BaseMobile) SetNetState(n NetState) {
	m.n = n
}

// ViewRange implements the Mobile interface.
func (m *BaseMobile) ViewRange() int { return m.viewRange }

// SetViewRange implements the Mobile interface.
func (m *BaseMobile) SetViewRange(r int) { m.viewRange = uo.BoundViewRange(r) }

// Body implements the Mobile interface.
func (m *BaseMobile) Body() uo.Body { return m.body }

// IsFemale implements the Mobile interface.
func (m *BaseMobile) IsFemale() bool { return m.isFemale }

// IsHumanBody implements the Mobile interface.
func (m *BaseMobile) IsHumanBody() bool {
	return m.body == uo.BodyHumanMale || m.body == uo.BodyHumanFemale
}

// IsRunning implements the Mobile interface.
func (m *BaseMobile) IsRunning() bool { return m.isRunning }

// SetRunning implements the Mobile interface.
func (m *BaseMobile) SetRunning(v bool) { m.isRunning = v }

// MobileFlags implements the Mobile interface.
func (m *BaseMobile) MobileFlags() uo.MobileFlags { return m.flags }

// Strength implements the Mobile interface.
func (m *BaseMobile) Strength() int { return m.baseStrength }

// Dexterity implements the Mobile interface.
func (m *BaseMobile) Dexterity() int { return m.baseDexterity }

// Intelligence implements the Mobile interface.
func (m *BaseMobile) Intelligence() int { return m.baseIntelligence }

// HitPoints implements the Mobile interface.
func (m *BaseMobile) HitPoints() int { return m.hitPoints }

// MaxHitPoints implements the Mobile interface.
func (m *BaseMobile) MaxHitPoints() int { return 50 + m.Strength()/2 }

// Mana implements the Mobile interface.
func (m *BaseMobile) Mana() int { return m.mana }

// MaxMana implements the Mobile interface.
func (m *BaseMobile) MaxMana() int { return m.Intelligence() }

// Stamina implements the Mobile interface.
func (m *BaseMobile) Stamina() int { return m.stamina }

// MaxStamina implements the Mobile interface.
func (m *BaseMobile) MaxStamina() int { return m.Dexterity() }

func (m *BaseMobile) ItemInCursor() Item { return m.itemInCursor }

// Returns true if the mobile's cursor has an item on it.
func (m *BaseMobile) IsItemOnCursor() bool { return m.itemInCursor != nil }

// SetItemInCursor sets the item held in the mobile's cursor. It returns true
// if successful.
func (m *BaseMobile) SetItemInCursor(item Item) bool {
	if m.itemInCursor == item {
		return true
	}
	if m.itemInCursor != nil {
		return false
	}
	m.itemInCursor = item
	return world.Map().SetNewParent(item, m)
}

// AddObject adds the object to the mobile. It returns true if successful.
func (m *BaseMobile) AddObject(o Object) bool {
	if item, ok := o.(Item); ok {
		if m.itemInCursor == item {
			return true
		}
	}
	return false
}

// RemoveObject removes the object from the mobile. It returns true if
// successful.
func (m *BaseMobile) RemoveObject(o Object) bool {
	if m.itemInCursor == o {
		m.itemInCursor = nil
		return true
	}
	return false
}

// Equip implements the Mobile interface.
func (m *BaseMobile) Equip(w Wearable) bool {
	if m.equipment == nil {
		m.equipment = NewEquipmentCollection()
	}
	return m.equipment.Equip(w)
}

// EquippedMobilePacket implements the Mobile interface.
func (m *BaseMobile) EquippedMobilePacket() *serverpacket.EquippedMobile {
	flags := uo.MobileFlagNone
	if m.isFemale {
		flags |= uo.MobileFlagFemale
	}
	p := &serverpacket.EquippedMobile{
		ID:        m.Serial(),
		Body:      m.Body(),
		X:         m.location.X,
		Y:         m.location.Y,
		Z:         m.location.Z,
		Facing:    m.facing,
		IsRunning: m.IsRunning(),
		Hue:       m.Hue(),
		Flags:     flags,
		Notoriety: m.notoriety,
	}
	if m.equipment != nil {
		m.equipment.Map(func(w Wearable) error {
			p.Equipment = append(p.Equipment, &serverpacket.EquippedMobileItem{
				ID:      w.Serial(),
				Graphic: w.Graphic(),
				Layer:   w.Layer(),
				Hue:     w.Hue(),
			})
			return nil
		})
	}
	return p
}
