package game

import (
	"io"
	"sort"
	"sync"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/uo/file"
	"github.com/qbradq/sharduo/lib/util"
)

// Map holds all of the static data and dynamic objects in the world.
type Map struct {
	chunks []*chunk          // Chunks of the map
	ds     map[uo.Serial]any // Deep storage objects
}

// NewMap returns a new Map with all large memory areas already allocated.
func NewMap() *Map {
	ret := &Map{
		chunks: make([]*chunk, uo.MapChunksWidth*uo.MapChunksHeight),
	}
	for i := range ret.chunks {
		ret.chunks[i] = &chunk{}
	}
	return ret
}

// Update is responsible for all chunk updates, item decay and mobile AI.
func (m *Map) Update(now uo.Time) {
}

// nsRetBuf is the static buffer used for the return value of
// [Map.NetStatesInRange].
var nsRetBuf []*Mobile

// NetStatesInRange returns a slice of all of the mobiles with attached net
// states who's view range is within range of the given point on the map.
// Subsequent calls to NetStatesInRange reuses the same return array.
func (m *Map) NetStatesInRange(p uo.Point) []*Mobile {
	nsRetBuf = nsRetBuf[:]
	return nsRetBuf
}

// Write writes out top-level map information.
func (m *Map) Write(wg *sync.WaitGroup, w io.Writer) {
	defer wg.Done()
}

// Read reads in top-level map information.
func (m *Map) Read(r io.Reader) {
}

// WriteObjects writes out all objects that are directly on the map split into
// pools to facilitate multi-goroutine saving.
func (m *Map) WriteObjects(wg *sync.WaitGroup, w io.Writer, pool, pools int) {
	defer wg.Done()
	l := len(m.chunks)
	util.PutUInt32(w, uint32(l/pools))
	for i := pool; i < l; i += pools {
		c := m.chunks[i]
		for _, i := range c.Items {
			if i.Removed || i.NoRent || i.Spawner != nil {
				// Ignore items we shouldn't persist
				continue
			}
			util.PutBool(w, true)
			i.Write(w)
		}
		util.PutBool(w, false)
		for _, m := range c.Mobiles {
			if m.Removed || m.NoRent || m.Spawner != nil {
				// Ignore mobiles we shouldn't persist
				continue
			}
			util.PutBool(w, true)
			m.Write(w)
		}
		util.PutBool(w, false)
	}
}

// ReadObjects reads in all objects that are directly on the map from the
// reader.
func (m *Map) ReadObjects(r io.Reader, ds *Datastore) {
	n := int(util.GetUInt32(r)) // Number of chunks in the file
	for i := 0; i < n; i++ {
		for util.GetBool(r) {
			item := &Item{}
			item.Read(r)
			ds.InsertItem(item)
		}
		for util.GetBool(r) {
			mob := &Mobile{}
			mob.Read(r)
			ds.InsertMobile(mob)
		}
	}
}

// WriteDeepStorage writes out all objects in deep storage.
func (m *Map) WriteDeepStorage(wg *sync.WaitGroup, w io.Writer) {
	util.PutUInt32(w, uint32(len(m.ds))) // Number of objects
	for _, obj := range m.ds {
		switch o := obj.(type) {
		case *Mobile:
			util.PutBool(w, true) // Is Mobile flag
			o.Write(w)
		case *Item:
			util.PutBool(w, false) // Is Mobile flag
			o.Write(w)
		}
	}
}

// ReadDeepStorage reads in all objects to deep storage.
func (m *Map) ReadDeepStorage(r io.Reader, ds *Datastore) {
	n := int(util.GetUInt32(r)) // Number of objects
	for i := 0; i < n; i++ {
		if util.GetBool(r) {
			mob := &Mobile{}
			mob.Read(r)
			m.ds[mob.Serial] = mob
			ds.InsertMobile(mob)
		} else {
			item := &Item{}
			item.Read(r)
			m.ds[item.Serial] = item
			ds.InsertItem(item)
		}
	}
}

