package uo

import "math/rand"

// A Hue is a 17-bit value that describes the rendering mode of an object.
// Hues have the following characteristics:
// The zero value means "default rendering mode"
// Values 1-3000 inclusive select a set of 16 colors from the file "hues.mul"
//   that replace the first 16 color indexes (the gray-scales).
// The special value -1 (0xffff) will do the shadow dragon alpha effect.
type Hue uint16

// Important hue values
const (
	HueDefault Hue = 0
	HueMin     Hue = 1
	HueBlack   Hue = 1
	HueDyeMin  Hue = 2
	HueDyeMax  Hue = 1001
	HueSkinMin Hue = 1002
	HueSkinMax Hue = 1058
	HueMax     Hue = 3000
	HueAlpha   Hue = 0xffff
)

// RandomSkinHue returns a random skin hue
func RandomSkinHue() Hue {
	return Hue(rand.Intn(int(HueSkinMax-HueSkinMin))) + HueSkinMin
}

// RandomDyeHue returns a random normal dye hue
func RandomDyeHue() Hue {
	return Hue(rand.Intn(int(HueDyeMax-HueDyeMin))) + HueDyeMin
}
