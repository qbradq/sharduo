package common

// A Role represents a single permission domain
type Role uint32

// Common role names
const (
	RoleNone          Role = 0x00000000
	RoleAuthenticated Role = 0x00000001
)

// HasAll returns true only if r contains all roles in v
func (r Role) HasAll(v Role) bool {
	return r&v == v
}

// HasAny returns true if r contains any roles in v
func (r Role) HasAny(v Role) bool {
	return r&v != 0
}

// ClientFlag represents the client features flags sent in packet 0xA9
type ClientFlag uint32

// All documented flags
const (
	ClientFlagNone                 ClientFlag = 0x00000000
	ClientFlagSiege                ClientFlag = 0x00000004
	ClientFlagLeftClickMenus       ClientFlag = 0x00000008
	ClientFlagAOS                  ClientFlag = 0x00000020
	ClientFlagSixthCharacterSlot   ClientFlag = 0x00000040
	ClientFlagAOSProfessions       ClientFlag = 0x00000080
	ClientFlagElvenRace            ClientFlag = 0x00000100
	ClientFlagSeventhCharacterSlot ClientFlag = 0x00001000
	ClientFlagNewMovementPackets   ClientFlag = 0x00004000
	ClientFlagNewFeluccaAreas      ClientFlag = 0x00008000
)