// GetTile returns the tile information at the given location.
func (m *Map) GetTile(x, y int) uo.Tile {
	l := uo.Point{
		X: x,
		Y: y,
	}.Bound()
	cx := l.X / uo.ChunkWidth
	cy := l.Y / uo.ChunkHeight
	c := m.chunks[cy*uo.MapChunksWidth+cx]
	tx := l.X % uo.ChunkWidth
	ty := l.Y % uo.ChunkHeight
	return c.Tiles[ty*uo.ChunkWidth+tx]
}

// LoadFromMul reads in all of the segments of the given MapMul object and
// updates the map
func (m *Map) LoadFromMuls(mMul *file.MapMul, sMul *file.StaticsMul) {
	fn := func(x, y int) (lowest, average, highest int) {
		zTop := m.GetTile(x, y).RawZ()
		zLeft := m.GetTile(x, y+1).RawZ()
		zRight := m.GetTile(x+1, y).RawZ()
		zBottom := m.GetTile(x+1, y+1).RawZ()
		lowest = zTop
		if zLeft < lowest {
			lowest = zLeft
		}
		if zRight < lowest {
			lowest = zRight
		}
		if zBottom < lowest {
			lowest = zBottom
		}
		highest = zTop
		if zLeft > highest {
			highest = zLeft
		}
		if zRight > highest {
			highest = zRight
		}
		if zBottom > highest {
			highest = zBottom
		}
		tbdif := zTop - zBottom
		if tbdif < 0 {
			tbdif *= -1
		}
		lrdif := zLeft - zRight
		if lrdif < 0 {
			lrdif *= -1
		}
		if tbdif > lrdif {
			average = zLeft + zRight
		} else {
			average = zTop + zBottom
		}
		if average < 0 {
			average--
		}
		average /= 2
		return lowest, average, highest
	}
	// Load the tiles
	for iy := 0; iy < uo.MapHeight; iy++ {
		for ix := 0; ix < uo.MapWidth; ix++ {
			cx := ix / uo.ChunkWidth
			cy := iy / uo.ChunkHeight
			c := m.chunks[cy*uo.MapChunksWidth+cx]
			tx := ix % uo.ChunkWidth
			ty := iy % uo.ChunkHeight
			c.Tiles[ty*uo.ChunkWidth+tx] = mMul.GetTile(ix, iy)
		}
	}
	// Pre-calculate tile elevations
	for iy := 0; iy < uo.MapHeight; iy++ {
		for ix := 0; ix < uo.MapWidth; ix++ {
			cx := ix / uo.ChunkWidth
			cy := iy / uo.ChunkHeight
			c := m.chunks[cy*uo.MapChunksWidth+cx]
			tx := ix % uo.ChunkWidth
			ty := iy % uo.ChunkHeight
			t := c.Tiles[ty*uo.ChunkWidth+tx]
			lowest, avg, height := fn(ix, iy)
			t = t.SetElevations(lowest, avg, height)
			c.Tiles[ty*uo.ChunkWidth+tx] = t
		}
	}
	// Load the statics
	for _, static := range sMul.Statics() {
		cx := static.Location.X / uo.ChunkWidth
		cy := static.Location.Y / uo.ChunkHeight
		c := m.chunks[cy*uo.MapChunksWidth+cx]
		c.Statics = append(c.Statics, static)
	}
	// Sort statics by bottom Z
	for iy := 0; iy < uo.MapChunksHeight; iy++ {
		for ix := 0; ix < uo.MapChunksWidth; ix++ {
			c := m.chunks[iy*uo.MapChunksWidth+ix]
			sort.Slice(c.Statics, func(i, j int) bool {
				si := c.Statics[i]
				sj := c.Statics[j]
				sit := si.Location.Z + si.Height()
				sjt := sj.Location.Z + sj.Height()
				if si.Location.Z == sj.Location.Z {
					return sit < sjt
				}
				return si.Location.Z < sj.Location.Z
			})
		}
	}
}

// AfterUnmarshal calls AfterUnmarshalOntoMap calls for all map objects. We do
// this with a pre-compiled list of objects so that calls to
// AfterUnmarshalOntoMap can call world.Remove() if needed.
func (m *Map) AfterUnmarshal() {
	var mobs []*Mobile
	for _, c := range m.chunks {
		mobs = append(mobs, c.Mobiles...)
	}
	for _, m := range mobs {
		m.AfterUnmarshalOntoMap()
	}
}
