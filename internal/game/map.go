package game

import (
	"io"
	"sync"

	"github.com/qbradq/sharduo/lib/uo"
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
