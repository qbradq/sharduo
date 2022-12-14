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

// WrapToOverworld returns the location wrapped to the overworld portion of the
// map.
func (l Location) WrapToOverworld() Location {
	for l.X < 0 {
		l.X += MapOverworldWidth
	}
	for l.X >= MapOverworldWidth {
		l.X -= MapOverworldWidth
	}
	for l.Y < 0 {
		l.Y += MapHeight
	}
	for l.Y >= MapHeight {
		l.Y -= MapHeight
	}
	return l
}

// WrapToDungeonServer returns the location wrapped to the dungeon server
// section of the map.
func (l Location) WrapToDungeonServer() Location {
	for l.X < MapOverworldWidth {
		l.X += MapWidth - MapOverworldWidth
	}
	for l.X > MapWidth {
		l.X -= MapWidth - MapOverworldWidth
	}
	for l.Y < 0 {
		l.Y += MapHeight
	}
	for l.Y >= MapHeight {
		l.Y -= MapHeight
	}
	return l
}

// WrapAndBound wraps and bounds the spacial portion of the location to the
// map dimensions relative to a reference point. UpdateAndBound will handle map
// wrapping as appropriate based on the reference location.
func (l Location) WrapAndBound(ref Location) Location {
	if ref.X < MapOverworldWidth {
		return l.WrapToOverworld()
	} else {
		return l.WrapToDungeonServer()
	}
}

// XYDistance returns the maximum distance from l to d along either the X or Y
// axis.
func (l Location) XYDistance(d Location) int {
	dx := l.X - d.X
	dy := l.Y - d.Y
	if dx < 0 {
		dx = dx * -1
	}
	if dy < 0 {
		dy = dy * -1
	}
	if dx > dy {
		return dx
	}
	return dy
}
