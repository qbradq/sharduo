package game

// Chunk constants
const (
	ChunkWidth  int = 8
	ChunkHeight int = 8
)

// Chunk contains the tile matrix, static and dynamic objects in one 8x8 chunk
type Chunk struct {
	// Left X position of the chunk relative to the top-left corner of the world
	X int
	// Top Y position of the chunk relative to the top-left corner of the world
	Y int
	// Collection of all of the objects in the chunk
	Objects map[Object]struct{}
}

// NewChunk creates and returns a new Chunk object
func NewChunk(x, y int) *Chunk {
	return &Chunk{
		X:       x,
		Y:       y,
		Objects: make(map[Object]struct{}),
	}
}

// MoveInto adds the object to the chunk and returns true if it is located in
// this chunk.
func (c *Chunk) MoveInto(o Object) bool {
	ol := o.GetLocation()
	if ol.X < c.X || ol.Y < c.Y || ol.X >= c.X+ChunkWidth || ol.Y >= c.Y+ChunkHeight {
		return false
	}
	c.Objects[o] = struct{}{}
	return true
}

// Remove removes the object from the chunk.
func (c *Chunk) Remove(o Object) {
	delete(c.Objects, o)
}
