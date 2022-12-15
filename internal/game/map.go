package game

import (
	"log"

	"github.com/qbradq/sharduo/lib/uo"
)

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
	chunks []*chunk
}

// NewMap creates and returns a new Map
func NewMap() *Map {
	m := &Map{
		chunks: make([]*chunk, MapChunksWidth*MapChunksHeight),
	}

	for cx := 0; cx < MapChunksWidth; cx++ {
		for cy := 0; cy < MapChunksHeight; cy++ {
			m.chunks[cy*MapChunksWidth+cx] = newChunk(cx*ChunkWidth, cy*ChunkHeight)
		}
	}
	return m
}

// getChunk returns a pointer to the chunk for the given location.
func (m *Map) getChunk(l Location) *chunk {
	l = l.WrapAndBound(l)
	cx := l.X / ChunkWidth
	cy := l.Y / ChunkHeight
	return m.chunks[cy*MapChunksWidth+cx]
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
		log.Println(len(mobs))
		for _, other := range mobs {
			log.Println(other)
			if o == other {
				continue
			}
			if item, ok := other.(Item); ok {
				mob.NetState().SendItem(item)
			}
		}
	}
}

// MoveObject moves an object in the given direction. Returns true if the
// movement was successfull.
func (m *Map) MoveObject(o Object, dir uo.Direction) bool {
	// Change facing request
	dir = dir.Bound()
	if o.Facing() != dir {
		o.SetFacing(dir)	
		return true
	}
	// Movement request
	oldLocation := o.Location()
	newLocation := o.Location().Forward(dir)	
	return false
}

// getChunksInBounds returns a slice of all the chunks within a given bounds.
func (m *Map) getChunksInBounds(b Bounds) []*chunk {
	var ret []*chunk
	l := Location{}
	for l.Y = b.Y; l.Y < b.Y+b.H; l.Y += ChunkHeight {
		for l.X = b.X; l.X < b.X+b.W; l.X += ChunkWidth {
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
