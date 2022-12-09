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
	// GetBody returns the animation body of the mobile.
	GetBody() uo.Body
	// Equip equips the given item in the item's layer, returns false if the
	// equip operation failed for any reason.
	Equip(Item) bool
	// EquippedMobilePacket returns a serverpacket.EquippedMobile packet for
	// this mobile.
	EquippedMobilePacket() *serverpacket.EquippedMobile
}

// BaseMobile provides the base implementation for Mobile
type BaseMobile struct {
	BaseObject
	// isFemale is true if the mobile is female
	isFemale bool
	// Animation body of the object
	body uo.Body
	// Notoriety of the mobile
	notoriety uo.Notoriety
	// equipment is the collection of equipment this mobile is wearing, if any
	equipment EquipmentCollection
}

// GetTypeName implements the util.Serializeable interface.
func (m *BaseMobile) GetTypeName() string {
	return "BaseMobile"
}

// Serialize implements the util.Serializeable interface.
func (m *BaseMobile) Serialize(f *util.TagFileWriter) {
	m.BaseObject.Serialize(f)
	f.WriteBool("IsFemale", m.isFemale)
	f.WriteNumber("Body", int(m.body))
	f.WriteNumber("Notoriety", int(m.notoriety))
	m.equipment.Write("Equipment", f)
}

// Deserialize implements the util.Serializeable interface.
func (m *BaseMobile) Deserialize(f *util.TagFileObject) {
	m.BaseObject.Deserialize(f)
}

// GetBody implements the Mobile interface.
func (m *BaseMobile) GetBody() uo.Body { return m.body }

// Equip implements the Mobile interface.
func (m *BaseMobile) Equip(item Item) bool {
	return m.equipment.Equip(item)
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
	m.equipment.Map(func(item Item) error {
		p.Equipment = append(p.Equipment, &serverpacket.EquippedMobileItem{
			ID:      item.Serial(),
			Graphic: item.Graphic(),
			Layer:   item.Layer(),
			Hue:     item.Hue(),
		})
		return nil
	})
	return p
}
