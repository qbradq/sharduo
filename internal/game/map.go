package game

// Map constants
const (
	MapWidth          int = 7168
	MapHeight         int = 4096
	MapOverworldWidth int = MapHeight
	MapChunksWidth    int = MapWidth / ChunkWidth
	MapChunksHeight   int = MapHeight / ChunkHeight
	MapMinZ           int = -127
	MapMaxZ           int = 128
)

// Map contains the tile matrix, static items, and all dynamic objects of a map.
type Map struct {
	// The chunks of the map
	chunks []*Chunk
}

// NewMap creates and returns a new Map
func NewMap() *Map {
	m := &Map{
		chunks: make([]*Chunk, MapChunksWidth*MapChunksHeight),
	}

	for cx := 0; cx < MapChunksWidth; cx++ {
		for cy := 0; cy < MapChunksHeight; cy++ {
			m.chunks[cy*MapChunksWidth+cx] = NewChunk(cx*ChunkWidth, cy*ChunkHeight)
		}
	}
	return m
}

// GetChunk returns a pointer to the chunk for the given location.
func (m *Map) GetChunk(l Location) *Chunk {
	l = l.WrapAndBound(l)
	cx := l.X / ChunkWidth
	cy := l.Y / ChunkHeight
	return m.chunks[cy*MapChunksWidth+cx]
}

// AddNewObject adds a new object to the map.
func (m *Map) AddNewObject(o Object) {
	c := m.GetChunk(o.Location())
	c.Add(o)
}

// GetChunksInBounds returns a slice of all the chunks within a given bounds.
func (m *Map) GetChunksInBounds(b Bounds) []*Chunk {
	var ret []*Chunk
	l := Location{}
	for l.Y = b.Y; l.Y < b.Y+b.H; l.Y += ChunkHeight {
		for l.X = b.X; l.X < b.X+b.W; l.X += ChunkWidth {
			ret = append(ret, m.GetChunk(l))
		}
	}
	return ret
}

// GetChunksInRange gets chunks in the given range of a reference point.
func (m *Map) GetChunksInRange(l Location, r int) []*Chunk {
	return m.GetChunksInBounds(Bounds{
		X: l.X - r,
		Y: l.Y - r,
		W: r*2 + 1,
		H: r*2 + 1,
	})
}

// GetObjectsInRange returns a slice of all objects within the given range of
// the given location.
func (m *Map) GetObjectsInRange(l Location, r int) []Object {
	var ret []Object
	for _, c := range m.GetChunksInRange(l, r) {
		for _, o := range c.objects {
			ol := o.Location()
			dx := l.X - ol.X
			if dx < 0 {
				dx *= -1
			}
			if dx > r {
				continue
			}
			dy := l.Y - ol.Y
			if dy < 0 {
				dx *= -1
			}
			if dy > r {
				continue
			}
			ret = append(ret, o)
		}
	}
	return ret
}
