package uo

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

// Contains returns true if the location is contained within these bounds.
func (b *Bounds) Contains(l Location) bool {
	return l.X >= b.X && l.X < b.X+b.W && l.Y >= b.Y && l.Y < b.Y+b.H && l.Z >= b.Z && l.Z < int8(int16(b.Z)+b.D)
}
