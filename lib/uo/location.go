package uo

// Location identifies an absolute location in the universe.
type Location struct {
	// Absolute X position on the map [0-)
	X int16
	// Absolute Y position on the map
	Y int16
	// Absolute Z position on the map
	Z int8
}

// This Location value indicates to a container that the item should be placed
// at a random location.
var RandomContainerLocation Location = Location{X: RandomDropX, Y: RandomDropY}

// WrapToOverworld returns the location wrapped to the overworld portion of the
// map.
func (l Location) WrapToOverworld() Location {
	for l.X < 0 {
		l.X += int16(MapOverworldWidth)
	}
	for l.X >= int16(MapOverworldWidth) {
		l.X -= int16(MapOverworldWidth)
	}
	for l.Y < 0 {
		l.Y += int16(MapHeight)
	}
	for l.Y >= int16(MapHeight) {
		l.Y -= int16(MapHeight)
	}
	return l
}

// WrapToDungeonServer returns the location wrapped to the dungeon server
// section of the map.
func (l Location) WrapToDungeonServer() Location {
	for l.X < int16(MapOverworldWidth) {
		l.X += int16(MapWidth) - int16(MapOverworldWidth)
	}
	for l.X > int16(MapWidth) {
		l.X -= int16(MapWidth) - int16(MapOverworldWidth)
	}
	for l.Y < 0 {
		l.Y += int16(MapHeight)
	}
	for l.Y >= int16(MapHeight) {
		l.Y -= int16(MapHeight)
	}
	return l
}

// WrapAndBound wraps and bounds the spacial portion of the location to the
// map dimensions relative to a reference point. UpdateAndBound will handle map
// wrapping as appropriate based on the reference location.
func (l Location) WrapAndBound(ref Location) Location {
	ref = ref.Bound()
	if ref.X < int16(MapOverworldWidth) {
		return l.WrapToOverworld()
	} else {
		return l.WrapToDungeonServer()
	}
}

// Bound bounds the spacial portion of the location to the absolute dimensions
// of the map.
func (l Location) Bound() Location {
	for l.X < 0 {
		l.X += int16(MapWidth)
	}
	for l.X > int16(MapWidth) {
		l.X -= int16(MapWidth)
	}
	for l.Y < 0 {
		l.Y += int16(MapHeight)
	}
	for l.Y >= int16(MapHeight) {
		l.Y -= int16(MapHeight)
	}
	if l.Z < MapMinZ {
		l.Z = MapMinZ
	}
	if l.Z > MapMaxZ {
		l.Z = MapMaxZ
	}
	return l
}

// XYDistance returns the maximum distance from l to d along either the X or Y
// axis.
func (l Location) XYDistance(d Location) int16 {
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

// Forward moves the location in the given direction, affecting only the X and
// Y coordinates.
func (l Location) Forward(d Direction) Location {
	d = d & 0x07
	return Location{
		X: l.X + dirOfs[d][0],
		Y: l.Y + dirOfs[d][1],
		Z: l.Z,
	}
}
