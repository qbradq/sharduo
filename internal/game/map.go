package game

import (
	"io"
	"strconv"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/uo/file"
	"github.com/qbradq/sharduo/lib/util"
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

// LoadFromMul reads in all of the segments of the given MapMul object and
// updates the map
func (m *Map) LoadFromMuls(mapmul *file.MapMul, staticsmul *file.StaticsMul) {
	// Load the tiles
	for iy := 0; iy < uo.MapHeight; iy++ {
		for ix := 0; ix < uo.MapWidth; ix++ {
			m.getChunk(uo.Location{X: ix, Y: iy}).setTile(ix, iy, mapmul.GetTile(ix, iy))
		}
	}
	// Load the statics
	for _, static := range staticsmul.Statics() {
		m.getChunk(static.Location).statics = append(m.getChunk(static.Location).statics,
			Static{
				Graphic:  static.Graphic,
				Location: static.Location,
			})
	}
}

// Read reads all map properties and object references from the file. This uses
// streaming to avoid allocating a large amount of memory all at once.
func (m *Map) Read(r io.Reader) []error {
	var errs []error
	lfr := &util.ListFileReader{}
	lfr.StartReading(r)
	for {
		sname := lfr.StreamNextSegmentHeader()
		// End of file or error condition
		if sname == "" {
			break
		} else if sname == "MapChildren" {
			// Object references to all of the child objects of the map
			for {
				e := lfr.StreamNextEntry()
				// End of segment
				if e == "" {
					break
				}
				n, err := strconv.ParseInt(e, 0, 32)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				o := world.Find(uo.Serial(n))
				m.getChunk(o.Location()).Add(o)
			}
		} else {
			// End of file or error
			if !lfr.SkipCurrentSegment() {
				break
			}
		}
	}
	return append(lfr.Errors(), errs...)
}

// Write writes all map object references to the file.
func (m *Map) Write(w io.WriteCloser) []error {
	lfw := util.NewListFileWriter(w)
	defer lfw.Close()

	// Write all object references of direct child objects of the map. Note that
	// we do this in a list file because using a TagFileObject would allocate
	// stupid amounts of memory when trying to load it.
	lfw.WriteComment("generated by game.Map.Write")
	lfw.WriteBlankLine()
	lfw.WriteSegmentHeader("MapChildren")
	for _, c := range m.chunks {
		for _, o := range c.objects {
			lfw.WriteLine(o.Serial().String())
		}
	}
	lfw.WriteBlankLine()
	lfw.WriteComment("END OF FILE")
	return nil
}

// getChunk returns a pointer to the chunk for the given location.
func (m *Map) getChunk(l uo.Location) *chunk {
	l = l.WrapAndBound(l)
	cx := l.X / uo.ChunkWidth
	cy := l.Y / uo.ChunkHeight
	return m.chunks[cy*uo.MapChunksWidth+cx]
}

// GetTile returns the Tile value for the location
func (m *Map) GetTile(x, y int) uo.Tile {
	return m.getChunk(uo.Location{
		X: x,
		Y: y,
	}).GetTile(x%uo.ChunkWidth, y%uo.ChunkHeight)
}

// AddNewObject adds a new object to the map at the given location
func (m *Map) AddNewObject(o Object) {
	c := m.getChunk(o.Location())
	c.Add(o)
	others := m.GetObjectsInRange(o.Location(), uo.MaxViewRange+1)
	// Send the item to all mobiles in range with an attached NetState
	for _, other := range others {
		if o == other {
			continue
		}
		mob, ok := other.(Mobile)
		if !ok {
			continue
		}
		if mob.NetState() != nil && o.Location().XYDistance(mob.Location()) <= mob.ViewRange() {
			if item, ok := o.(Item); ok {
				mob.NetState().SendItem(item)
			}
		}
	}

	// If this is a mobile with a NetState we have to send all of the items
	// and mobiles in range.
	if mob, ok := o.(Mobile); ok && mob.NetState() != nil {
		mobs := m.GetObjectsInRange(mob.Location(), mob.ViewRange())
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

// MoveObject moves an object in the given direction. Returns true if the
// movement was successfull.
func (m *Map) MoveObject(o Object, dir uo.Direction) bool {
	mob, ok := o.(Mobile)
	if !ok {
		return false
	}
	// Change facing request
	dir = dir.Bound()
	if mob.Facing() != dir {
		mob.SetFacing(dir)
		return true
	}
	// Movement request
	oldLocation := mob.Location()
	newLocation := mob.Location().Forward(dir).WrapAndBound(oldLocation)
	oldChunk := m.getChunk(oldLocation)
	newChunk := m.getChunk(newLocation)
	others := m.GetObjectsInRange(oldLocation, mob.ViewRange()+2)
	// If this is a mobile with an attached net state we need to check for
	// new and old objects.
	if ok && mob.NetState() != nil {
		for _, other := range others {
			if other == mob {
				continue
			}
			// Object used to be in range and isn't anymore, delete it
			if oldLocation.XYDistance(other.Location()) <= mob.ViewRange() &&
				newLocation.XYDistance(other.Location()) > mob.ViewRange() {
				mob.NetState().RemoveObject(other)
			} else if oldLocation.XYDistance(other.Location()) > mob.ViewRange() &&
				newLocation.XYDistance(other.Location()) <= mob.ViewRange() {
				// Object used to be out of range but is in range now, send information about it
				if item, ok := other.(Item); ok {
					mob.NetState().SendItem(item)
				}
			}
		}
	}
	// Now we need to check for attached net states that we might need to push
	// a new object to
	for _, other := range others {
		if other == o {
			continue
		}
		mob, ok := other.(Mobile)
		if !ok || mob.NetState() == nil {
			continue
		}
	}
	// Chunk updates
	if oldChunk != newChunk {
		oldChunk.Remove(o)
	}
	o.SetLocation(newLocation)
	if oldChunk != newChunk {
		newChunk.Add(o)
	}
	return true
}

// getChunksInBounds returns a slice of all the chunks within a given bounds.
func (m *Map) getChunksInBounds(b uo.Bounds) []*chunk {
	var ret []*chunk
	scx := b.X / uo.ChunkWidth
	scy := b.Y / uo.ChunkHeight
	ecx := (b.X + b.W - 1) / uo.ChunkWidth
	ecy := (b.Y + b.H - 1) / uo.ChunkHeight
	for cy := scy; cy <= ecy; cy++ {
		for cx := scx; cx <= ecx; cx++ {
			ret = append(ret, m.chunks[cy*uo.MapChunksWidth+cx])
		}
	}
	return ret
}

// getChunksInRange gets chunks in the given range of a reference point.
func (m *Map) getChunksInRange(l uo.Location, r int) []*chunk {
	return m.getChunksInBounds(uo.Bounds{
		X: l.X - r,
		Y: l.Y - r,
		W: r*2 + 1,
		H: r*2 + 1,
	})
}

// GetObjectsInRange returns a slice of all objects within the given range of
// the given location.
func (m *Map) GetObjectsInRange(l uo.Location, r int) []Object {
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
