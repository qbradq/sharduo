package game

// Location identifies an absolute location in the universe.
type Location struct {
	// Absolute X position on the map [0-)
	X int
	// Absolute Y position on the map
	Y int
	// Absolute Z position on the map
	Z int
}

// WrapAndBound wraps and bounds the spacial portion of the location to the
// map dimensions.
func (l Location) WrapAndBound() Location {
	ret := l
	for {
		if ret.X < 0 {
			ret.X += MapWidth
		} else if l.X >= MapWidth {
			ret.X -= MapWidth
		} else {
			break
		}
	}
	for {
		if ret.Y < 0 {
			ret.Y += MapHeight
		} else if l.Y >= MapHeight {
			ret.Y -= MapHeight
		} else {
			break
		}
	}
	if ret.Z < -127 {
		ret.Z = -127
	} else if ret.Z >= 128 {
		ret.Z = 127
	}
	return ret
}
