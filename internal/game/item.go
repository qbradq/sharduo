package game

import (
	"fmt"
	"io"

	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
	"golang.org/x/image/colornames"
)

// ItemFlags encode boolean information about an item.
type ItemFlags uint64

const (
	ItemFlagsContainer ItemFlags = 0x0000000000000001 // Item is a container
	ItemFlagsFixed     ItemFlags = 0x0000000000000002 // Item is fixed within the world and cannot be moved normally
	ItemFlagsStatic    ItemFlags = 0x0000000000000004 // Item is static
)

// Item describes any item in the world.
type Item struct {
	Object
	// Static variables
	Flags       ItemFlags            // Boolean item flags
	Graphic     uo.Graphic           // Base graphic to use for the item
	Layer       uo.Layer             // Layer the object is worn on
	Weight      float64              // Weight of the item, NOT the stack, just one of these items
	MaxWeight   float64              // Maximum weight that can be held in this container
	MaxItems    int                  // Maximum number of items that can be held in this container
	GUMPGraphic uo.GUMP              // GUMP graphic to use for containers
	def         *uo.StaticDefinition // Item properties pointer
	// Persistent variables
	Amount   int     // Stack amount
	Contents []*Item // Contents of the container
	// Transient values
	Wearer          *Mobile                 // Pointer to the mobile currently wearing this item if any, note this only indicates if the item is directly equipped to the mobile, not within equipped containers
	Container       *Item                   // Pointer to the parent container if any
	Observers       []ContainerObserver     // All observers currently observing this container
	ItemCount       int                     // Cache of the total number of contained items including all sub-containers
	Gold            int                     // Cache of the total amount of gold coins contained in this and all sub containers
	ContainedWeight float64                 // Cache of the total weight held in this and all sub containers
	opl             *serverpacket.OPLPacket // Cached OPLPacket
	oplInfo         *serverpacket.OPLInfo   // Cached OPLInfo
}

// Write writes the persistent data of the item to w.
func (i *Item) Write(w io.Writer) {
	i.Object.Write(w)
	util.PutUInt32(w, 0)                       // Version
	util.PutUInt16(w, uint16(i.Amount))        // Stack amount
	util.PutUInt32(w, uint32(len(i.Contents))) // Contents
	for _, item := range i.Contents {
		item.Write(w)
	}
}

// Read reads the persistent data of the item from r.
func (i *Item) Read(r io.Reader) {
	i.Object.Read(r)
	_ = util.GetUInt32(r)             // Version
	i.Amount = int(util.GetUInt16(r)) // Stack amount
	n := int(util.GetUInt32(r))       // Contents item count
	i.Contents = make([]*Item, n)     // Contents
	for idx := 0; idx < n; idx++ {
		item := &Item{}
		item.Read(r)
		i.Contents[idx] = item
	}
}

// RecalculateStats recalculates all internal cache states.
func (i *Item) RecalculateStats() {
	i.ItemCount = len(i.Contents)
	i.Gold = 0
	i.ContainedWeight = 0
	if !i.HasFlags(ItemFlagsContainer) {
		return
	}
	for _, c := range i.Contents {
		i.ContainedWeight += c.Weight
		if c.HasFlags(ItemFlagsContainer) {
			c.RecalculateStats()
			i.ItemCount += c.ItemCount
			i.Gold += c.Gold
			i.ContainedWeight += c.ContainedWeight
		}
	}
}

// AddObserver adds a ContainerObserver to the list of current observers.
func (i *Item) AddObserver(o ContainerObserver) {
	for _, co := range i.Observers {
		if co == o {
			return
		}
	}
	i.Observers = append(i.Observers, o)
}

// RemoveObserver removes the ContainerObserver from the list of current
// observers.
func (i *Item) RemoveObserver(o ContainerObserver) {
	idx := -1
	for ii, co := range i.Observers {
		if co == o {
			idx = ii
			break
		}
	}
	if idx >= 0 {
		i.Observers = append(i.Observers[:idx], i.Observers[idx+1:]...)
	}
}

// HasFlags returns true if all of the given flags is set on this item.
func (i *Item) HasFlags(f ItemFlags) bool {
	return i.Flags&f == f
}

// OPLPackets constructs new OPL packets if needed and returns cached packets.
func (i *Item) OPLPackets() (*serverpacket.OPLPacket, *serverpacket.OPLInfo) {
	if i.opl == nil {
		i.opl = &serverpacket.OPLPacket{
			Serial: i.Serial,
		}
		// Base item properties
		i.opl.AppendColor(colornames.White, i.DisplayName(), false)
		if i.HasFlags(ItemFlagsContainer) {
			i.opl.AppendColor(colornames.Gray, fmt.Sprintf(
				"%d/%d items, %d/%d stones",
				i.ItemCount, i.MaxItems,
				int(i.ContainedWeight), int(i.MaxWeight)),
				false)
		}
		i.opl.Compile()
		i.oplInfo = &serverpacket.OPLInfo{
			Serial: i.Serial,
			Hash:   i.opl.Hash,
		}
	}
	return i.opl, i.oplInfo
}

// RootContainer returns the top-most item containing this item. If this item
// has no container, which is the case for items directly on the map or
// equipped to a mobile, nil is returned.
func (i *Item) RootContainer() *Item {
	if i.Container == nil {
		return nil
	}
	p := i
	for {
		if p.Container == nil {
			return p
		}
		p = p.Container
	}
}

