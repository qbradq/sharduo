package game

import (
	"github.com/qbradq/sharduo/lib/uo"
)

// Map constants
const ()

// Map contains the tile matrix, static items, and all dynamic objects of a map.
type Map struct {
	// The chunks of the map
	chunks []*chunk
}

// NewMap creates and returns a new Map
func NewMap() *Map {
	m := &Map{
		chunks: make([]*chunk, uo.MapChunksWidth*uo.MapChunksHeight),
	}

	for cx := 0; cx < uo.MapChunksWidth; cx++ {
		for cy := 0; cy < uo.MapChunksHeight; cy++ {
			m.chunks[cy*uo.MapChunksWidth+cx] = newChunk(cx*uo.ChunkWidth, cy*uo.ChunkHeight)
		}
	}
	return m
}

// getChunk returns a pointer to the chunk for the given location.
func (m *Map) getChunk(l Location) *chunk {
	l = l.WrapAndBound(l)
	cx := l.X / uo.ChunkWidth
	cy := l.Y / uo.ChunkHeight
	return m.chunks[cy*uo.MapChunksWidth+cx]
}

// AddNewObject adds a new object to the map at the given location
func (m *Map) AddNewObject(o Object) {
	c := m.getChunk(o.Location())
	c.Add(o)

	// Send the item to all mobiles in range with an attached NetState
	for _, other := range m.GetObjectsInRange(o.Location(), uo.MaxViewRange) {
		if o == other {
			continue
		}
		m, ok := other.(Mobile)
		if !ok {
			continue
		}
		if m.NetState() != nil {
			if item, ok := o.(Item); ok {
				m.NetState().SendItem(item)
			}
		}
	}

	// If this is a mobile with a NetState we have to send all of the items
	// and mobiles in range.
	mob, ok := o.(Mobile)
	if ok && mob.NetState() != nil {
		mobs := m.GetObjectsInRange(mob.Location(), uo.MaxViewRange)
		for _, other := range mobs {
			if o == other {
				continue
			}
			if item, ok := other.(Item); ok {
				mob.NetState().SendItem(item)
			}
		}
	}
}

// getChunksInBounds returns a slice of all the chunks within a given bounds.
func (m *Map) getChunksInBounds(b Bounds) []*chunk {
	var ret []*chunk
	l := Location{}
	for l.Y = b.Y; l.Y < b.Y+b.H; l.Y += uo.ChunkHeight {
		for l.X = b.X; l.X < b.X+b.W; l.X += uo.ChunkWidth {
			ret = append(ret, m.getChunk(l))
		}
	}
	return ret
}

// getChunksInRange gets chunks in the given range of a reference point.
func (m *Map) getChunksInRange(l Location, r int) []*chunk {
	return m.getChunksInBounds(Bounds{
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
	for _, c := range m.getChunksInRange(l, r) {
		for _, o := range c.objects {
			d := l.XYDistance(o.Location())
			if d > r {
				continue
			}
			ret = append(ret, o)
		}
	}
	return ret
}
