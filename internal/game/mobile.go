package game

import (
	"io"

	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
	"golang.org/x/image/colornames"
)

// Mobile describes a thinking actor.
type Mobile struct {
	Object
	// Static values
	Body          uo.Body      // Body to use for this mobile
	BaseNotoriety uo.Notoriety // Base notoriety level for this mobile
	// Persistent values
	Female           bool                 // If true the mobile is female
	Player           bool                 // If true this is a player's character
	BaseStrength     int                  // Base strength value
	BaseDexterity    int                  // Base dexterity value
	BaseIntelligence int                  // Base intelligence value
	MaxHits          int                  // Current max hit points
	MaxMana          int                  // Current max mana
	MaxStamina       int                  // Current max stamina
	BaseSkills       [uo.SkillCount]int   // Base skill values in tenths of a percent
	Equipment        [uo.LayerCount]*Item // Current equipment set
	// Transient values
	Account       *Account                // Connected account if any
	NetState      NetState                // Connected net state if any
	ControlMaster *Mobile                 // Mobile that is currently commanding this mobile
	Hits          int                     // Current hit points
	Mana          int                     // Current mana
	Stamina       int                     // Current stamina
	Strength      int                     // Current strength value
	Dexterity     int                     // Current dexterity value
	Intelligence  int                     // Current intelligence value
	Skills        [uo.SkillCount]int      // Current skill values in tenths of a percent
	Running       bool                    // If true the mobile is running
	Weight        float64                 // Current weight of all equipment plus the contents of the backpack
	MaxWeight     float64                 // Max carry weight of the mobile
	ViewRange     int                     // Range at which items are reported to the client, valid values are [5-18]
	StandingOn    uo.CommonObject         // Object the mobile is standing on
	opl           *serverpacket.OPLPacket // Cached OPLPacket
	oplInfo       *serverpacket.OPLInfo   // Cached OPLInfo packet
}

// Write writes the persistent data of the item to w.
func (m *Mobile) Write(w io.Writer) {
	m.Object.Write(w)
	util.PutUInt32(w, 0)                          // Version
	util.PutBool(w, m.Female)                     // Female flag
	util.PutBool(w, m.Player)                     // Player flag
	util.PutUInt16(w, uint16(m.BaseStrength))     // Strength
	util.PutUInt16(w, uint16(m.BaseDexterity))    // Dexterity
	util.PutUInt16(w, uint16(m.BaseIntelligence)) // Intelligence
	util.PutUInt16(w, uint16(m.MaxHits))          // Max hits
	util.PutUInt16(w, uint16(m.MaxMana))          // Max mana
	util.PutUInt16(w, uint16(m.MaxStamina))       // Max stamina
	for _, v := range m.BaseSkills {              // Skills
		util.PutUInt16(w, uint16(v))
	}
	for _, e := range m.Equipment { // Equipment
		if e == nil {
			util.PutBool(w, false)
			continue
		}
		util.PutBool(w, true)
		e.Write(w)
	}
}

// Read reads the persistent data of the mobile from r.
func (m *Mobile) Read(r io.Reader) {
	m.Object.Read(r)
	_ = util.GetUInt32(r)                       // Version
	m.Female = util.GetBool(r)                  // Female flag
	m.Player = util.GetBool(r)                  // Player flag
	m.BaseStrength = int(util.GetUInt16(r))     // Strength
	m.BaseDexterity = int(util.GetUInt16(r))    // Dexterity
	m.BaseIntelligence = int(util.GetUInt16(r)) // Intelligence
	m.MaxHits = int(util.GetUInt16(r))          // Max hits
	m.MaxMana = int(util.GetUInt16(r))          // Max mana
	m.MaxStamina = int(util.GetUInt16(r))       // Max stamina
	for i := range m.BaseSkills {               // Skills
		m.BaseSkills[i] = int(util.GetUInt16(r))
	}
	for layer := range m.Equipment { // Equipment
		if util.GetBool(r) {
			i := &Item{}
			i.Read(r)
			m.Equipment[layer] = i
		}
	}
	// Establish sane defaults for variables that need non-zero default values
	m.ViewRange = uo.MaxViewRange
}