// UpdateItem updates the item for all observers.
func (i *Item) UpdateItem(item *Item) {
	for _, o := range i.Observers {
		o.ContainerItemAdded(i, item)
	}
}

// UpdateItemOPL updates the OPL information for the given item for every
// observer currently observing this container.
func (i *Item) UpdateItemOPL(item *Item) {
	for _, o := range i.Observers {
		o.ContainerItemOPLChanged(i, item)
	}
}

// ContextMenu returns a new context menu packet.
func (i *Item) ContextMenu(p *ContextMenu, m *Mobile) {
	if i.HasFlags(ItemFlagsContainer) {
		p.Append("OpenContainer", 3000362) // Open
	}
}

func (i *Item) StandingHeight() int {
	if !i.Surface() && !i.Wet() && !i.Impassable() {
		return 0
	}
	if i.Bridge() {
		return i.def.Height / 2
	}
	return i.def.Height
}
func (i *Item) Height() int             { return i.def.Height }
func (i *Item) Highest() int            { return i.Location.Z + i.def.Height }
func (i *Item) Z() int                  { return i.Location.Z }
func (i *Item) Background() bool        { return i.def.TileFlags&uo.TileFlagsBackground != 0 }
func (i *Item) Weapon() bool            { return i.def.TileFlags&uo.TileFlagsWeapon != 0 }
func (i *Item) Transparent() bool       { return i.def.TileFlags&uo.TileFlagsTransparent != 0 }
func (i *Item) Translucent() bool       { return i.def.TileFlags&uo.TileFlagsTranslucent != 0 }
func (i *Item) Wall() bool              { return i.def.TileFlags&uo.TileFlagsWall != 0 }
func (i *Item) Damaging() bool          { return i.def.TileFlags&uo.TileFlagsDamaging != 0 }
func (i *Item) Impassable() bool        { return i.def.TileFlags&uo.TileFlagsImpassable != 0 }
func (i *Item) Wet() bool               { return i.def.TileFlags&uo.TileFlagsWet != 0 }
func (i *Item) Surface() bool           { return i.def.TileFlags&uo.TileFlagsSurface != 0 }
func (i *Item) Bridge() bool            { return i.def.TileFlags&uo.TileFlagsBridge != 0 }
func (i *Item) Generic() bool           { return i.def.TileFlags&uo.TileFlagsGeneric != 0 }
func (i *Item) Window() bool            { return i.def.TileFlags&uo.TileFlagsWindow != 0 }
func (i *Item) NoShoot() bool           { return i.def.TileFlags&uo.TileFlagsNoShoot != 0 }
func (i *Item) ArticleA() bool          { return i.def.TileFlags&uo.TileFlagsArticleA != 0 }
func (i *Item) ArticleAn() bool         { return i.def.TileFlags&uo.TileFlagsArticleAn != 0 }
func (i *Item) Internal() bool          { return i.def.TileFlags&uo.TileFlagsInternal != 0 }
func (i *Item) Foliage() bool           { return i.def.TileFlags&uo.TileFlagsFoliage != 0 }
func (i *Item) PartialHue() bool        { return i.def.TileFlags&uo.TileFlagsPartialHue != 0 }
func (i *Item) NoHouse() bool           { return i.def.TileFlags&uo.TileFlagsNoHouse != 0 }
func (i *Item) Map() bool               { return i.def.TileFlags&uo.TileFlagsMap != 0 }
func (i *Item) StaticContainer() bool   { return i.def.TileFlags&uo.TileFlagsContainer != 0 }
func (i *Item) Wearable() bool          { return i.def.TileFlags&uo.TileFlagsWearable != 0 }
func (i *Item) LightSource() bool       { return i.def.TileFlags&uo.TileFlagsLightSource != 0 }
func (i *Item) Animation() bool         { return i.def.TileFlags&uo.TileFlagsAnimation != 0 }
func (i *Item) NoDiagonal() bool        { return i.def.TileFlags&uo.TileFlagsNoDiagonal != 0 }
func (i *Item) Armor() bool             { return i.def.TileFlags&uo.TileFlagsArmor != 0 }
func (i *Item) Roof() bool              { return i.def.TileFlags&uo.TileFlagsRoof != 0 }
func (i *Item) Door() bool              { return i.def.TileFlags&uo.TileFlagsDoor != 0 }
func (i *Item) StairBack() bool         { return i.def.TileFlags&uo.TileFlagsStairBack != 0 }
func (i *Item) StairRight() bool        { return i.def.TileFlags&uo.TileFlagsStairRight != 0 }
func (i *Item) AlphaBlend() bool        { return i.def.TileFlags&uo.TileFlagsAlphaBlend != 0 }
func (i *Item) UseNewArt() bool         { return i.def.TileFlags&uo.TileFlagsUseNewArt != 0 }
func (i *Item) ArtUsed() bool           { return i.def.TileFlags&uo.TileFlagsArtUsed != 0 }
func (i *Item) NoShadow() bool          { return i.def.TileFlags&uo.TileFlagsBackground != 0 }
func (i *Item) PixelBleed() bool        { return i.def.TileFlags&uo.TileFlagsPixelBleed != 0 }
func (i *Item) PlayAnimOnce() bool      { return i.def.TileFlags&uo.TileFlagsPlayAnimOnce != 0 }
func (i *Item) MultiMovable() bool      { return i.def.TileFlags&uo.TileFlagsMultiMovable != 0 }
func (i *Item) BaseGraphic() uo.Graphic { return i.Graphic }
