package common

import "math/rand"

// A Serial is a 31-bit value with the following characteristics:
// The zero value is also the "invalid value" value
// No Serial will have a value greater than 2^31-1
// A Serial can always be cast to a uint32 without data loss
type Serial int32

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

// A Hue is a 16-bit value that describes the rendering mode of an object.
// Hues have the following characteristics:
// The zero value means "default rendering mode"
// Values 1-3000 inclusive select a set of 16 colors from the file "hues.mul"
//   that replace the first 16 color indicies (the grayscales).
// The special value -1 (0xffff) will do the shadow dragon alpha effect.
type Hue uint16

// Important hue values
const (
	HueDefault uint16 = 0
	HueMin     uint16 = 1
	HueBlack   uint16 = 1
	HueDieMin  uint16 = 2
	HueDieMax  uint16 = 1001
	HueSkinMin uint16 = 1002
	HueSkinMax uint16 = 1058
	HueMax     uint16 = 3000
)

// RandomSkinHue returns a random skin hue
func RandomSkinHue() Hue {
	return Hue(uint16(rand.Intn(int(HueSkinMax-HueSkinMin))) + HueSkinMin)
}
