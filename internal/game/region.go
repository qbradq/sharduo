package game

import (
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// RegionFeature is a flag that turns on the given feature for a region. Note
// that feature flags can be turned on by a region but not turned off.
type RegionFeature uint16

const (
	RegionFeatureSafeLogout    RegionFeature = 0b0000000000000001 // Disables the 10 minute logout wait
	RegionFeatureGuarded       RegionFeature = 0b0000000000000010 // Enables guards
	RegionFeatureNoTeleport    RegionFeature = 0b0000000000000100 // Disables teleporting into and out of the region
	RegionFeatureSpawnOnGround RegionFeature = 0b0000000000001000 // Forces spawned objects to follow the terrain
)

// spawnedObject describes one object that was spawned.
type spawnedObject struct {
	Object            any     // Pointer to the object that was spawned
	NextSpawnDeadline uo.Time // When should the object be spawned again
}

// SpawnerEntry describes one object to spawn.
type SpawnerEntry struct {
	Template string           // Name of the template of the object
	Amount   int              // Amount of objects to spawn in the area
	Delay    uo.Time          // Delay between object disappearance and respawn
	objects  []*spawnedObject // Pointers to the spawned objects
}

// RemoveObjects removes all objects that have been spawned by this entry.
func (e *SpawnerEntry) RemoveObjects() {
	for _, obj := range e.objects {
		switch o := obj.Object.(type) {
		case *Mobile:
			World.Map().RemoveMobile(o)
			World.RemoveMobile(o)
		case *Item:
			o.Remove()
			World.RemoveItem(o)
		}
	}
}

// Region defines a geographical bounding area and data that defines various
// behaviors.
type Region struct {
	Name      string          // Name of the region
	Bounds    uo.Bounds       `json:"-"` // Bounds of all bounding rects for first-level inclusion detection
	Rects     []uo.Bounds     // Bounds of all the rects region
	Features  RegionFeature   // Feature flags for this region
	Music     string          // Music track to play for the region as a range expression
	Spawns    []*SpawnerEntry // All of the entries for region spawning
	SpawnMinZ int             // Minimum Z position for spawned objects
	SpawnMaxZ int             // Maximum Z position for spawned objects
}

// AddRect adds a rect to the region.
func (r *Region) AddRect(b uo.Bounds) {
	if r.Bounds == uo.BoundsZero {
		r.Bounds = b
	} else {
		r.Bounds = uo.BoundsFit(r.Bounds, b)
	}
	r.Rects = append(r.Rects, b)
}

// ForceRecalculateBounds recalculates the Bounds member variable.
func (r *Region) ForceRecalculateBounds() {
	if len(r.Rects) < 1 {
		r.Bounds = uo.BoundsZero
		return
	}
	r.Bounds = r.Rects[0]
	for i := 1; i < len(r.Rects); i++ {
		r.Bounds = uo.BoundsFit(r.Bounds, r.Rects[i])
	}
}

// Contains returns true if the location is contained within one of the rects of
// the region.
func (r *Region) Contains(l uo.Point) bool {
	if !r.Bounds.Contains(l) {
		return false
	}
	for _, b := range r.Rects {
		if b.Contains(l) {
			return true
		}
	}
	return false
}

// Overlaps returns true if the given bounds overlap with any of the rects of
// the region.
func (r *Region) Overlaps(b uo.Bounds) bool {
	if !r.Bounds.Overlaps(b) {
		return false
	}
	for _, a := range r.Rects {
		if a.Overlaps(b) {
			return true
		}
	}
	return false
}

// Update is the periodic update function for the region.
func (r *Region) Update(t uo.Time) {
	// Process all owned objects
	for _, e := range r.Spawns {
		for _, o := range e.objects {
			if o.Object == nil {
				// Try to respawn
				if t >= o.NextSpawnDeadline {
					o.Object = r.Spawn(e.Template)
				}
			} else {
				removed := false
				switch t := o.Object.(type) {
				case *Mobile:
					removed = t.Removed
				case *Item:
					removed = t.Removed
				}
				if removed {
					// Schedule next respawn
					o.Object = nil
					o.NextSpawnDeadline = t + e.Delay
				}
			}
		}
	}
}

// FullRespawn removes all objects spawned by this region and then fully
// respawns all entries.
func (r *Region) FullRespawn() {
	for _, e := range r.Spawns {
		for _, o := range e.objects {
			switch t := o.Object.(type) {
			case *Mobile:
				World.Map().RemoveMobile(t)
				World.RemoveMobile(t)
			case *Item:
				t.Remove()
			}
		}
		e.objects = make([]*spawnedObject, e.Amount)
		for i := range e.objects {
			e.objects[i] = &spawnedObject{
				Object:            r.Spawn(e.Template),
				NextSpawnDeadline: World.Time() + e.Delay,
			}
		}
	}
}

// Spawn spawns an object by template name and returns a pointer to it or nil.
func (r *Region) Spawn(which string) any {
	// List expansion
	tn := which
	if len(tn) > 0 && tn[0] == '+' {
		tn = ListMember(tn[1:]).String()
	}
	// Object creation
	var m *Mobile
	var item *Item
	m = NewMobile(tn)
	if m == nil {
		item = NewItem(tn)
		if item == nil {
			return nil
		}
		item.Spawner = r
	} else {
		m.Spawner = r
	}
	// Object placement
	for i := 0; i < 8; i++ {
		l := util.RandomLocation(r.Bounds)
		if !r.Contains(l) {
			// The location is not within one of our rects, try again
			continue
		}
		l.Z = r.SpawnMinZ
		var s uo.CommonObject
		if m != nil {
			s = World.Map().GetSpawnableSurface(l, r.SpawnMaxZ, uo.PlayerHeight)
		} else {
			s = World.Map().GetSpawnableSurface(l, r.SpawnMaxZ, item.Def.Height)
		}
		if s == nil {
			// The location is not suitable for spawning, try again
			continue
		}
		if r.Features&RegionFeatureSpawnOnGround != 0 {
			// Forced ground spawning
			if _, ok := s.(uo.Tile); !ok {
				// The given floor is not from the tile matrix, reject it
				continue
			}
		}
		l.Z = s.StandingHeight()
		if m != nil {
			m.Location = l
			if !World.Map().AddMobile(m, false) {
				// Failed to place the object, just try again
				continue
			}
			return m
		} else {
			item.Location = l
			if !World.Map().AddItem(item, false) {
				// Failed to place the object, just try again
				continue
			}
			return item
		}
	}
	// If we reach this point we have ran out of tries to spawn something and
	// give up
	return nil
}

// ReleaseObject releases the given object from this spawner so another may
// spawn in its place.
func (r *Region) ReleaseObject(o any) {
	if o == nil {
		return
	}
	for _, e := range r.Spawns {
		for _, oo := range e.objects {
			if oo.Object == o {
				oo.Object = nil
				oo.NextSpawnDeadline = World.Time() + e.Delay
				return
			}
		}
	}
}
