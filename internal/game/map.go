package game

import (
	"fmt"
	"io"
	"sort"
	"strings"
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

// mwRetBuf is the static buffer used for the return value of
// [Map.MobilesWithin].
var mwRetBuf []*Mobile

// iwRetBuf is the static buffer used for the return value of
// [Map.ItemsWithin].
var iwRetBuf []*Item

// NetStatesInRange returns a slice of all of the mobiles with attached net
// states who's view range is within range of the given point on the map. The
// second parameter is an additional range to apply to view ranges. This is
// useful for example when distributing mobile movements. Subsequent calls to
// NetStatesInRange reuses the same return array.
func (m *Map) NetStatesInRange(cp uo.Point, extra int) []*Mobile {
	var p uo.Point
	b := cp.BoundsByRadius(uo.MaxViewRange)
	cb := uo.Bounds{
		X: b.X / uo.ChunkWidth,
		Y: b.Y / uo.ChunkHeight,
		W: b.W / uo.ChunkWidth,
		H: b.H / uo.ChunkHeight,
	}
	if b.W%uo.ChunkWidth != 0 {
		cb.W++
	}
	if b.H%uo.ChunkHeight != 0 {
		cb.H++
	}
	nsRetBuf = nsRetBuf[:0]
	for p.Y = cb.Y; p.Y < cb.Y+cb.H; p.Y++ {
		for p.X = cb.X; p.X < cb.X+cb.W; p.X++ {
			c := m.chunks[p.Y*uo.MapChunksWidth+p.X]
			for _, m := range c.Mobiles {
				if m.NetState == nil ||
					m.Location.XYDistance(cp) > m.ViewRange+extra {
					continue
				}
				nsRetBuf = append(nsRetBuf, m)
			}
		}
	}
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

// StoreObject moves an object to deep storage.
func (m *Map) StoreObject(obj any) {
	switch o := obj.(type) {
	case *Mobile:
		m.ds[o.Serial] = o
	case *Item:
		m.ds[o.Serial] = o
	}
}

// RetrieveObject removes an object from deep storage and returns it.
func (m *Map) RetrieveObject(s uo.Serial) any {
	o, found := m.ds[s]
	if found {
		delete(m.ds, s)
		return o
	}
	return nil
}

// MobilesWithin returns all mobiles within the bounds. Subsequent calls to
// MobilesWithin reuse the same backing array for return values.
func (m *Map) MobilesWithin(b uo.Bounds) []*Mobile {
	var p uo.Point
	cb := uo.Bounds{
		X: b.X / uo.ChunkWidth,
		Y: b.Y / uo.ChunkHeight,
		W: b.W / uo.ChunkWidth,
		H: b.H / uo.ChunkHeight,
	}
	if b.W%uo.ChunkWidth != 0 {
		cb.W++
	}
	if b.H%uo.ChunkHeight != 0 {
		cb.H++
	}
	mwRetBuf = mwRetBuf[:0]
	for p.Y = cb.Y; p.Y < cb.Y+cb.H; p.Y++ {
		for p.X = cb.X; p.X < cb.X+cb.W; p.X++ {
			c := m.chunks[p.Y*uo.MapChunksWidth+p.X]
			for _, m := range c.Mobiles {
				if b.Contains(m.Location) {
					mwRetBuf = append(mwRetBuf, m)
				}
			}
		}
	}
	return mwRetBuf
}

// ItemsWithin returns all items within the bounds. Subsequent calls to
// ItemsWithin reuse the same backing array for return values.
func (m *Map) ItemsWithin(b uo.Bounds) []*Item {
	var p uo.Point
	cb := uo.Bounds{
		X: b.X / uo.ChunkWidth,
		Y: b.Y / uo.ChunkHeight,
		W: b.W / uo.ChunkWidth,
		H: b.H / uo.ChunkHeight,
	}
	if b.W%uo.ChunkWidth != 0 {
		cb.W++
	}
	if b.H%uo.ChunkHeight != 0 {
		cb.H++
	}
	iwRetBuf = iwRetBuf[:0]
	for p.Y = cb.Y; p.Y < cb.Y+cb.H; p.Y++ {
		for p.X = cb.X; p.X < cb.X+cb.W; p.X++ {
			c := m.chunks[p.Y*uo.MapChunksWidth+p.X]
			for _, i := range c.Items {
				if b.Contains(i.Location) {
					iwRetBuf = append(iwRetBuf, i)
				}
			}
		}
	}
	return iwRetBuf
}

// EverythingWithin returns all items and mobiles within the bounds. Subsequent
// calls to EverythingWithin will reuse the return buffers for
// [Map.MobilesWithin] and [Map.ItemsWithin].
func (m *Map) EverythingWithin(b uo.Bounds) ([]*Mobile, []*Item) {
	var p uo.Point
	cb := uo.Bounds{
		X: b.X / uo.ChunkWidth,
		Y: b.Y / uo.ChunkHeight,
		W: b.W / uo.ChunkWidth,
		H: b.H / uo.ChunkHeight,
	}
	if b.W%uo.ChunkWidth != 0 {
		cb.W++
	}
	if b.H%uo.ChunkHeight != 0 {
		cb.H++
	}
	mwRetBuf = mwRetBuf[:0]
	iwRetBuf = iwRetBuf[:0]
	for p.Y = cb.Y; p.Y < cb.Y+cb.H; p.Y++ {
		for p.X = cb.X; p.X < cb.X+cb.W; p.X++ {
			c := m.chunks[p.Y*uo.MapChunksWidth+p.X]
			for _, m := range c.Mobiles {
				if b.Contains(m.Location) {
					mwRetBuf = append(mwRetBuf, m)
				}
			}
			for _, i := range c.Items {
				if b.Contains(i.Location) {
					iwRetBuf = append(iwRetBuf, i)
				}
			}
		}
	}
	return mwRetBuf, iwRetBuf
}

// SendEverything sends all mobiles and items to the given mobile's net state.
func (m *Map) SendEverything(mob *Mobile) {
	if mob.NetState == nil {
		return
	}
	mobs, items := m.EverythingWithin(mob.Location.BoundsByRadius(mob.ViewRange))
	for _, m := range mobs {
		mob.NetState.SendObject(m)
	}
	for _, i := range items {
		mob.NetState.SendObject(i)
	}
}

// SendSpeech sends speech to all mobiles in range.
func (m *Map) SendSpeech(from *Mobile, r int, format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	mobs := m.MobilesWithin(from.Location.BoundsByRadius(r))
	sort.Slice(mobs, func(i, j int) bool {
		return mobs[i].Location.XYDistance(from.Location) <
			mobs[j].Location.XYDistance(from.Location)
	})
	isAllCommand := len(text) >= 4 && strings.ToLower(text[:4]) == "all "
	speechEventHandled := false
	for _, mob := range mobs {
		if from.Location.XYDistance(mob.Location) <= mob.ViewRange {
			if mob.NetState != nil {
				mob.NetState.Speech(from, text)
			}
			// Make sure we don't trigger every listener in range, just the
			// closest, unless it is an "all" command.
			if !speechEventHandled {
				speechEventHandled = mob.ExecuteEvent("Speech", from, text) && !isAllCommand
			}
		}
	}
}

// MoveMobile moves a mobile in the given direction. Returns true if the
// movement was successful.
func (m *Map) MoveMobile(mob *Mobile, dir uo.Direction) bool {
	dir = dir.Bound().StripRunningFlag()
	if mob.Facing != dir {
		// Change facing request
		mob.Facing = dir
		for _, om := range m.NetStatesInRange(mob.Location, 1) {
			om.NetState.MoveMobile(mob)
		}
		return true
	} // else move request
	// Stamina check
	if mob.Stamina <= 0 {
		if mob.NetState != nil {
			mob.NetState.Cliloc(nil, 500110) // You are too fatigued to move.
			return false
		}
	}
	ol := mob.Location
	// Check movement
	success, nl, floor := m.canMoveTo(mob, dir)
	if !success {
		return false
	}
	// Check diagonals if required
	if dir.IsDiagonal() {
		if success, _, _ := m.canMoveTo(mob, dir.Left()); !success {
			return false
		}
		if success, _, _ := m.canMoveTo(mob, dir.Right()); !success {
			return false
		}
	}
	oc := m.chunks[(ol.Y/uo.ChunkHeight)*uo.MapChunksWidth+(ol.X/uo.ChunkWidth)]
	nc := m.chunks[(nl.Y/uo.ChunkHeight)*uo.MapChunksWidth+(nl.X/uo.ChunkWidth)]
	// If this is a mobile with an attached net state we need to check for
	// new and old objects.
	if mob.NetState != nil {
		mobs, items := m.EverythingWithin(mob.Location.BoundsByRadius(mob.ViewRange + 1))
		for _, om := range mobs {
			if ol.XYDistance(om.Location) <= mob.ViewRange &&
				nl.XYDistance(om.Location) > mob.ViewRange {
				// Mobile used to be in range and isn't anymore, delete it
				mob.NetState.RemoveObject(om)
			} else if ol.XYDistance(om.Location) > mob.ViewRange &&
				nl.XYDistance(om.Location) <= mob.ViewRange {
				// Mobile used to be out of range but is in range now, send
				// information about it
				mob.NetState.SendObject(om)
			}
		}
		for _, oi := range items {
			if ol.XYDistance(oi.Location) <= mob.ViewRange &&
				nl.XYDistance(oi.Location) > mob.ViewRange {
				// Item used to be in range and isn't anymore, delete it
				mob.NetState.RemoveObject(oi)
			} else if ol.XYDistance(oi.Location) > mob.ViewRange &&
				nl.XYDistance(oi.Location) <= mob.ViewRange {
				// Item used to be out of range but is in range now, send
				// information about it
				mob.NetState.SendObject(oi)
			}
		}
	}
	// Chunk updates
	if oc != nc {
		oc.RemoveMobile(mob)
	}
	mob.Location = nl
	// Now we need to check for attached net states that we might need to push
	// the movement to
	otherMobs := m.NetStatesInRange(mob.Location, 1)
	for _, om := range otherMobs {
		if ol.XYDistance(om.Location) <= om.ViewRange &&
			nl.XYDistance(om.Location) > om.ViewRange {
			// We used to be in visible range of the other mobile but are
			// moving out of that range, delete us
			om.NetState.RemoveObject(mob)
		} else if ol.XYDistance(om.Location) > om.ViewRange &&
			nl.XYDistance(om.Location) <= om.ViewRange {
			// We used to be outside the visible range of the other mobile but
			// are moving into that range, send us
			om.NetState.SendObject(mob)
		} else {
			om.NetState.MoveMobile(mob)
		}
	}
	// Update mobile standing
	mob.StandingOn = floor
	if oc != nc {
		nc.AddMobile(mob)
	}
	mob.AfterMove()
	return true
}

// canMoveTo returns true if the mobile can move from its current location in
// the given direction. If the first return value is true the second return
// value will be the new location of the mobile if it were to move to the new
// location, and the third return value is a description of the surface they
// would be standing on. This method enforces rules about surfaces that block
// movement and minimum height clearance. Note that the required clearance for
// all mobiles is uo.PlayerHeight. Many places in Britannia - especially in and
// around dungeons - would block monster movement if they were given heights
// greater than this value.
func (m *Map) canMoveTo(mob *Mobile, d uo.Direction) (bool, uo.Point, uo.CommonObject) {
	ol := mob.Location
	nl := ol.Forward(d.Bound()).WrapAndBound(ol)
	nl.Z = mob.StandingOn.Highest()
	floor, ceiling := m.GetFloorAndCeiling(nl, false, true)
	// No floor to stand on, bail
	if floor == nil {
		return false, ol, nil
	}
	fz := floor.StandingHeight()
	if ceiling != nil {
		// See if there is enough room for the mobile to fit if it took the step
		cz := ceiling.Z()
		if cz-fz < uo.PlayerHeight {
			return false, ol, floor
		}
	}
	// Consider the step height
	if tile, ok := floor.(uo.Tile); ok {
		// The mobile is standing on the tile matrix
		if tile.Impassable() {
			// Check tile flags
			return false, ol, floor
		}
		// There are no step height restrictions when following the terrain
		nl.Z = fz
	} else {
		// The mobile is standing on an item, either static or dynamic
		if !floor.Surface() && !floor.Wet() {
			// Check tile flags
			return false, ol, floor
		}
		// Consider step height
		oldFloor := mob.StandingOn
		oldTop := oldFloor.Highest()
		if !floor.Bridge() && fz-oldTop > uo.StepHeight {
			// Can't go up that much in one step
			return false, ol, floor
		}
		nl.Z = fz
	}
	return true, nl, floor
}

// GetFloorAndCeiling returns the objects that make up the floor below and the
// ceiling above the given reference location. These objects may be any of the
// objects contained within the map such as Tiles, Items, Statics, and Multis.
// A nil return value means that there is no floor below the position, or that
// there is no ceiling above the position. Normally at least one of the return
// values will be non-nil referencing at least the tile matrix. However there
// are certain places on the map - such as cave entrances - where the tile
// matrix is ignored. In these cases both return values may be nil if there are
// no items or statics to create a surface. If the ignoreDynamicItems argument
// is true then only Items with the static flag will be considered. If the
// considerStepHeight parameter is true then gaps less than or equal to
// uo.StepHeight will be ignored.
//
// NOTE: This function requires that all statics and items are z-sorted bottom
// to top.
func (m *Map) GetFloorAndCeiling(l uo.Point, ignoreDynamicItems, considerStepHeight bool) (uo.CommonObject, uo.CommonObject) {
	var floorObject uo.CommonObject
	var ceilingObject uo.CommonObject
	floor := int(uo.MapMinZ)
	ceiling := int(uo.MapMaxZ)
	footHeight := int(l.Z)
	// Consider tile matrix
	c := m.chunks[(l.Y/uo.ChunkHeight)*uo.MapChunksWidth+(l.X/uo.ChunkWidth)]
	t := c.Tiles[(l.Y%uo.ChunkHeight)*uo.ChunkWidth+(l.X%uo.ChunkWidth)]
	if !t.Ignore() {
		bottom := int(t.Z())
		avg := int(t.StandingHeight())
		if footHeight+int(uo.PlayerHeight) < bottom {
			// Mobile is completely below ground
			ceiling = avg
			ceilingObject = t
		} else if footHeight < bottom {
			// Mobile's feet are below the tile matrix but the head is above,
			// project upward
			floor = avg
			floorObject = t
			if floor > footHeight {
				footHeight = floor
			}
		} else if footHeight >= avg {
			// Mobile is above or on the ground
			floor = avg
			floorObject = t
		} else {
			// Mobile is down inside a tile in the tile matrix
			floor = avg
			floorObject = t
			if floor > footHeight {
				footHeight = floor
			}
		}
	}
	// Consider statics
	for _, static := range c.Statics {
		// Ignore statics that are not at the location
		if static.Location.X != l.X || static.Location.Y != l.Y {
			continue
		}
		// Only select solid statics ignoring things like leaves
		if !static.Surface() && !static.Impassable() {
			continue
		}
		sz := int(static.Z())
		stz := int(static.Highest())
		if stz < floor {
			// Static is below our current floor position, ignore
			continue
		}
		if stz == floor {
			// Static is even with our current floor position, so we need to
			// try to defer to the object with the most passability
			if floorObject.Impassable() {
				floorObject = static
			}
			continue
		}
		if (considerStepHeight && stz <= footHeight+int(uo.StepHeight)) || stz <= footHeight {
			// Static is underfoot, consider it a possible floor
			floor = stz
			floorObject = static
			continue
		}
		if sz <= footHeight {
			// Feet are inside or resting on this static so project upward
			floor = stz
			footHeight = floor
			floorObject = static
			continue
		}
		if considerStepHeight && static.Bridge() && sz > footHeight {
			// Feet are between the floor and a section of stair that is
			// floating more than uo.StepHeight units above the floor. This is a
			// common case an the client expects to be able to "hop" up onto
			// stairs like this. So we project upward.
			floor = stz
			footHeight = floor
			floorObject = static
			continue
		}
		// Underside of the static is above the foot position so that is the
		// ceiling
		ceiling = sz
		ceilingObject = static
		break
	}
	// Consider items
	for _, item := range c.Items {
		// Ignore dynamic items if requested
		if ignoreDynamicItems && !item.HasFlags(ItemFlagsStatic) {
			continue
		}
		// Only look at items at our location
		if item.Location.X != l.X || item.Location.Y != l.Y {
			continue
		}
		// Only select solid items. This ignores passible items like gold.
		if !item.Surface() && !item.Impassable() {
			continue
		}
		iz := int(item.Z())
		itz := int(item.Highest())
		if itz < floor {
			// Item is below the current floor, ignore it
			continue
		}
		if itz == floor {
			// Item is even with our current floor, defer to the object with the
			// most passability.
			if floorObject.Impassable() {
				floorObject = item
			}
			continue
		}
		if (considerStepHeight && itz <= footHeight+int(uo.StepHeight)) || itz <= footHeight {
			// Item is underfoot, consider it a possible floor
			if itz >= floor {
				// Surface of item is between the static floor and the foot
				// height so consider it a possible floor
				floor = itz
				floorObject = item
			} // Else the item is below the static floor so ignore it
			continue
		}
		if iz <= footHeight {
			// Feet are inside or resting on this item so project upward
			floor = itz
			footHeight = floor
			floorObject = item
			continue
		}
		if considerStepHeight && item.Bridge() && iz > footHeight {
			// Feet are between the floor and a section of stair that is
			// floating more than uo.StepHeight units above the floor. This is a
			// common case an the client expects to be able to "hop" up onto
			// stairs like this. So we project upward.
			floor = itz
			footHeight = floor
			floorObject = item
			continue
		}
		// Underside of the item is above the foot position so this is the last
		// item we need to check
		if iz < ceiling {
			// Underside of item is below the static ceiling so this item is the
			// ceiling
			ceilingObject = item
		}
		break
	}
	return floorObject, ceilingObject
}
