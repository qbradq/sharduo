package game

// Chunk constants
const (
	ChunkWidth  int = 16
	ChunkHeight int = 16
)

// Chunk contains the tile matrix, static and dynamic objects in one 8x8 chunk
type Chunk struct {
	// Bounds of the chunk
	bounds Bounds
	// Collection of all of the objects in the chunk
	objects ObjectCollection
}

// NewChunk creates and returns a new Chunk object
func NewChunk(x, y int) *Chunk {
	return &Chunk{
		bounds: Bounds{
			X: x,
			Y: y,
			Z: MapMinZ,
			W: ChunkWidth,
			H: ChunkHeight,
			D: MapMaxZ - MapMinZ,
		},
	}
}

// Add adds the object to the chunk and returns true if it is located in this
// chunk.
func (c *Chunk) Add(o Object) bool {
	if !c.bounds.Contains(o.GetLocation()) {
		return false
	}
	c.objects = c.objects.Append(o)
	return true
}

// Remove removes the object from the chunk.
func (c *Chunk) Remove(o Object) {
	c.objects = c.objects.Remove(o)
}
