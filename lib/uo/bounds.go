package uo

// Bounds represents a 3D bounding box in the world.
type Bounds struct {
	// X location of the top-left corner
	X int
	// Y location of the top-left corner
	Y int
	// Z location of the floor
	Z int
	// Width of the bounds (X-axis)
	W int
	// Height of the bounds (Y-axis)
	H int
	// Depth of the bounds (Z-axis)
	D int
}

// Contains returns true if the location is contained within these bounds.
func (b *Bounds) Contains(l Location) bool {
	return l.X >= b.X && l.X < b.X+b.W && l.Y >= b.Y && l.Y < b.Y+b.H && l.Z >= b.Z && l.Z < b.Z+b.D
}
