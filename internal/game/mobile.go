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
	// NetState returns the NetState implementation currently bound to this
	// mobile, or nil if there is none.
	NetState() NetState
	// SetNetState sets the currently bound NetState. Use SetNetState(nil) to
	// disconnect the mobile.
	SetNetState(NetState)
	// GetBody returns the animation body of the mobile.
	Body() uo.Body
	// Equip equips the given item in the item's layer, returns false if the
	// equip operation failed for any reason.
	Equip(Wearable) bool
	// EquippedMobilePacket returns a serverpacket.EquippedMobile packet for
	// this mobile.
	EquippedMobilePacket() *serverpacket.EquippedMobile
}

// BaseMobile provides the base implementation for Mobile
type BaseMobile struct {
	BaseObject
	// Attached NetState implementation
	n NetState
	// isFemale is true if the mobile is female
	isFemale bool
	// Animation body of the object
	body uo.Body
	// Notoriety of the mobile
	notoriety uo.Notoriety
	// equipment is the collection of equipment this mobile is wearing, if any
	equipment *EquipmentCollection
}

// GetTypeName implements the util.Serializeable interface.
func (m *BaseMobile) TypeName() string {
	return "BaseMobile"
}

// Serialize implements the util.Serializeable interface.
func (m *BaseMobile) Serialize(f *util.TagFileWriter) {
	m.BaseObject.Serialize(f)
	f.WriteBool("IsFemale", m.isFemale)
	f.WriteNumber("Body", int(m.body))
	f.WriteNumber("Notoriety", int(m.notoriety))
	if m.equipment != nil {
		m.equipment.Write("Equipment", f)
	}
}

// Deserialize implements the util.Serializeable interface.
func (m *BaseMobile) Deserialize(f *util.TagFileObject) {
	m.BaseObject.Deserialize(f)
	m.isFemale = f.GetBool("IsFemale", false)
	m.body = uo.Body(f.GetNumber("Body", int(uo.BodyDefault)))
	// Special case for human bodies to select between male and female models
	if m.body == uo.BodyHuman && m.isFemale {
		m.body += 1
	}
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

// Body implements the Mobile interface.
func (m *BaseMobile) Body() uo.Body { return m.body }

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
		Body:      m.body,
		X:         m.location.X,
		Y:         m.location.Y,
		Z:         m.location.Z,
		Facing:    m.facing,
		Hue:       m.hue,
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
