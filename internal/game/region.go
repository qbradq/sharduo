package game

import (
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
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

// SpawnedObject describes one object that was spawned.
type SpawnedObject struct {
	Object            Object  // Pointer to the object that was spawned
	NextSpawnDeadline uo.Time // When should the object be spawned again
}

// SpawnerEntry describes one object to spawn.
type SpawnerEntry struct {
	Template string           // Name of the template of the object
	Amount   int              // Amount of objects to spawn in the area
	Delay    uo.Time          // Delay between object disappearance and respawn
	Objects  []*SpawnedObject // Pointers to the spawned objects
}

// Region defines a geographical bounding area and data that defines various
// behaviors.
type Region struct {
	Name      string          // Name of the region
	Bounds    uo.Bounds       // Bounds of all bounding rects for first-level inclusion detection
	Rects     []uo.Bounds     // Bounds of all the rects region
	Features  RegionFeature   // Feature flags for this region
	Music     string          // Music track to play for the region as a range expression
	Entries   []*SpawnerEntry // All of the entries for region spawning
	SpawnMinZ int8            // Minimum Z position for spawned objects
	SpawnMaxZ int8            // Maximum Z position for spawned objects
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
func (r *Region) Contains(l uo.Location) bool {
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
	for _, e := range r.Entries {
		for _, o := range e.Objects {
			if o.Object == nil {
				// Try to respawn
				if t >= o.NextSpawnDeadline {
					o.Object = r.Spawn(e.Template)
				}
			} else {
				if o.Object.Removed() {
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
	for _, e := range r.Entries {
		for _, o := range e.Objects {
			Remove(o.Object)
		}
		e.Objects = make([]*SpawnedObject, e.Amount)
		for i := range e.Objects {
			e.Objects[i] = &SpawnedObject{
				Object: r.Spawn(e.Template),
			}
		}
	}
}

// Spawn spawns an object by template name and returns a pointer to it or nil.
func (r *Region) Spawn(which string) Object {
	// List expansion
	tn := which
	if len(tn) > 0 && tn[0] == '+' {
		tn = template.RandomListMember(tn[1:])
	}
	// Object creation
	o := template.Create[Object](tn)
	if o == nil {
		return nil
	}
	o.SetSpawnerRegion(r)
	// Object placement
	for i := 0; i < 8; i++ {
		l := r.Bounds.RandomLocation(world.Random())
		if !r.Contains(l) {
			// The location is not within one of our rects, try again
			continue
		}
		l.Z = r.SpawnMinZ
		s := world.Map().GetSpawnableSurface(l, r.SpawnMaxZ, o)
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
		o.SetLocation(l)
		if !world.Map().AddObject(o) {
			// Failed to add to the map for some reason, don't leak the object
			// just remove it and try again
			Remove(o)
			continue
		}
		// If we get here we have successfully placed the object on the map and
		// can return it
		return o
	}
	// If we reach this point we have ran out of tries to spawn something and
	// give up
	return nil
}

// ReleaseObject releases the given object from this spawner so another may
// spawn in its place.
func (r *Region) ReleaseObject(o Object) {
	if o == nil {
		return
	}
	for _, e := range r.Entries {
		for _, oo := range e.Objects {
			if oo.Object.Serial() == o.Serial() {
				oo.Object = nil
				oo.NextSpawnDeadline = world.Time() + e.Delay
				return
			}
		}
	}
}
