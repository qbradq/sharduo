package game

import (
	"sort"

	"github.com/qbradq/sharduo/lib/uo"
)

// chunk manages the data for a single chunk of the map.
type chunk struct {
	Items   []*Item                                 // List of all items within the chunk
	Mobiles []*Mobile                               // List of all mobiles within the chunk
	Tiles   [uo.ChunkWidth * uo.ChunkHeight]uo.Tile // Tile matrix
	Statics []uo.Static                             // All statics within the chunk
}

// AddMobile adds the mobile to the chunk.
func (c *chunk) AddMobile(m *Mobile) {
	c.Mobiles = append(c.Mobiles, m)
}

// RemoveMobile removes the mobile from the chunk.
func (c *chunk) RemoveMobile(m *Mobile) {
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
}

// AddItem adds the item to the chunk.
func (c *chunk) AddItem(item *Item) {
	c.Items = append(c.Items, item)
	// Keep items Z-sorted
	sort.Slice(c.Items, func(i, j int) bool {
		return c.Items[i].Location.Z < c.Items[j].Location.Z
	})
}

// RemoveItem removes the item from the chunk.
func (c *chunk) RemoveItem(item *Item) {
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
}
