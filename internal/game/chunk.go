package game

import (
	"sort"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// cuBuf is the buffer of item pointers used by [Chunk.Update].
var cuBuf []*Item

// Chunk manages the data for a single Chunk of the map.
type Chunk struct {
	Bounds      uo.Bounds                               // Bounds of the chunk
	Items       []*Item                                 // List of all items within the chunk
	Mobiles     []*Mobile                               // List of all mobiles within the chunk
	Tiles       [uo.ChunkWidth * uo.ChunkHeight]uo.Tile // Tile matrix
	Statics     []uo.Static                             // All statics within the chunk
	Regions     []*Region                               // All regions that overlap this chunk
	Ore         int                                     // Amount of ore left in the chunk
	oreDeadline uo.Time                                 // Next ore respawn
	x, y        int
}

// newChunk creates a new chunk struct ready for use.
func newChunk(cx, cy int) *Chunk {
	return &Chunk{
		Bounds: uo.Bounds{
			X: cx * uo.ChunkWidth,
			Y: cy * uo.ChunkHeight,
			Z: uo.MapMinZ,
			W: uo.ChunkWidth,
			H: uo.ChunkHeight,
			D: uo.MapMaxZ - uo.MapMinZ,
		},
		x: cx,
		y: cy,
	}
}

// AddMobile adds the mobile to the chunk.
func (c *Chunk) AddMobile(m *Mobile) {
	c.Mobiles = append(c.Mobiles, m)
}

// RemoveMobile removes the mobile from the chunk.
func (c *Chunk) RemoveMobile(m *Mobile) {
	idx := -1
	for i, mob := range c.Mobiles {
		if mob == m {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	copy(c.Mobiles[idx:], c.Mobiles[idx+1:])
	c.Mobiles[len(c.Mobiles)-1] = nil
	c.Mobiles = c.Mobiles[:len(c.Mobiles)-1]
}

// AddItem adds the item to the chunk.
func (c *Chunk) AddItem(item *Item) {
	c.Items = append(c.Items, item)
	// Keep items Z-sorted
	sort.Slice(c.Items, func(i, j int) bool {
		return c.Items[i].Location.Z < c.Items[j].Location.Z
	})
}

// RemoveItem removes the item from the chunk.
func (c *Chunk) RemoveItem(item *Item) {
	idx := -1
	for i, oi := range c.Items {
		if oi == item {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	copy(c.Items[idx:], c.Items[idx+1:])
	c.Items[len(c.Items)-1] = nil
	c.Items = c.Items[:len(c.Items)-1]
}

// Update handles the 1-minute periodic update for this chunk.
func (c *Chunk) Update(t uo.Time) {
	// Ore respawn
	if t >= c.oreDeadline {
		c.Ore = util.Random(10, 24)
		c.oreDeadline = t + uo.DurationMinute*uo.Time(util.Random(25, 35))
	}
	// Item updates for items on the ground
	if len(cuBuf) < len(c.Items) {
		cuBuf = make([]*Item, len(c.Items)*2)
	}
	items := cuBuf[:len(c.Items)]
	copy(items, c.Items)
	for _, i := range items {
		i.Update(t)
	}
}

// AddRegion adds the region to this chunk if it overlaps.
func (c *Chunk) AddRegion(r *Region) bool {
	if !c.Bounds.Overlaps(r.Bounds) {
		return false
	}
	for _, region := range c.Regions {
		if region == r {
			return false
		}
	}
	c.Regions = append(c.Regions, r)
	return true
}

// RemoveRegion removes the region pointed to.
func (c *Chunk) RemoveRegion(r *Region) {
	idx := -1
	for i, region := range c.Regions {
		if region == r {
			idx = i
			break
		}
	}
	if idx >= 0 {
		copy(c.Regions[idx:], c.Regions[idx+1:])
		c.Regions[len(c.Regions)-1] = nil
		c.Regions = c.Regions[:len(c.Regions)-1]
	}
}
