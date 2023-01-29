package game

import (
	"io"
	"log"
	"strconv"

	"github.com/qbradq/sharduo/internal/marshal"
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
			m.getChunk(uo.Location{X: int16(ix), Y: int16(iy)}).setTile(ix, iy, mapmul.GetTile(ix, iy))
		}
	}
	// Pre-calculate tile elevations
	for iy := int16(0); iy < int16(uo.MapHeight); iy++ {
		for ix := int16(0); ix < int16(uo.MapWidth); ix++ {
			c := m.getChunk(uo.Location{X: ix, Y: iy})
			t := c.GetTile(ix, iy)
			lowest, avg, height := m.getTerrainElevations(ix, iy)
			t = t.SetElevations(lowest, avg, height)
			c.setTile(int(ix), int(iy), t)
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

// Marshal writes out all object references of objects that are directly on the
// map.
func (m *Map) Marshal(s *marshal.TagFileSegment) {
	for _, c := range m.chunks {
		for _, item := range c.items {
			s.PutInt(uint32(item.Serial()))
			s.IncrementRecordCount()
		}
		for _, mobile := range c.mobiles {
			s.PutInt(uint32(mobile.Serial()))
			s.IncrementRecordCount()
		}
	}
}

// Unmarshal reads all object references of the objects that are parented
// directly to the map.
func (m *Map) Unmarshal(s *marshal.TagFileSegment) {
	for i := uint32(0); i < s.RecordCount(); i++ {
		serial := uo.Serial(s.Int())
		o := world.Find(serial)
		if o == nil {
			log.Printf("warning: map referenced leaked object %s", serial.String())
			continue
		}
		c := m.getChunk(o.Location())
		c.Add(o)
	}
}

// getChunk returns a pointer to the chunk for the given location.
func (m *Map) getChunk(l uo.Location) *chunk {
	l = l.Bound()
	cx := int(l.X) / uo.ChunkWidth
	cy := int(l.Y) / uo.ChunkHeight
	return m.chunks[cy*uo.MapChunksWidth+cx]
}

// GetTile returns the Tile value for the location
func (m *Map) GetTile(x, y int16) uo.Tile {
	return m.getChunk(uo.Location{
		X: x,
		Y: y,
	}).GetTile(x, y)
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
	itemRoot := RootParent(item)
	newLocation := itemRoot.Location()
	if itemRoot.Serial().IsMobile() {
		newLocation.Z += uo.PlayerHeight
	}
	oldLocation := newLocation
	newParentMobile, _ := p.(Mobile)
	oldParentMobile, _ := oldParent.(Mobile)
	if oldParent != nil {
		oldRoot := RootParent(oldParent)
		oldLocation = oldRoot.Location()
		if oldRoot.Serial().IsMobile() {
			oldLocation.Z += uo.PlayerHeight
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
// the object was not on the map to begin with. If the object removed was an
// item any other items that were resting on this one will be plopped downward.
func (m *Map) RemoveObject(o Object) bool {
	var l uo.Location
	item, ok := o.(Item)
	if ok {
		l = item.Location()
	}
	m.ForceRemoveObject(o)
	if ok {
		m.plopItems(l, item.Height())
	}
	return true
}

// AddObject adds an object to the map sending all proper updates. If the object
// is an item it will be stacked on top of any other items at the location. It
// returns false only if the object is an item and could not fit on the map.
func (m *Map) AddObject(o Object) bool {
	op := o.Parent()
	isVoid := op != nil && op.Serial() == TheVoid.serial
	// Make sure there is enough room for items
	if item, ok := o.(Item); ok && !m.plop(item) {
		if isVoid {
			// New item coming from the void, don't leak it
			m.ForceAddObject(item)
			return true
		}
		return false
	}
	m.ForceAddObject(o)
	return true
}

// ForceAddObject places the object on the map without regard to any blockers.
func (m *Map) ForceAddObject(o Object) {
	if o == nil {
		return
	}
	o.SetParent(nil)
	c := m.getChunk(o.Location())
	c.Add(o)
	// Send the new object to all mobiles in range with an attached net state
	for _, mob := range m.GetNetStatesInRange(o.Location(), uo.MaxViewRange) {
		mob.NetState().SendObject(o)
	}
	// If this is a mobile with a NetState we have to send all of the items
	// and mobiles in range.
	mob, ok := o.(Mobile)
	if ok && mob.NetState() != nil {
		m.SendEverything(mob)
	}
	// If this is a mobile that doesn't know what it's standing on we need to
	// tell it. This is the case when a mobile is first created from template.
	if ok && mob.StandingOn() == nil {
		floor, _ := m.GetFloorAndCeiling(o.Location())
		mob.StandOn(floor)
	}
}

// ForceRemoveObject removes the object from the map and always succeeds.
func (m *Map) ForceRemoveObject(o Object) {
	c := m.getChunk(o.Location().Bound())
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

// canMoveTo returns true if the mobile can move from its current location in
// the given direction. If the first return value is true the second return
// value will be the new location of the mobile if it were to move to the new
// location, and the third return value is a description of the surface they
// would be standing on. This method enforces rules about surfaces that block
// movement and minimum height clearance. Note that the required clearance for
// all mobiles is uo.PlayerHeight. Many places in Britannia - especially in and
// around dungeons - that would block monster movement if they were given
// heights greater than this value.
func (m *Map) canMoveTo(mob Mobile, d uo.Direction) (bool, uo.Location, uo.CommonObject) {
	ol := mob.Location()
	nl := ol.Forward(d.Bound()).WrapAndBound(ol)
	floor, ceiling := m.GetFloorAndCeiling(nl)
	// No floor to stand on, bail
	if floor == nil {
		return false, ol, nil
	}
	fz := floor.Z() + floor.StandingHeight()
	cz := ceiling.Z()
	// See if there is enough room for the mobile to fit if it took the step
	if cz-fz < uo.PlayerHeight {
		return false, ol, floor
	}
	// Consider the step height
	if tile, ok := floor.(uo.Tile); ok {
		// The mobile is standing on the tile matrix
		if (!tile.Surface() && !tile.Wet()) || tile.Impassable() {
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
		oldFloor := mob.StandingOn()
		oldTop := oldFloor.Z() + oldFloor.Height()
		if fz-oldTop > uo.StepHeight {
			// Can't go up that much in one step
			return false, ol, floor
		}
		nl.Z = fz
	}
	return true, nl, floor
}

// MoveMobile moves a mobile in the given direction. Returns true if the
// movement was successful.
func (m *Map) MoveMobile(mob Mobile, dir uo.Direction) bool {
	// Change facing request
	dir = dir.Bound().StripRunningFlag()
	if mob.Facing() != dir {
		mob.SetFacing(dir)
		for _, othermob := range m.GetNetStatesInRange(mob.Location(), uo.MaxViewRange+1) {
			othermob.NetState().MoveMobile(mob)
		}
		return true
	}
	// Movement request
	oldLocation := mob.Location()
	// Check movement
	success, newLocation, floor := m.canMoveTo(mob, dir)
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
		othermob.NetState().MoveMobile(mob)
	}
	// Update mobile standing
	mob.StandOn(floor)
	// TODO Trigger events for moving onto the tile
	if oldChunk != newChunk {
		newChunk.Add(mob)
	}
	mob.AfterMove()
	log.Printf("%+v", mob.Location())
	return true
}

// TeleportMobile moves a mobile from where it is now to the new location. This
// returns false if there is not enough room at that location for the mobile.
// This will also trigger all events as if the mobile left the tile normally,
// and arrived at the new tile normally. As a special case this function skips
// attempting to remove the mobile from its current location if the mobile's
// current parent is TheVoid. This is the case when the mobile is being created.
func (m *Map) TeleportMobile(mob Mobile, l uo.Location) bool {
	oldLocation := mob.Location()
	world.Map().RemoveObject(mob) // This triggers on leave events
	mob.SetLocation(l)
	if !world.Map().AddObject(mob) { // This triggers on enter events
		// Don't leak the mobile, just force it back where it came from.
		mob.SetLocation(oldLocation)
		world.Map().ForceAddObject(mob)
		floor, _ := m.GetFloorAndCeiling(mob.Location())
		mob.StandOn(floor)
		return false
	}
	// Update standing
	floor, _ := m.GetFloorAndCeiling(mob.Location())
	mob.StandOn(floor)
	// If this mobile has a net state attached we need to fully refresh the
	// client's object collection.
	if mob.NetState() != nil {
		mob.SetLocation(oldLocation)
		m.RemoveEverything(mob)
		mob.SetLocation(l)
		m.SendEverything(mob)
		mob.NetState().DrawPlayer()
	}
	return true
}

// Query returns true if there is a static or dynamic item within range of the
// given location who's BaseGraphic property is contained within the given set.
func (m *Map) Query(center uo.Location, queryRange int16, set map[uo.Graphic]struct{}) bool {
	b := uo.Bounds{
		X: center.X - queryRange,
		Y: center.Y - queryRange,
		W: queryRange*2 + 1,
		H: queryRange*2 + 1,
	}
	tl := uo.Location{
		X: b.X,
		Y: b.Y,
	}
	scx := int(b.X) / uo.ChunkWidth
	scy := int(b.Y) / uo.ChunkHeight
	ecx := int(b.X+b.W-1) / uo.ChunkWidth
	ecy := int(b.Y+b.H-1) / uo.ChunkHeight
	for cy := scy; cy <= ecy; cy++ {
		for cx := scx; cx <= ecx; cx++ {
			l := uo.Location{
				X: int16(cx * uo.ChunkWidth),
				Y: int16(cy * uo.ChunkHeight),
			}.WrapAndBound(tl)
			ccx := int(l.X) / uo.ChunkWidth
			ccy := int(l.Y) / uo.ChunkHeight
			c := m.chunks[ccy*uo.MapChunksWidth+ccx]
			for _, s := range c.statics {
				if _, ok := set[s.BaseGraphic()]; ok && center.XYDistance(s.Location) <= queryRange {
					return true
				}
			}
			for _, i := range c.items {
				if _, ok := set[i.BaseGraphic()]; ok && center.XYDistance(i.Location()) <= queryRange {
					return true
				}
			}
		}
	}
	return false
}

func (m *Map) getChunksInBounds(b uo.Bounds) []*chunk {
	var ret []*chunk
	tl := uo.Location{
		X: b.X,
		Y: b.Y,
	}
	scx := int(b.X) / uo.ChunkWidth
	scy := int(b.Y) / uo.ChunkHeight
	ecx := int(b.X+b.W-1) / uo.ChunkWidth
	ecy := int(b.Y+b.H-1) / uo.ChunkHeight
	for cy := scy; cy <= ecy; cy++ {
		for cx := scx; cx <= ecx; cx++ {
			l := uo.Location{
				X: int16(cx * uo.ChunkWidth),
				Y: int16(cy * uo.ChunkHeight),
			}.WrapAndBound(tl)
			ccx := int(l.X) / uo.ChunkWidth
			ccy := int(l.Y) / uo.ChunkHeight
			ret = append(ret, m.chunks[ccy*uo.MapChunksWidth+ccx])
		}
	}
	return ret
}

// getChunksInRange gets chunks in the given range of a reference point.
func (m *Map) getChunksInRange(l uo.Location, r int16) []*chunk {
	return m.getChunksInBounds(uo.Bounds{
		X: l.X - r,
		Y: l.Y - r,
		W: r*2 + 1,
		H: r*2 + 1,
	})
}

// GetItemsInRange returns a slice of all items within the given range of the
// given location.
func (m *Map) GetItemsInRange(l uo.Location, r int16) []Item {
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
func (m *Map) GetMobilesInRange(l uo.Location, r int16) []Mobile {
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
func (m *Map) GetNetStatesInRange(l uo.Location, r int16) []Mobile {
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
func (m *Map) GetObjectsInRange(l uo.Location, r int16) []Object {
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
func (m *Map) UpdateViewRangeForMobile(mob Mobile, r int16) {
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
func (m *Map) GetTopSurface(l uo.Location, zLimit int8) uo.CommonObject {
	var topObj uo.CommonObject
	topZ := uo.MapMinZ
	c := m.getChunk(l)
	t := c.GetTile(l.X, l.Y)
	topObj = t
	if !t.Ignore() {
		avgZ := t.Z() + t.StandingHeight()
		if avgZ == zLimit {
			return t
		} else if avgZ < zLimit {
			topZ = avgZ
		}
		// Else topz is still uo.MapMinZ and we are looking for statics and
		// items underground.
	}
	for _, static := range c.statics {
		// Ignore statics that are not at the location
		if static.Location.X != l.X || static.Location.Y != l.Y {
			continue
		}
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
		// Ignore items that are not at the location
		if item.Location().X != l.X || item.Location().Y != l.Y {
			continue
		}
		// Only consider surface items
		if !item.Surface() && !item.Wet() {
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

// getTerrainElevations returns the bottom, average, and top Z coordinate of the
// tile matrix at the location.
func (m *Map) getTerrainElevations(x, y int16) (lowest, average, highest int8) {
	zTop := m.GetTile(x, y).Z()
	zLeft := m.GetTile(x, y+1).Z()
	zRight := m.GetTile(x+1, y).Z()
	zBottom := m.GetTile(x+1, y+1).Z()
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

// plop attempts to adjust the Z position of the object such that it fits on the
// map, sits above any other objects already at the location, and does not poke
// through a floor or ceiling of some kind. It returns true on success. The
// object's Z position might be altered on success, but not always. The object
// is never altered on failure.
//
// NOTES
//
//	This function enforces the uo.MaxItemStackHeight restriction. If the total
//	height of items at the location would be greater than this limit this
//	function will fail.
//
// Explanation
//  1. Determine the floor altitude considering the tile matrix and statics.
//  2. Determine the ceiling altitude considering the tile matrix and statics.
//  3. If the gap between the floor and ceiling is less than the object
//     height the object is untouched and we return false.
//  4. Process all items to adjust floor and ceiling values.
//  5. If the gap between the floor and ceiling is less than the object height\
//     the object is untouched and we return false. Otherwise the object's Z
//     position is updated to rest on the calculated floor.
func (m *Map) plop(toPlop Item) bool {
	l := toPlop.Location()
	ceiling := uo.MapMaxZ
	floor := l.Z
	c := m.getChunk(l)
	// Consider the tile matrix
	t := c.GetTile(l.X, l.Y)
	if !t.Ignore() {
		avgZ := t.Z() + t.StandingHeight()
		if avgZ < floor {
			// Object is above the ground
			floor = avgZ
		} else if avgZ > floor {
			// Object is below the ground
			ceiling = avgZ
		} // Else the object is already on the ground
	} else {
		// This is a cave entrance or other ignorable tile, so we'll need to
		// select the static surface below the object
		floor = uo.MapMinZ
	}
	// Process statics to find the static floor and ceiling
	for _, static := range c.statics {
		// Ignore statics that are not at the location
		if static.Location.X != l.X || static.Location.Y != l.Y {
			continue
		}
		// Only select surfaces
		if !static.Surface() && !static.Wet() {
			continue
		}
		staticTopZ := static.Z() + static.Height()
		if staticTopZ > floor && staticTopZ <= l.Z {
			// This static is between the tile matrix and the object's starting
			// location so consider it a possible floor
			floor = staticTopZ
		} else if static.Z() < ceiling && static.Z() >= floor {
			// This static is above us so consider it a possible ceiling
			ceiling = static.Z()
		} // Else the static is outside the current gap
	}
	// See if there is enough room to begin with
	if ceiling-floor < toPlop.Height() {
		return false
	}
	// Limit item stack height to the max
	if ceiling-floor > uo.MaxItemStackHeight {
		ceiling = floor + uo.MaxItemStackHeight
	}
	// Process items to adjust floor and ceiling values
	for _, item := range c.items {
		// Only look at items at our location
		if item.Location().X != l.X || item.Location().Y != l.Y {
			continue
		}
		itemTopZ := item.Z() + item.Height()
		if itemTopZ > floor && itemTopZ <= ceiling {
			// The item is between the static floor and the current ceiling so
			// consider it a possible floor
			floor = itemTopZ
		}
		if item.Z() < ceiling && item.Z() > floor {
			// This item is above us so consider it a possible ceiling
			ceiling = item.Z()
		} // Else the item is outside the current gap
	}
	// See if there is still enough room
	if ceiling-floor < toPlop.Height() {
		return false
	}
	// Note this will move the item upward when ploping under a solid stack of
	// items which is the intended behavior
	l.Z = floor
	toPlop.SetLocation(l)
	return true
}

// plopItems adjusts the Z position of every object in the location above the Z
// position and below the ceiling. This has the effect of letting all these
// items fall by the given amount.
//
// NOTES
//
//	As a side-effect this function calls world.Update on each item modified.
//
// Explanation
//  1. Determine the ceiling altitude considering the tile matrix and statics.
//  2. Process all items in the location:
//     a. If the item's Z position is under the ceiling apply the offset.
//     b. Make sure no item's Z position is set lower than the static floor.
func (m *Map) plopItems(l uo.Location, drop int8) {
	ceiling := uo.MapMaxZ
	floor := l.Z
	c := m.getChunk(l)
	// Consider the tile matrix
	t := c.GetTile(l.X, l.Y)
	if !t.Ignore() {
		avgZ := t.Z() + t.StandingHeight()
		if avgZ > floor {
			// Location is below the ground
			ceiling = avgZ
		} // Else the location is already on the ground or above it
	}
	// Process statics to find the static ceiling
	for _, static := range c.statics {
		// Ignore statics that are not at the location
		if static.Location.X != l.X || static.Location.Y != l.Y {
			continue
		}
		// Only select surfaces
		if !static.Surface() && !static.Wet() {
			continue
		}
		if static.Z() < ceiling && static.Z() >= floor {
			ceiling = static.Z()
		}
	}
	// Process items to adjust Z position
	for _, item := range c.items {
		// Only look at items at our location
		if item.Location().X != l.X || item.Location().Y != l.Y {
			continue
		}
		if item.Z() > floor && item.Z() < ceiling {
			il := item.Location()
			il.Z -= drop
			if il.Z < floor {
				il.Z = floor
			}
			item.SetLocation(il)
			world.Update(item)
		}
	}
}

// GetFloorAndCeiling returns the objects that make up the floor below and the
// ceiling above the given reference location. These objects may be any of the
// objects contained within the map such as Tiles, Items, Statics, and Multis.
// A nil return value means that there is no floor below the position, or that
// there is no ceiling above the position. Normally at least one of the return
// values will be non-nil referencing at least the tile matrix. However there
// are certain places on the map - such as cave entrances - where the tile
// matrix is ignored. In these cases both return values may be nil if there are
// no items or statics to create a surface.
func (m *Map) GetFloorAndCeiling(l uo.Location) (uo.CommonObject, uo.CommonObject) {
	floor := uo.MapMinZ
	var floorObject uo.CommonObject
	ceiling := uo.MapMaxZ
	var ceilingObject uo.CommonObject
	footLevel := l.Z
	// Consider tile matrix
	c := m.getChunk(l)
	t := c.GetTile(l.X, l.Y)
	if !t.Ignore() {
		bottom := t.Z()
		avg := bottom + t.StandingHeight()
		if footLevel < bottom {
			// Mobile is below ground
			ceiling = avg
			ceilingObject = t
		} else if footLevel >= avg {
			// Mobile is above or on the ground
			floor = avg
			floorObject = t
		} else {
			// Mobile is down inside a tile in the tile matrix
			floor = avg
			floorObject = t
			footLevel = floor
		}
	}
	// Consider statics
	for _, static := range c.statics {
		// Ignore statics that are not at the location
		if static.Location.X != l.X || static.Location.Y != l.Y {
			continue
		}
		// Only select solid statics. This ignores things like leaves and stuff
		if !static.Surface() && !static.Wet() && !static.Impassable() {
			continue
		}
		sz := static.Z()
		stz := sz + static.Height()
		if stz > floor && stz <= footLevel {
			// Static is between the tile matrix and the mob's feet, so consider
			// it a possible floor.
			floor = stz
			floorObject = static
		} else if sz < ceiling && sz > floor {
			// Static is above us so consider it a possible ceiling.
			ceiling = sz
			ceilingObject = static
		} // Else the static is outside the current gap
	}
	// Consider items
	for _, item := range c.items {
		// Only look at items at our location
		if item.Location().X != l.X || item.Location().Y != l.Y {
			continue
		}
		// Only select solid items. This ignores passible items like gold.
		if !item.Surface() && !item.Wet() && !item.Impassable() {
			continue
		}
		iz := item.Z()
		itz := iz + item.Height()
		if itz > floor && itz <= ceiling {
			// Item is between the static floor and ceiling, consider it a
			// possible floor.
			floor = itz
			floorObject = item
		}
		if iz < ceiling && iz > floor {
			// Item is between the ceiling and the floor, consider it a possible
			// ceiling.
			ceiling = iz
			ceilingObject = item
		} // Else the item is outside the current gap;
	}
	return floorObject, ceilingObject
}
