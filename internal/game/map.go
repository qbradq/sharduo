package game

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/uo/file"
)

// Map contains the tile matrix, static items, and all dynamic objects of a map.
type Map struct {
	chunks      []*Chunk             // The chunks of the map
	regions     []*Region            // A list of all of the regions of the map
	deepStorage map[uo.Serial]Object // Deep storage for objects like stabled pets and logged out characters
}

// NewMap creates and returns a new Map
func NewMap() *Map {
	m := &Map{
		chunks:      make([]*Chunk, uo.MapChunksWidth*uo.MapChunksHeight),
		deepStorage: make(map[uo.Serial]Object),
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
			m.GetChunk(uo.Location{X: int16(ix), Y: int16(iy)}).setTile(ix, iy, mapmul.GetTile(ix, iy))
		}
	}
	// Pre-calculate tile elevations
	for iy := int16(0); iy < int16(uo.MapHeight); iy++ {
		for ix := int16(0); ix < int16(uo.MapWidth); ix++ {
			c := m.GetChunk(uo.Location{X: ix, Y: iy})
			t := c.GetTile(ix, iy)
			lowest, avg, height := m.getTerrainElevations(ix, iy)
			t = t.SetElevations(lowest, avg, height)
			c.setTile(int(ix), int(iy), t)
		}
	}
	// Load the statics
	for _, static := range staticsmul.Statics() {
		c := m.GetChunk(static.Location)
		c.statics = append(c.statics, static)
	}
	// Sort statics by bottom Z
	for iy := int16(0); iy < int16(uo.MapChunksHeight); iy++ {
		for ix := int16(0); ix < int16(uo.MapChunksWidth); ix++ {
			c := m.GetChunk(uo.Location{X: ix * int16(uo.ChunkWidth), Y: iy * int16(uo.ChunkHeight)})
			sort.Slice(c.statics, func(i, j int) bool {
				si := c.statics[i]
				sj := c.statics[j]
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

// MarshalObjects writes out all objects that are directly on the map split into
// pools to facilitate multi-goroutine saving.
func (m *Map) MarshalObjects(wg *sync.WaitGroup, s *marshal.TagFileSegment, pool, pools int) {
	defer wg.Done()
	l := len(m.chunks)
	for i := pool; i < l; i += pools {
		c := m.chunks[i]
		for _, o := range c.items {
			if o.Removed() || o.NoRent() || o.SpawnerRegion() != nil {
				continue
			}
			s.PutObject(o)
			s.IncrementRecordCount()
		}
		for _, o := range c.mobiles {
			if o.Removed() || o.NoRent() || o.SpawnerRegion() != nil {
				continue
			}
			s.PutObject(o)
			s.IncrementRecordCount()
		}
	}
}

// Marshal writes out top-level map information
func (m *Map) Marshal(wg *sync.WaitGroup, s *marshal.TagFileSegment) {
	defer wg.Done()
	// Ore map
	for _, c := range m.chunks {
		s.PutByte(byte(c.ore))
	}
}

// UnmarshalObjects unmarshals all of the objects directly parented to the map.
func (m *Map) UnmarshalObjects(s *marshal.TagFileSegment) {
	// Place all map objects onto the map
	for i := uint32(0); i < s.RecordCount(); i++ {
		oum := s.Object()
		o, ok := oum.(Object)
		if !ok {
			panic("map object did not implement the Object interface")
		}
		c := m.GetChunk(o.Location())
		c.Add(o)
	}
}

// AfterUnmarshal calls AfterUnmarshalOntoMap calls for all map objects. We do
// this with a pre-compiled list of objects so that calls to
// AfterUnmarshalOntoMap can call world.Remove() if needed.
func (m *Map) AfterUnmarshal() {
	var objs []Object
	for _, c := range m.chunks {
		for _, item := range c.items {
			objs = append(objs, item)
		}
		for _, mobile := range c.mobiles {
			objs = append(objs, mobile)
		}
	}
	for _, o := range objs {
		o.AfterUnmarshalOntoMap()
	}
}

// Unmarshal reads in top-level map information
func (m *Map) Unmarshal(s *marshal.TagFileSegment) {
	// Load the ore map data
	for _, c := range m.chunks {
		c.ore = uint8(s.Byte())
	}
}

// MarshalDeepStorage writes out the off-map objects
func (m *Map) MarshalDeepStorage(wg *sync.WaitGroup, s *marshal.TagFileSegment) {
	defer wg.Done()
	for _, o := range m.deepStorage {
		s.PutObject(o)
		s.IncrementRecordCount()
	}
}

// UnmarshalDeepStorage unmarshals all of the objects in map deep storage.
func (m *Map) UnmarshalDeepStorage(s *marshal.TagFileSegment) {
	for i := uint32(0); i < s.RecordCount(); i++ {
		oum := s.Object()
		o, ok := oum.(Object)
		if !ok {
			panic("deep storage object did not implement the Object interface")
		}
		m.deepStorage[o.Serial()] = o
	}
}

// GetChunk returns a pointer to the chunk for the given location.
func (m *Map) GetChunk(l uo.Location) *Chunk {
	l = l.Bound()
	cx := int(l.X) / uo.ChunkWidth
	cy := int(l.Y) / uo.ChunkHeight
	return m.chunks[cy*uo.MapChunksWidth+cx]
}

// GetTile returns the Tile value for the location
func (m *Map) GetTile(x, y int16) uo.Tile {
	return m.GetChunk(uo.Location{
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
	if o.SpawnerRegion() != nil {
		o.SpawnerRegion().ReleaseObject(o)
		o.SetSpawnerRegion(nil)
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
	c := m.GetChunk(o.Location())
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
	// tell it. This is the case when a mobile is first created from template
	// and when loading from save.
	if ok && mob.StandingOn() == nil {
		floor, _ := m.GetFloorAndCeiling(o.Location(), false, false)
		mob.StandOn(floor)
	}
	// If this is an item we need to update the decay deadline.
	if item, ok := o.(Item); ok {
		item.RefreshDecayDeadline()
	}
}

// ForceRemoveObject removes the object from the map and always succeeds.
func (m *Map) ForceRemoveObject(o Object) {
	c := m.GetChunk(o.Location().Bound())
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
	nl.Z = mob.StandingOn().Highest()
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
		oldFloor := mob.StandingOn()
		oldTop := oldFloor.Highest()
		if !floor.Bridge() && fz-oldTop > uo.StepHeight {
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
	dir = dir.Bound().StripRunningFlag()
	if mob.Facing() != dir {
		// Change facing request
		mob.SetFacing(dir)
		for _, othermob := range m.GetNetStatesInRange(mob.Location(), uo.MaxViewRange+1) {
			othermob.NetState().MoveMobile(mob)
		}
		return true
	} // else move request
	// Stamina check
	if mob.Stamina() <= 0 {
		if mob.NetState() != nil {
			mob.NetState().Cliloc(nil, 500110) // You are too fatigued to move.
			return false
		}
	}
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
	oldChunk := m.GetChunk(oldLocation)
	newChunk := m.GetChunk(newLocation)
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
	otherMobs := m.GetNetStatesInRange(mob.Location(), uo.MaxViewRange+1)
	for _, otherMob := range otherMobs {
		if oldLocation.XYDistance(otherMob.Location()) <= otherMob.ViewRange() &&
			newLocation.XYDistance(otherMob.Location()) > otherMob.ViewRange() {
			// We used to be in visible range of the other mobile but are
			// moving out of that range, delete us
			otherMob.NetState().RemoveObject(mob)
		} else if oldLocation.XYDistance(otherMob.Location()) > otherMob.ViewRange() &&
			newLocation.XYDistance(otherMob.Location()) <= otherMob.ViewRange() {
			// We used to be outside the visible range of the other mobile but
			// are moving into that range, send us
			otherMob.NetState().SendObject(mob)
		} else {
			otherMob.NetState().MoveMobile(mob)
		}
	}
	// Update mobile standing
	mob.StandOn(floor)
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
	if !world.Map().AddObject(mob) {
		// Map.AddObject() checks height
		// Don't leak the mobile, just force it back where it came from.
		mob.SetLocation(oldLocation)
		world.Map().ForceAddObject(mob)
		floor, _ := m.GetFloorAndCeiling(mob.Location(), false, false)
		mob.StandOn(floor)
		return false
	}
	// Update standing
	floor, _ := m.GetFloorAndCeiling(mob.Location(), false, false)
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

func (m *Map) getChunksInBounds(b uo.Bounds) []*Chunk {
	var ret []*Chunk
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
func (m *Map) getChunksInRange(l uo.Location, r int16) []*Chunk {
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

// getTerrainElevations returns the bottom, average, and top Z coordinate of the
// tile matrix at the location.
func (m *Map) getTerrainElevations(x, y int16) (lowest, average, highest int8) {
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
//  1. Query floor and ceiling
//  2. Process all items to find top of stack relative to the floor
//  3. If the gap between the top of the stack and either the ceiling or the
//     stack height limit is less than the object height the object is untouched
//     and we return false. Otherwise the object's Z position is updated to rest
//     on top of the stack.
func (m *Map) plop(toPlop Item) bool {
	l := toPlop.Location()
	floor, ceiling := m.GetFloorAndCeiling(l, true, false)
	if floor == nil {
		// No floor to place item on, bail
		return false
	}
	fz := floor.StandingHeight()
	cz := uo.MapMaxZ
	if ceiling != nil {
		cz = ceiling.Z()
	}
	// Find the Z level that would make this item set on top of other non-solid
	// items at the same location
	tz := fz
	c := m.GetChunk(l)
	for _, item := range c.items {
		// Only look at items at our location
		if item.Location().X != l.X || item.Location().Y != l.Y {
			continue
		}
		iz := item.Z()
		itz := item.Highest()
		if itz < tz {
			// Item is below the stack so go to the next item
			continue
		}
		if iz < cz {
			// The item is between the current stack top and the current ceiling
			// so it is the new top item
			tz = itz
			continue
		}
		// Else the item's underside is above the ceiling so we have hit the end
		// of the stack
		break
	}
	// Limit item stack height to the max
	if (tz-fz)+toPlop.Height() > uo.MaxItemStackHeight {
		return false
	}
	// Note this will move the item upward when ploping under a solid stack of
	// items which is the intended behavior
	l.Z = tz
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
	_, ceiling := m.GetFloorAndCeiling(l, true, false)
	fz := l.Z
	cz := uo.MapMaxZ
	if ceiling != nil {
		cz = ceiling.Z()
	}
	// Process items to adjust Z position
	c := m.GetChunk(l)
	for _, item := range c.items {
		// Only look at items at our location
		if item.Location().X != l.X || item.Location().Y != l.Y {
			continue
		}
		iz := item.Z()
		if iz < fz {
			// Item is below the reference point, ignore it
			continue
		}
		if iz >= cz {
			// Item is on or above the ceiling so we are done
			break
		}
		// Item is in the stack
		il := item.Location()
		il.Z -= drop
		if il.Z < fz {
			il.Z = fz
		}
		item.SetLocation(il)
		world.Update(item)
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
// no items or statics to create a surface. If the ignoreDynamicItems argument
// is true then only Items of concrete type *StaticItem are considered. If the
// considerStepHeight parameter is true then gaps less than or equal to
// uo.StepHeight will be ignored.
//
// NOTE: This function requires that all statics and items are z-sorted bottom
// to top.
func (m *Map) GetFloorAndCeiling(l uo.Location, ignoreDynamicItems, considerStepHeight bool) (uo.CommonObject, uo.CommonObject) {
	var floorObject uo.CommonObject
	var ceilingObject uo.CommonObject
	floor := int(uo.MapMinZ)
	ceiling := int(uo.MapMaxZ)
	footHeight := int(l.Z)
	// Consider tile matrix
	c := m.GetChunk(l)
	t := c.GetTile(l.X, l.Y)
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
	for _, static := range c.statics {
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
	for _, item := range c.items {
		// Ignore dynamic items if requested
		if ignoreDynamicItems {
			_, isStatic := item.(*StaticItem)
			if !isStatic {
				continue
			}
		}
		// Only look at items at our location
		if item.Location().X != l.X || item.Location().Y != l.Y {
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

// PlaySound plays a sound at the given location for all clients in range.
func (m *Map) PlaySound(which uo.Sound, from uo.Location) {
	for _, mob := range m.GetNetStatesInRange(from, uo.MaxUpdateRange) {
		if mob.Location().XYDistance(from) <= mob.ViewRange() {
			mob.NetState().Sound(which, from)
		}
	}
}

// PlayEffect plays a graphic effect with the given parameters for all clients
// in range.
func (m *Map) PlayEffect(t uo.GFXType, src, trg Object, gfx uo.Graphic, speed, duration uint8, fixed, explodes bool, hue uo.Hue, bm uo.GFXBlendMode) {
	var sl uo.Location
	ss := uo.SerialMobileNil
	if src != nil {
		ss = src.Serial()
		sl = src.Location()
	}
	var tl uo.Location
	ts := uo.SerialMobileNil
	if trg != nil {
		ts = trg.Serial()
		tl = trg.Location()
	}
	p := &serverpacket.GraphicalEffect{
		GFXType:        t,
		Source:         ss,
		Target:         ts,
		Graphic:        gfx,
		SourceLocation: sl,
		TargetLocation: tl,
		Speed:          speed,
		Duration:       duration,
		Fixed:          fixed,
		Explodes:       explodes,
		Hue:            hue,
		GFXBlendMode:   bm,
	}
	l := trg.Location()
	for _, mob := range m.GetNetStatesInRange(l, uo.MaxUpdateRange) {
		if mob.Location().XYDistance(l) <= mob.ViewRange() {
			mob.NetState().Send(p)
		}
	}
}

// PlayAnimation plays an animation for a mobile.
func (m *Map) PlayAnimation(who Mobile, at uo.AnimationType, aa uo.AnimationAction) {
	p := &serverpacket.Animation{
		Serial:          who.Serial(),
		AnimationType:   at,
		AnimationAction: aa,
	}
	l := who.Location()
	for _, mob := range m.GetNetStatesInRange(l, uo.MaxUpdateRange) {
		if mob.Location().XYDistance(l) <= mob.ViewRange() {
			mob.NetState().Send(p)
		}
	}
}

// SendSpeech sends speech to all mobiles in range.
func (m *Map) SendSpeech(from Object, r int16, format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	mobs := m.GetMobilesInRange(from.Location(), r)
	sort.Slice(mobs, func(i, j int) bool {
		return mobs[i].Location().XYDistance(from.Location()) <
			mobs[j].Location().XYDistance(from.Location())
	})
	isAllCommand := len(text) >= 4 && strings.ToLower(text[:4]) == "all "
	speechEventHandled := false
	for _, mob := range mobs {
		if from.Location().XYDistance(mob.Location()) <= mob.ViewRange() {
			if mob.NetState() != nil {
				mob.NetState().Speech(from, text)
			}
			// Make sure we don't trigger every listener in range, just the
			// closest, unless it is an "all" command.
			if !speechEventHandled {
				speechEventHandled = DynamicDispatch("Speech", mob, from, text) && !isAllCommand
			}
		}
	}
}

// SendCliloc sends cliloc speech to all mobiles in range.
func (m *Map) SendCliloc(from Object, r int16, c uo.Cliloc, args ...string) {
	for _, mob := range m.GetMobilesInRange(from.Location(), r) {
		if from.Location().XYDistance(mob.Location()) <= mob.ViewRange() {
			if mob.NetState() != nil {
				mob.NetState().Cliloc(from, c, args...)
			}
		}
	}
}

// Update calls Update on a few chunks every tick such that every chunk gets an
// Update call once every real-world minute or twelve in-game minutes. It also
// calls Update on a few regions every tick such that every region is updated
// over fifteen real-world seconds or three in-game minutes.
func (m *Map) Update(t uo.Time) {
	// Interleaved chunk updates, updates every chunk over a minute
	nChunks := uint64(uo.MapChunksWidth * uo.MapChunksHeight)
	step := uint64(uo.DurationMinute)
	start := uint64(t % uo.Time(step))
	for idx := start; idx < nChunks; idx += step {
		m.chunks[idx].Update(t)
	}
	// Interleaved region updates, updates every region over fifteen seconds
	nRegions := uint64(len(m.regions))
	step = uint64(uo.DurationSecond * 15)
	start = uint64(t % uo.Time(step))
	for idx := start; idx < nRegions; idx += step {
		m.regions[idx].Update(t)
	}
	// Update all mobiles
	var mobs []Mobile
	for _, c := range m.chunks {
		mobs = append(mobs, c.mobiles...)
	}
	for _, mob := range mobs {
		mob.Update(t)
	}
}

// ConsumeOre attempts to consume the specified amount of ore from the chunk at
// the specified location and returns the number of ore piles consumed.
func (m *Map) ConsumeOre(l uo.Location, n int) int {
	if n < 1 || n > 255 {
		return 0
	}
	amount := uint8(n)
	c := m.GetChunk(l)
	if amount > c.ore {
		amount = c.ore
		c.ore = 0
	} else {
		c.ore -= amount
	}
	return int(amount)
}

// HasOre returns true if the chunk at the given location has any ore.
func (m *Map) HasOre(l uo.Location) bool { return m.GetChunk(l).ore != 0 }

// ItemQuery returns a slice of all of the items matching the given template
// name. The second parameter may be the zero value, in which case the entire
// map is searched. WARNING: This can be expensive and will hang the server.
func (m *Map) ItemQuery(tn string, bounds uo.Bounds) []Item {
	var ret []Item
	if bounds == uo.BoundsZero {
		// Full map query
		for _, c := range m.chunks {
			for _, i := range c.items {
				if i.TemplateName() == tn {
					ret = append(ret, i)
				}
			}
		}
		return ret
	}
	// Spacial query
	chunks := m.getChunksInBounds(bounds)
	for _, c := range chunks {
		for _, item := range c.items {
			if !bounds.Contains(item.Location()) {
				continue
			}
			if item.TemplateName() == tn {
				ret = append(ret, item)
			}
		}
	}
	return ret
}

// ItemBaseQuery returns a slice of all of the items who's direct BaseTemplate
// property matches the given template name. The second parameter may be the
// zero value, in which case the entire map is searched. WARNING: This can be
// expensive and will hang the server.
func (m *Map) ItemBaseQuery(tn string, bounds uo.Bounds) []Item {
	var ret []Item
	if bounds == uo.BoundsZero {
		// Full map query
		for _, c := range m.chunks {
			for _, i := range c.items {
				if i.BaseTemplate() == tn {
					ret = append(ret, i)
				}
			}
		}
		return ret
	}
	// Spacial query
	chunks := m.getChunksInBounds(bounds)
	for _, c := range chunks {
		for _, item := range c.items {
			if !bounds.Contains(item.Location()) {
				continue
			}
			if item.BaseTemplate() == tn {
				ret = append(ret, item)
			}
		}
	}
	return ret
}

// StoreObject places an object into deep storage.
func (m *Map) StoreObject(o Object) {
	m.SetNewParent(o, TheVoid)
	m.deepStorage[o.Serial()] = o
}

// RetrieveObject retrieves and object from deep storage.
func (m *Map) RetrieveObject(s uo.Serial) Object {
	o, found := m.deepStorage[s]
	if found {
		delete(m.deepStorage, s)
	}
	return o
}

// GetSpawnableSurface returns the surface on which something should be spawned
// in the given location, or nil if no suitable surface for the object was
// found. If the parameter object is nil, the height of the object is assumed
// to be 0.
func (m *Map) GetSpawnableSurface(l uo.Location, maxZ int8, o Object) uo.CommonObject {
	f, c := m.GetFloorAndCeiling(l, false, true)
	if f == nil {
		// If we are below the ground project upward to the next surface and try
		// again
		nl := l
		nl.Z = c.Highest()
		f, c = m.GetFloorAndCeiling(nl, false, true)
	}
	if f == nil {
		// Failed to find a valid floor
		return nil
	}
	// Flag checks
	if !f.Surface() || f.Impassable() {
		return nil
	}
	// Z check
	if f.StandingHeight() < l.Z || f.StandingHeight() > maxZ {
		return nil
	}
	// Height check
	if c == nil {
		return f
	}
	oh := 0
	if o != nil {
		if _, ok := o.(Mobile); ok {
			oh = int(uo.PlayerHeight)
		} else if item, ok := o.(Item); ok {
			oh = int(item.Height())
		}
	}
	if int(c.Z())-int(f.StandingHeight()) < oh {
		return nil
	}
	return f
}

// CanFit returns true if the object can fit between the statics and items in
// the given new location.
func (m *Map) CanFit(o Object, l uo.Location) bool {
	z := l.Z
	f, c := m.GetFloorAndCeiling(l, false, false)
	if f != nil {
		tz := f.Z() + f.Height()
		if tz > z {
			// Object had to be projected up through something
			return false
		}
	}
	if c != nil {
		h := uo.PlayerHeight
		if i, ok := o.(Item); ok {
			h = i.Height()
		}
		bz := c.Z()
		if bz < z+h {
			// Ceiling is too low to fit the object
			return false
		}
	}
	return true
}

// StaticsAt returns a slice of the statics and static items at the given
// location.
func (m *Map) StaticsAt(l uo.Location) []uo.CommonObject {
	var ret []uo.CommonObject
	c := m.GetChunk(l)
	for _, s := range c.statics {
		if s.Location.X != l.X || s.Location.Y != l.Y {
			continue
		}
		ret = append(ret, s)
	}
	for _, i := range c.items {
		il := i.Location()
		if il.X != l.X || il.Y != l.Y {
			continue
		}
		ret = append(ret, i)
	}
	return ret
}

// AddRegion adds the given region to the map.
func (m *Map) AddRegion(r *Region) {
	m.regions = append(m.regions, r)
	for _, c := range m.getChunksInBounds(r.Bounds) {
		c.AddRegion(r)
	}
}

// RemoveRegion removes the given region from the map.
func (m *Map) RemoveRegion(r *Region) {
	idx := -1
	for i, region := range m.regions {
		if r == region {
			idx = i
			break
		}
	}
	if idx >= 0 {
		m.regions = append(m.regions[:idx], m.regions[idx+1:]...)
	}
	for _, c := range m.getChunksInBounds(r.Bounds) {
		c.RemoveRegion(r)
	}
}

// RegionsAt returns a slice of all of the regions that overlap the given
// location.
func (m *Map) RegionsAt(l uo.Location) []*Region {
	var ret []*Region
	c := m.GetChunk(l)
	for _, r := range c.regions {
		if r.Contains(l) {
			ret = append(ret, r)
		}
	}
	return ret
}

// RegionsWithin returns a slice of all of the regions that overlap the given
// bounds.
func (m *Map) RegionsWithin(b uo.Bounds) []*Region {
	rset := make(map[*Region]struct{})
	for _, c := range m.getChunksInBounds(b) {
		for _, r := range c.regions {
			if !r.Overlaps(b) {
				continue
			}
			rset[r] = struct{}{}
		}
	}
	ret := make([]*Region, len(rset))
	i := 0
	for r := range rset {
		ret[i] = r
		i++
	}
	return ret
}

// RegionFeaturesAt returns the accumulated region features from all regions at
// the given location.
func (m *Map) RegionFeaturesAt(l uo.Location) RegionFeature {
	var ret RegionFeature
	c := m.GetChunk(l)
	for _, r := range c.regions {
		if r.Contains(l) {
			ret |= r.Features
		}
	}
	return ret
}
