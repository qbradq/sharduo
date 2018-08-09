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
