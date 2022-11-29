package game

// Map constants
const (
	MapWidth        int = 7168
	MapHeight       int = 4096
	MapChunksWidth  int = MapWidth / ChunkWidth
	MapChunksHeight int = MapHeight / ChunkHeight
)

// Map contains the tile matrix, static items, and all dynamic objects of a map.
type Map struct {
	// The chunks of the map
	Chunks []*Chunk
}

// NewMap creates and returns a new Map
func NewMap() *Map {
	return &Map{
		Chunks: make([]*Chunk, MapChunksWidth*MapChunksHeight),
	}
}

// GetChunk returns a pointer to the chunk for the given location.
func (m *Map) GetChunk(l Location) *Chunk {
	l = l.WrapAndBound()
	cx := l.X / ChunkWidth
	cy := l.Y / ChunkHeight
	return m.Chunks[cy*MapChunksWidth+cx]
}

// AddNewObject adds a new object to the map.
func (m *Map) AddNewObject(o Object) {
	c := m.GetChunk(o.GetLocation())
	c.MoveInto(o)
}
