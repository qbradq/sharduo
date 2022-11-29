package game

import (
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

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
	// Collection of currently equipped items
	Equipment map[uo.Layer]Item
	// Notoriety of the mobile
	Notoriety uo.Notoriety
}

// GetBody implements the Mobile interface.
func (m *BaseMobile) GetBody() uo.Body { return m.Body }

// Equip implements the Mobile interface.
func (m *BaseMobile) Equip(item Item) bool {
	if m.Equipment == nil {
		m.Equipment = make(map[uo.Layer]Item)
	}
	if _, duplicate := m.Equipment[item.GetLayer()]; duplicate {
		return false
	}
	m.Equipment[item.GetLayer()] = item
	return true
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
	for _, item := range m.Equipment {
		p.Equipment = append(p.Equipment, &serverpacket.EquippedMobileItem{
			ID:      item.GetSerial(),
			Graphic: item.GetGraphic(),
			Layer:   item.GetLayer(),
			Hue:     item.GetHue(),
		})
	}
	return p
}
