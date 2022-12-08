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
	// IsFemale is true if the mobile is female
	IsFemale bool
	// Animation body of the object
	Body uo.Body
	// Notoriety of the mobile
	Notoriety uo.Notoriety
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
	f.WriteBool("IsFemale", m.IsFemale)
	f.WriteNumber("Body", int(m.Body))
	f.WriteNumber("Notoriety", int(m.Notoriety))
	m.equipment.Write("Equipment", f)
}

// Deserialize implements the util.Serializeable interface.
func (m *BaseMobile) Deserialize(f *util.TagFileObject) {
	m.BaseObject.Deserialize(f)
}

// GetBody implements the Mobile interface.
func (m *BaseMobile) GetBody() uo.Body { return m.Body }

// Equip implements the Mobile interface.
func (m *BaseMobile) Equip(item Item) bool {
	return m.equipment.Equip(item)
}

// EquippedMobilePacket implements the Mobile interface.
func (m *BaseMobile) EquippedMobilePacket() *serverpacket.EquippedMobile {
	flags := uo.MobileFlagNone
	if m.IsFemale {
		flags |= uo.MobileFlagFemale
	}
	p := &serverpacket.EquippedMobile{
		ID:        m.Serial,
		Body:      m.Body,
		X:         m.Location.X,
		Y:         m.Location.Y,
		Z:         m.Location.Z,
		Facing:    m.Facing,
		Hue:       m.Hue,
		Flags:     flags,
		Notoriety: m.Notoriety,
	}
	m.equipment.Map(func(item Item) error {
		p.Equipment = append(p.Equipment, &serverpacket.EquippedMobileItem{
			ID:      item.GetSerial(),
			Graphic: item.GetGraphic(),
			Layer:   item.GetLayer(),
			Hue:     item.GetHue(),
		})
		return nil
	})
	return p
}
