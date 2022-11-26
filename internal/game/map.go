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
