package game

import (
	"io"
	"strconv"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/uo/file"
	"github.com/qbradq/sharduo/lib/util"
)

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
		m.getChunk(static.Location).statics = append(m.getChunk(static.Location).statics, static)
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
	lfw.WriteComment("items")
	for _, c := range m.chunks {
		for _, item := range c.items {
			lfw.WriteLine(item.Serial().String())
		}
	}
	lfw.WriteComment("mobiles")
	for _, c := range m.chunks {
		for _, mobile := range c.mobiles {
			lfw.WriteLine(mobile.Serial().String())
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

// SetNewParent sets the parent object of this object. It properly removes
// the object from the old parent and adds the object to the new parent. Use
// nil to represent the world. This function returns false if the operation
// failed for any reason.
func (m *Map) SetNewParent(o, p Object) bool {
	oldParent := o.Parent()
	if oldParent == nil {
		if !world.Map().RemoveObject(o) {
			return false
		}
	} else {
		if !oldParent.RemoveObject(o) {
			return false
		}
	}
	addFailed := false
	if p == nil {
		if !world.Map().AddObject(o) {
			addFailed = true
		}
	} else {
		if !p.AddObject(o) {
			addFailed = true
		}
	}
	if addFailed {
		// Don't leak the object
		if oldParent == nil {
			world.Map().ForceAddObject(o)
		} else {
			oldParent.ForceAddObject(o)
		}
		return false
	}
	// Figure out if we need to send drag packets
	if _, ok := oldParent.(*VoidObject); ok {
		// If the item was coming to the Void it doesn't need a drag packet
		return true
	}
	if p == nil && oldParent == nil {
		// No drag packets if this is map-to-map.
		return true
	}
	if p != nil && oldParent != nil && p.Serial() == oldParent.Serial() {
		// We don't need a drag packet if we are not changing parents. This
		// is the case when lifting an item off the paper doll, dropping an
		// item onto the same mobile's paper doll, and dropping an item into
		// any container within the same mobile's inventory. This is also
		// true of map-to-map teleports.
		return true
	}
	item, ok := o.(Item)
	if !ok {
		// Drag packets only happen for items
		return true
	}
	newLocation := item.RootParent().Location()
	if item.RootParent().Serial().IsMobile() {
		newLocation.Z += 18
	}
	oldLocation := newLocation
	newParentMobile, _ := p.(Mobile)
	oldParentMobile, _ := oldParent.(Mobile)
	if oldParent != nil {
		oldLocation = oldParent.RootParent().Location()
		if oldParent.RootParent().Serial().IsMobile() {
			oldLocation.Z += 18
		}
	} else {
		oldLocation = item.Location()
	}
	r := oldLocation.XYDistance(newLocation) + uo.MaxViewRange
	r = uo.BoundUpdateRange(r)
	for _, mob := range m.GetNetStatesInRange(oldLocation, r) {
		mob.NetState().DragItem(item, oldParentMobile, oldLocation, newParentMobile, newLocation)
	}
	return true
}

// SendEverything sends everything in range to the mobile
func (m *Map) SendEverything(mob Mobile) {
	if mob.NetState() == nil {
		return
	}
	for _, o := range m.GetObjectsInRange(mob.Location(), mob.ViewRange()) {
		mob.NetState().SendObject(o)
	}
}

// RemoveEverything removes everything in range of the mobile
func (m *Map) RemoveEverything(mob Mobile) {
	if mob.NetState() == nil {
		return
	}
	for _, o := range m.GetObjectsInRange(mob.Location(), mob.ViewRange()) {
		mob.NetState().RemoveObject(o)
	}
}

// RemoveObject removes the object from the map. It always returns true, even if
// the object was not on the map to begin with.
func (m *Map) RemoveObject(o Object) bool {
	m.ForceRemoveObject(o)
	return true
}

// AddObject adds an object to the map sending all proper updates. It returns
// false only if the object could not fit on the map.
func (m *Map) AddObject(o Object) bool {
	// TODO Make sure there is enough room
	m.ForceAddObject(o)
	return true
}

// ForceAddObject places the object on the map without regard to any blockers.
func (m *Map) ForceAddObject(o Object) {
	if o == nil {
		return
	}
	o.SetParent(nil)
	c := m.getChunk(o.Location().Bound())
	c.Add(o)
	// Send the new object to all mobiles in range with an attached net state
	for _, mob := range m.GetNetStatesInRange(o.Location(), uo.MaxViewRange) {
		mob.NetState().SendObject(o)
	}
	// If this is a mobile with a NetState we have to send all of the items
	// and mobiles in range.
	if mob, ok := o.(Mobile); ok && mob.NetState() != nil {
		m.SendEverything(mob)
	}
}

// ForceRemoveObject removes the object from the map and always succeeds.
func (m *Map) ForceRemoveObject(o Object) {
	c := m.getChunk(o.Location())
	c.Remove(o)
	// Tell other mobiles with net states in range about the object removal
	for _, mob := range m.GetNetStatesInRange(o.Location(), uo.MaxViewRange) {
		if mob.Location().XYDistance(o.Location()) <= mob.ViewRange() {
			mob.NetState().RemoveObject(o)
		}
	}
	// If this is a mobile with a net state we need to remove all objects
	if mob, ok := o.(Mobile); ok && mob.NetState() != nil {
		m.RemoveEverything(mob)
	}
}

// MoveMobile moves a mobile in the given direction. Returns true if the
// movement was successful.
func (m *Map) MoveMobile(mob Mobile, dir uo.Direction) bool {
	// Change facing request
	dir = dir.Bound().StripRunningFlag()
	if mob.Facing() != dir {
		mob.SetFacing(dir)
		for _, othermob := range m.GetNetStatesInRange(mob.Location(), uo.MaxViewRange+1) {
			othermob.NetState().UpdateMobile(mob)
		}
		mob.AfterMove()
		return true
	}
	// Movement request
	oldLocation := mob.Location()
	newLocation := mob.Location().Forward(dir).WrapAndBound(oldLocation)
	oldChunk := m.getChunk(oldLocation)
	newChunk := m.getChunk(newLocation)
	// If this is a mobile with an attached net state we need to check for
	// new and old objects.
	if mob.NetState() != nil {
		for _, o := range m.GetObjectsInRange(mob.Location(), mob.ViewRange()+1) {
			if oldLocation.XYDistance(o.Location()) <= mob.ViewRange() &&
				newLocation.XYDistance(o.Location()) > mob.ViewRange() {
				// Object used to be in range and isn't anymore, delete it
				mob.NetState().RemoveObject(o)
			} else if oldLocation.XYDistance(o.Location()) > mob.ViewRange() &&
				newLocation.XYDistance(o.Location()) <= mob.ViewRange() {
				// Object used to be out of range but is in range now, send information about it
				mob.NetState().SendObject(o)
			}
		}
	}
	// Chunk updates
	if oldChunk != newChunk {
		oldChunk.Remove(mob)
	}
	// TODO Trigger events for moving off the tile
	mob.SetLocation(newLocation)
	// Now we need to check for attached net states that we might need to push
	// the movement to
	for _, othermob := range m.GetNetStatesInRange(mob.Location(), uo.MaxViewRange+1) {
		othermob.NetState().UpdateMobile(mob)
	}
	// TODO Trigger events for moving onto the tile
	if oldChunk != newChunk {
		newChunk.Add(mob)
	}
	mob.AfterMove()
	return true
}

// TeleportMobile moves a mobile from where it is now to the new location. This
// returns false if there is not enough room at that location for the mobile.
// This will also trigger all events as if the mobile left the tile normally,
// and arrived at the new tile normally.
func (m *Map) TeleportMobile(mob Mobile, l uo.Location) bool {
	oldLocation := mob.Location()
	world.Map().RemoveObject(mob) // This triggers on leave events
	mob.SetLocation(l)
	if !world.Map().AddObject(mob) { // This triggers on enter events
		// Best effort to not leak the mobile. Note that if we leak a player
		// mobile it will still be in the data set and will be retrieved when
		// the player logs in again.
		mob.SetLocation(oldLocation)
		world.Map().AddObject(mob)
		return false
	}
	if mob.NetState() != nil {
		mob.NetState().DrawPlayer()
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

// GetItemsInRange returns a slice of all items within the given range of the
// given location.
func (m *Map) GetItemsInRange(l uo.Location, r int) []Item {
	var ret []Item
	for _, c := range m.getChunksInRange(l, r) {
		for _, item := range c.items {
			d := l.XYDistance(item.Location())
			if d > r {
				continue
			}
			ret = append(ret, item)
		}
	}
	return ret
}

// GetMobilesInRange returns a slice of all items within the given range of the
// given location.
func (m *Map) GetMobilesInRange(l uo.Location, r int) []Mobile {
	var ret []Mobile
	for _, c := range m.getChunksInRange(l, r) {
		for _, mob := range c.mobiles {
			d := l.XYDistance(mob.Location())
			if d > r {
				continue
			}
			ret = append(ret, mob)
		}
	}
	return ret
}

// GetNetStatesInRange returns a slice of all mobiles in range of the given
// location with attached net states. Mobile.n / Mobile.NetState() will always
// be non-null.
func (m *Map) GetNetStatesInRange(l uo.Location, r int) []Mobile {
	var ret []Mobile
	for _, c := range m.getChunksInRange(l, r) {
		for _, mob := range c.mobiles {
			if mob.NetState() == nil {
				continue
			}
			d := l.XYDistance(mob.Location())
			if d > r {
				continue
			}
			ret = append(ret, mob)
		}
	}
	return ret
}

// GetObjectsInRange returns a slice of all objects within the given range of
// the given location.
func (m *Map) GetObjectsInRange(l uo.Location, r int) []Object {
	var ret []Object
	for _, c := range m.getChunksInRange(l, r) {
		for _, item := range c.items {
			d := l.XYDistance(item.Location())
			if d > r {
				continue
			}
			ret = append(ret, item)
		}
		for _, mob := range c.mobiles {
			d := l.XYDistance(mob.Location())
			if d > r {
				continue
			}
			ret = append(ret, mob)
		}
	}
	return ret
}

// UpdateViewRangeForMobile handles an update of the mobiles ViewRange value
// in a way that sends the correct packets to the attached NetState, if any.
func (m *Map) UpdateViewRangeForMobile(mob Mobile, r int) {
	r = uo.BoundViewRange(r)
	if r == mob.ViewRange() {
		return
	}
	if r < mob.ViewRange() {
		// Look for the set of currently-visible objects that will no longer be
		for _, o := range m.GetObjectsInRange(mob.Location(), mob.ViewRange()) {
			if mob.Location().XYDistance(o.Location()) > r {
				mob.NetState().RemoveObject(o)
			}
		}
	} else {
		// Look for the set of currently-non-visible objects that will be
		for _, o := range m.GetObjectsInRange(mob.Location(), r) {
			if mob.Location().XYDistance(o.Location()) > mob.ViewRange() {
				mob.NetState().SendObject(o)
			}
		}
	}
	mob.SetViewRange(r)
}

// GetTopSurface returns the highest solid object at the given location that has
// a Z altitude less than or equal to zLimit.
func (m *Map) GetTopSurface(l uo.Location, zLimit int) uo.CommonObject {
	var topObj uo.CommonObject
	zLimit = uo.BoundZ(zLimit)
	topZ := uo.MapMinZ
	c := m.getChunk(l)
	t := c.GetTile(l.X%uo.ChunkWidth, l.Y%uo.ChunkHeight)
	topObj = t
	if !t.Ignore() {
		avgZ := m.GetAverageTerrainZ(l)
		if avgZ == zLimit {
			return t
		} else if avgZ < zLimit {
			topZ = avgZ
		}
		// Else topz is still uo.MapMinZ and we are looking for statics and
		// items underground.
	}
	for _, static := range c.statics {
		// Only select surfaces
		if !static.Surface() && !static.Wet() {
			continue
		}
		staticTopZ := static.Z() + static.Height()
		if staticTopZ > topZ && staticTopZ <= zLimit {
			if staticTopZ == zLimit {
				return static
			} else {
				topZ = staticTopZ
				topObj = static
			}
		}
	}
	for _, item := range c.items {
		// Only consider surface items
		if !item.Surface() && !item.Wet() {
			continue
		}
		// Only look at items at our location
		if item.Location().X != l.X || item.Location().Y != l.Y {
			continue
		}
		itemTopZ := item.Z() + item.Height()
		if itemTopZ > topZ && itemTopZ <= zLimit {
			if itemTopZ == zLimit {
				return item
			} else {
				topZ = itemTopZ
				topObj = item
			}
		}
	}
	return topObj
}

// GetAverageTerrainZ returns the average Z coordinate of the terrain at the
// location.
func (m *Map) GetAverageTerrainZ(l uo.Location) int {
	var ret int
	zTop := m.GetTile(l.X, l.Y).Z()
	zLeft := m.GetTile(l.X, l.Y+1).Z()
	zRight := m.GetTile(l.X+1, l.Y).Z()
	zBottom := m.GetTile(l.X+1, l.Y+1).Z()
	z := zTop
	if zLeft < z {
		z = zLeft
	}
	if zRight < z {
		z = zRight
	}
	if zBottom < z {
		z = zBottom
	}
	top := zTop
	if zLeft > top {
		top = zLeft
	}
	if zRight > top {
		top = zRight
	}
	if zBottom > top {
		top = zBottom
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
		ret = zLeft + zRight
	} else {
		ret = zTop + zBottom
	}
	if ret < 0 {
		ret--
	}
	return ret / 2
}
