package game

import "github.com/qbradq/sharduo/lib/uo"

// RegionFeature is a flag that turns on the given feature for a region. Note
// that feature flags can be turned on by a region but not turned off.
type RegionFeature uint16

const (
	RegionFeatureSafeLogout RegionFeature = 0b0000000000000001 // Disables the 10 minute logout wait
	RegionFeatureGuarded    RegionFeature = 0b0000000000000010 // Enables guards
	RegionFeatureNoTeleport RegionFeature = 0b0000000000000100 // Disables teleporting into and out of the region
)

// Region defines a geographical bounding area and data that defines various
// behaviors.
type Region struct {
	Name     string        // Name of the region
	Bounds   uo.Bounds     // Bounds of all bounding rects for first-level inclusion detection
	Rects    []uo.Bounds   // Bounds of all the rects region
	Features RegionFeature // Feature flags for this region
	Music    string        // Music track to play for the region as a range expression
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
