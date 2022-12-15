package file

import (
	"image/color"

	"github.com/qbradq/sharduo/lib/dataconv"
)

// RadarColMul represents "radarcol.mul" in the client files
type RadarColMul struct {
	// Array of 0x8000 radar colors
	colors []color.RGBA
}

// NewRadarColMulFromFile loads "radarcol.mul" from the named file
func NewRadarColMulFromFile(fname string) *RadarColMul {
	m := NewStaticMulFromFile(fname, 2, 0)
	if m == nil {
		return nil
	}
	ret := &RadarColMul{
		colors: make([]color.RGBA, m.NumberOfSegments()),
	}
	for i := range ret.colors {
		c := dataconv.GetUint16(m.GetSegment(i))
		r := (c & 0b0111110000000000) >> 7
		// r |= r >> 5
		g := (c & 0b0000001111100000) >> 2
		// g |= g >> 5
		b := (c & 0b0000000000011111) << 3
		// b |= b >> 5
		ret.colors[i] = color.RGBA{
			R: uint8(r),
			G: uint8(g),
			B: uint8(b),
			A: uint8(0xFF),
		}
	}
	return ret
}

// Colors returns a slice of all of the colors in the file
func (m *RadarColMul) Colors() []color.RGBA {
	return m.colors
}