// RecalculateStats recalculates all internal cache states.
func (m *Mobile) RecalculateStats() {
	m.Strength = m.BaseStrength
	m.Dexterity = m.BaseDexterity
	m.Intelligence = m.BaseIntelligence
	if m.Player {
		m.MaxHits = m.Strength/2 + 50
		m.MaxMana = m.Intelligence
		m.MaxStamina = m.Dexterity
		m.MaxWeight = float64(int(float64(m.Strength)*3.5 + 40))
	}
	m.Hits = m.MaxHits
	m.Mana = m.MaxMana
	m.Stamina = m.MaxStamina
	m.Skills = m.BaseSkills // Note to self, this does an array copy
	m.Weight = 0
	for layer, e := range m.Equipment {
		if layer < int(uo.LayerFirstValid) || layer == int(uo.LayerBankBox) {
			continue
		}
		if layer > int(uo.LayerLastValid) {
			break
		}
		m.Weight += e.Weight
		if e.HasFlags(ItemFlagsContainer) {
			e.RecalculateStats()
			m.Weight += e.ContainedWeight
		}
	}
}

// OPLPackets constructs new OPL packets if needed and returns cached packets.
func (m *Mobile) OPLPackets() (*serverpacket.OPLPacket, *serverpacket.OPLInfo) {
	if m.opl == nil {
		m.opl = &serverpacket.OPLPacket{
			Serial: m.Serial,
		}
		// Base mobil eproperties
		m.opl.AppendColor(colornames.White, m.DisplayName(), false)
		m.opl.Compile()
		m.oplInfo = &serverpacket.OPLInfo{
			Serial: m.Serial,
			Hash:   m.opl.Hash,
		}
	}
	return m.opl, m.oplInfo
}

// Flags returns the compile [uo.MobileFlags] value for this mobile.
func (m *Mobile) Flags() uo.MobileFlags {
	ret := uo.MobileFlagNone
	if m.Female {
		ret |= uo.MobileFlagFemale
	}
	return ret
}

// NotorietyFor returns the [uo.Notoriety] value for mobile other from the
// perspective of this mobile.
func (m *Mobile) NotorietyFor(other *Mobile) uo.Notoriety {
	if other.Player {
		return uo.NotorietyInnocent
	}
	return m.BaseNotoriety
}

// CanSee returns true if the object can be seen by this mobile. This function
// *does not* test for line of sight.
func (m *Mobile) CanSee(o *Object) bool {
	switch o.Visibility {
	case uo.VisibilityVisible:
		return true
	case uo.VisibilityInvisible:
		return false
	case uo.VisibilityHidden:
		return false
	case uo.VisibilityStaff:
		if m.Account == nil {
			return false
		}
		return m.Account.HasRole(RoleGameMaster)
	case uo.VisibilityNone:
		return false
	}
	return false
}

// AfterUnmarshalOntoMap is called after all of the tiles, statics, items and
// mobiles have been loaded and placed on the map. It updates internal states
// that are required for proper movement and control.
func (m *Mobile) AfterUnmarshalOntoMap() {
}

// ContextMenu returns a new context menu packet.
func (m *Mobile) ContextMenu(p *ContextMenu, mob *Mobile) {
}

// AfterMove handles things that happen every time a mobile steps such as
// stamina decay.
func (m *Mobile) AfterMove() {
	// Max weight checks
	w := int(m.Weight)
	mw := int(m.MaxWeight)
	if w > mw {
		sc := w - mw
		m.Stamina -= sc
		if m.Stamina < 0 {
			m.Stamina = 0
		}
		World.UpdateMobile(m)
	}
	// Check for containers that we need to close
	if m.NetState != nil {
		m.NetState.ContainerRangeCheck()
	}
}
