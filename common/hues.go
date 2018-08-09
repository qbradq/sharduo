package common

import "math/rand"

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
