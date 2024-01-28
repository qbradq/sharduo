package game

import (
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"golang.org/x/image/colornames"
)

// Mobile describes a thinking actor.
type Mobile struct {
	Object
	// Static values
	Body          uo.Body      // Body to use for this mobile
	Player        bool         // If true this is a player's character
	BaseNotoriety uo.Notoriety // Base notoriety level for this mobile
	// Persistent values
	Female           bool                 // If true the mobile is female
	Hits             int                  // Current hit points
	MaxHits          int                  // Max hit points
	Mana             int                  // Current mana
	MaxMana          int                  // Max mana
	Stamina          int                  // Current stamina
	MaxStamina       int                  // Maximum stamina
	BaseStrength     int                  // Base strength value
	BaseDexterity    int                  // Base dexterity value
	BaseIntelligence int                  // Base intelligence value
	BaseSkills       [uo.SkillCount]int   // Base skill values in tenths of a percent
	Equipment        [uo.LayerCount]*Item // Current equipment set
	// Transient values
	Account       *Account                // Connected account if any
	NetState      NetState                // Connected net state if any
	ControlMaster *Mobile                 // Mobile that is currently commanding this mobile
	Skills        [uo.SkillCount]int      // Current skill values in tenths of a percent
	Strength      int                     // Current strength value
	Dexterity     int                     // Current dexterity value
	Intelligence  int                     // Current intelligence value
	Running       bool                    // If true the mobile is running
	Weight        float64                 // Current weight of the mobile including all equipment excluding special equipment like mount items and the bank box
	MaxWeight     float64                 // Max carry weight of the mobile
	opl           *serverpacket.OPLPacket // Cached OPLPacket
	oplInfo       *serverpacket.OPLInfo   // Cached OPLInfo packet
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
