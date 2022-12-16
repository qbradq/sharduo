package game

import (
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// chunk contains the tile matrix, static and dynamic objects in one 8x8 chunk
type chunk struct {
	// Bounds of the chunk
	bounds uo.Bounds
	// The slice of all tiles in the chunk
	tiles []uo.Tile
	// Collection of all of the objects in the chunk
	objects util.Slice[Object]
}

// newChunk creates and returns a new Chunk object
func newChunk(x, y int) *chunk {
	return &chunk{
		bounds: uo.Bounds{
			X: x,
			Y: y,
			Z: uo.MapMinZ,
			W: uo.ChunkWidth,
			H: uo.ChunkHeight,
			D: uo.MapMaxZ - uo.MapMinZ,
		},
		tiles: make([]uo.Tile, uo.ChunkWidth*uo.ChunkHeight),
	}
}

// Add adds the object to the chunk and returns true if it is located in this
// chunk.
func (c *chunk) Add(o Object) bool {
	if !c.bounds.Contains(o.Location()) {
		return false
	}
	c.objects = c.objects.Append(o)
	return true
}

// Remove removes the object from the chunk.
func (c *chunk) Remove(o Object) {
	c.objects = c.objects.Remove(o)
}

// GetTile reutrns the Tile value for the given chunk-relative location. x and y
// must be between 0 and 7 inclusive.
func (c *chunk) GetTile(x, y int) uo.Tile {
	return c.tiles[(y%uo.ChunkHeight)*uo.ChunkWidth+(x%uo.ChunkWidth)]
}

// setTile sets the tile value at the given chunk-relative location. x and y
// must be between 0 and 7 inclusive.
func (c *chunk) setTile(x, y int, t uo.Tile) {
	c.tiles[(y%uo.ChunkHeight)*uo.ChunkWidth+(x%uo.ChunkWidth)] = t
}
