package uo

import (
	"math"
)

// Point identifies an absolute location in the universe.
type Point struct {
	// Absolute X position on the map [0-)
	X int
	// Absolute Y position on the map
	Y int
	// Absolute Z position on the map
	Z int
}

// This Location value indicates to a container that the item should be placed
// at a random location.
var RandomContainerLocation Point = Point{X: RandomDropX, Y: RandomDropY}

// WrapToOverworld returns the location wrapped to the overworld portion of the
// map.
func (l Point) WrapToOverworld() Point {
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
func (l Point) WrapToDungeonServer() Point {
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
func (l Point) WrapAndBound(ref Point) Point {
	ref = ref.Bound()
	if ref.X < MapOverworldWidth {
		return l.WrapToOverworld()
	} else {
		return l.WrapToDungeonServer()
	}
}

// Bound bounds the spacial portion of the location to the absolute dimensions
// of the map.
func (l Point) Bound() Point {
	for l.X < 0 {
		l.X += MapWidth
	}
	for l.X >= MapWidth {
		l.X -= MapWidth
	}
	for l.Y < 0 {
		l.Y += MapHeight
	}
	for l.Y >= MapHeight {
		l.Y -= MapHeight
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
func (l Point) XYDistance(d Point) int {
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
func (l Point) Forward(d Direction) Point {
	d = d & 0x07
	return Point{
		X: l.X + dirOfs[d][0],
		Y: l.Y + dirOfs[d][1],
		Z: l.Z,
	}
}

// DirectionTo returns the direction code that most closely matches the
// direction of the argument location.
func (l Point) DirectionTo(a Point) Direction {
	r := math.Atan2(float64(a.X-l.X), float64(a.Y-l.Y)) * 180 / math.Pi
	b := -157.5
	if r < b+45*0 {
		return DirectionNorth
	}
	if r < b+45*1 {
		return DirectionNorthWest
	}
	if r < b+45*2 {
		return DirectionWest
	}
	if r < b+45*3 {
		return DirectionSouthWest
	}
	if r < b+45*4 {
		return DirectionSouth
	}
	if r < b+45*5 {
		return DirectionSouthEast
	}
	if r < b+45*6 {
		return DirectionEast
	}
	if r < b+45*7 {
		return DirectionNorthEast
	}
	return DirectionNorth
}

// BoundsByRadius returns a Bounds value that contains this location and the
// locations within the radius r. The Z portion of the bounds will be maximized.
func (l Point) BoundsByRadius(r int) Bounds {
	return Bounds{
		X: l.X - r,
		Y: l.Y - r,
		Z: MapMinZ,
		W: r*2 + 1,
		H: r*2 + 1,
		D: MapMaxZ - MapMinZ,
	}
}

// ChunkBound returns a Point value that is properly wrapped and bounded.
func (l Point) ChunkBound(r Point) Point {
	l.X *= ChunkWidth
	l.Y *= ChunkHeight
	l = l.WrapAndBound(r)
	l.X /= ChunkWidth
	l.Y /= ChunkHeight
	return l
}
