package uo

var BoundsZero = Bounds{} // Bounds zero value

// Bounds represents a 3D bounding box in the world.
type Bounds struct {
	// X location of the top-left corner
	X int16
	// Y location of the top-left corner
	Y int16
	// Z location of the floor
	Z int8
	// Width of the bounds (X-axis)
	W int16
	// Height of the bounds (Y-axis)
	H int16
	// Depth of the bounds (Z-axis)
	D int16
}

// BoundsOf returns a bounds value that fits both locations tightly. This can
// be used to create a bounds value from a start and end position.
func BoundsOf(s, e Location) Bounds {
	var ret Bounds
	if s.X < e.X {
		ret.X = s.X
		ret.W = e.X - s.X + 1
	} else {
		ret.X = e.X
		ret.W = s.X - e.X + 1
	}
	if s.Y < e.Y {
		ret.Y = s.Y
		ret.H = e.Y - s.Y + 1
	} else {
		ret.Y = e.Y
		ret.H = s.Y - e.Y + 1
	}
	if s.Z < e.Z {
		ret.Z = s.Z
		ret.D = int16(e.Z) - int16(s.Z)
	} else {
		ret.Z = e.Z
		ret.D = int16(s.Z) - int16(e.Z) + 1
	}
	return ret
}

// Contains returns true if the location is contained within these bounds.
func (b *Bounds) Contains(l Location) bool {
	return l.X >= b.X && l.X < b.X+b.W && l.Y >= b.Y && l.Y < b.Y+b.H && l.Z >= b.Z && l.Z < int8(int16(b.Z)+b.D)
}

// East returns the east-most point within these bounds.
func (b *Bounds) East() int16 { return b.X + b.W - 1 }

// South returns the south-most point within these bounds.
func (b *Bounds) South() int16 { return b.Y + b.H - 1 }

// Top returns the top-most point within these bounds.
func (b *Bounds) Top() int8 { return int8(int(b.Z) + int(b.D) - 1) }
