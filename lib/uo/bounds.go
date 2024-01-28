package uo

var BoundsZero = Bounds{} // Bounds zero value

var BoundsFullMap = Bounds{
	Z: MapMinZ,
	W: MapWidth,
	H: MapHeight,
	D: MapMaxZ - MapMinZ,
}

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

// BoundsFit returns a bounds value that fits both bounds tightly.
func BoundsFit(a, b Bounds) Bounds {
	var ret Bounds
	if a.X < b.X {
		ret.X = a.X
	} else {
		ret.X = b.X
	}
	if a.Y < b.Y {
		ret.Y = a.Y
	} else {
		ret.Y = b.Y
	}
	if a.Z < b.Z {
		ret.Z = a.Z
	} else {
		ret.Z = b.Z
	}
	if a.East() > b.East() {
		ret.W = a.East() - ret.X + 1
	} else {
		ret.W = b.East() - ret.X + 1
	}
	if a.South() > b.South() {
		ret.H = a.South() - ret.Y + 1
	} else {
		ret.H = b.South() - ret.Y + 1
	}
	if a.Top() > b.Top() {
		ret.D = a.Top() - ret.Z + 1
	} else {
		ret.D = b.Top() - ret.Z + 1
	}
	return ret
}

// BoundsOf returns a bounds value that fits both locations tightly. This can
// be used to create a bounds value from a start and end position.
func BoundsOf(s, e Point) Bounds {
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
		ret.D = e.Z - s.Z
	} else {
		ret.Z = e.Z
		ret.D = s.Z - e.Z + 1
	}
	return ret
}

// RandomLocation returns a random location within these bounds.
func (b Bounds) RandomLocation(rng RandomSource) Point {
	return Point{
		X: rng.Random(b.X, b.East()),
		Y: rng.Random(b.Y, b.South()),
		Z: rng.Random(b.Z, b.Top()),
	}
}

// Contains returns true if the location is contained within these bounds.
func (b Bounds) Contains(l Point) bool {
	return l.X >= b.X && l.X <= b.East() && l.Y >= b.Y && l.Y <= b.South() && l.Z >= b.Z && l.Z <= b.Top()
}

// Overlaps returns true if the two bound values overlap
func (b Bounds) Overlaps(a Bounds) bool {
	return !(a.South() < b.Y || b.South() < a.Y || a.East() < b.X || b.East() < a.X)
}

// East returns the east-most point within these bounds.
func (b Bounds) East() int { return b.X + b.W - 1 }

// South returns the south-most point within these bounds.
func (b Bounds) South() int { return b.Y + b.H - 1 }

// Top returns the top-most point within these bounds.
func (b Bounds) Top() int { return b.Z + b.D - 1 }
