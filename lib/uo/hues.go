package uo

// A Hue is a 17-bit value that describes the rendering mode of an object.
// Hues have the following characteristics:
// The zero value means "default rendering mode"
// Values 1-3000 inclusive select a set of 16 colors from the file "hues.mul"
//   that replace the first 16 color indexes (the gray-scales).
// The special value -1 (0xffff) will do the shadow dragon alpha effect.
type Hue uint16

// Important hue values
const (
	HueDefault     Hue = 0
	HueMin         Hue = 1
	HueBlack       Hue = 1
	HueDyeMin      Hue = 2
	HueDyeMax      Hue = 1001
	HueSkinMin     Hue = 1002
	HueSkinMax     Hue = 1058
	HueIce5        Hue = 1150
	HueIce4        Hue = 1151
	HueIce1        Hue = 1152
	HueIce2        Hue = 1153
	HueIce3        Hue = 1154
	HueMax         Hue = 3000
	HueHidden      Hue = 0x4000
	HueAlpha       Hue = 0xffff
	HuePartialFlag Hue = 0x8000
)

// RandomSkinHue returns a random skin hue
func RandomSkinHue(r RandomSource) Hue {
	return Hue(r.Random(int(HueSkinMin), int(HueSkinMax)))
}

// RandomDyeHue returns a random normal dye hue
func RandomDyeHue(r RandomSource) Hue {
	return Hue(r.Random(int(HueDyeMin), int(HueDyeMax)))
}

// SetPartialHue returns the hue value with the partial hue flag set
func (h Hue) SetPartialHue() Hue { return h | HuePartialFlag }
