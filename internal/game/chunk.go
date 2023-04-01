package game

import (
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// Chunk contains the tile matrix, static and dynamic objects in one 8x8 Chunk
type Chunk struct {
	// Bounds of the chunk
	bounds uo.Bounds
	// The slice of all tiles in the chunk
	tiles []uo.Tile
	// Collection of all statics in the chunk
	statics []uo.Static
	// Collection of all items in the chunk
	items util.Slice[Item]
	// Collection of all mobiles in the chunk
	mobiles util.Slice[Mobile]
	// Amount of ore left in this chunk
	ore uint8
	// oreDeadline is the time when this chunk should regenerate its ore
	oreDeadline uo.Time
}

// newChunk creates and returns a new Chunk object
func newChunk(x, y int) *Chunk {
	return &Chunk{
		bounds: uo.Bounds{
			X: int16(x),
			Y: int16(y),
			Z: uo.MapMinZ,
			W: int16(uo.ChunkWidth),
			H: int16(uo.ChunkHeight),
			D: int16(uo.MapMaxZ) - int16(uo.MapMinZ),
		},
		tiles: make([]uo.Tile, uo.ChunkWidth*uo.ChunkHeight),
	}
}

// Add adds the object to the chunk and returns true if it is located in this
// chunk.
func (c *Chunk) Add(o Object) bool {
	if !c.bounds.Contains(o.Location()) {
		return false
	}
	if item, ok := o.(Item); ok {
		// Keep items Z-sorted
		insertPoint := len(c.items)
		for idx, otherItem := range c.items {
			if item.Z() < otherItem.Z() {
				insertPoint = idx
			} else {
				break
			}
		}
		c.items = c.items.Insert(insertPoint, item)
	} else if mobile, ok := o.(Mobile); ok {
		c.mobiles = c.mobiles.Append(mobile)
	} else {
		// Unknown object interface
		return false
	}
	return true
}

// Remove removes the object from the chunk.
func (c *Chunk) Remove(o Object) {
	if item, ok := o.(Item); ok {
		c.items = c.items.Remove(item)
	} else if mobile, ok := o.(Mobile); ok {
		c.mobiles = c.mobiles.Remove(mobile)
	}
}

// GetTile returns the Tile value for the given chunk-relative location. x and y
// must be between 0 and 7 inclusive.
func (c *Chunk) GetTile(x, y int16) uo.Tile {
	return c.tiles[(int(y)%uo.ChunkHeight)*uo.ChunkWidth+(int(x)%uo.ChunkWidth)]
}

// setTile sets the tile value at the given chunk-relative location. x and y
// must be between 0 and 7 inclusive.
func (c *Chunk) setTile(x, y int, t uo.Tile) {
	c.tiles[(y%uo.ChunkHeight)*uo.ChunkWidth+(x%uo.ChunkWidth)] = t
}

// Update handles the 1-minute periodic update for this chunk.
func (c *Chunk) Update(t uo.Time) {
	// Ore respawn
	if t >= c.oreDeadline {
		c.ore = uint8(world.Random().Random(10, 24))
		c.oreDeadline = t + uo.DurationMinute*30
	}
	// Item updates for items on the ground
	for _, i := range c.items {
		i.Update(t)
	}
}
