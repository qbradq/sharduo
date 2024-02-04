package game

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"reflect"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
	"golang.org/x/image/colornames"
)

func init() {
	// Load all item templates
	errors := false
	for _, err := range data.Walk("templates/items", func(s string, b []byte) []error {
		// Ignore legacy files
		if filepath.Ext(s) != ".json" {
			return nil
		}
		// Load prototypes
		ps := map[string]*Item{}
		if err := json.Unmarshal(b, &ps); err != nil {
			return []error{err}
		}
		// Prototype prep
		for k, p := range ps {
			// Check for duplicates
			if _, duplicate := itemPrototypes[k]; duplicate {
				return []error{fmt.Errorf("duplicate item prototype %s", k)}
			}
			// Initialize non-zero default values
			p.Def = World.ItemDefinition(p.Graphic)
			p.TemplateName = k
			itemPrototypes[k] = p
		}
		return nil
	}) {
		errors = true
		log.Printf("error: during item prototype load: %v", err)
	}
	if errors {
		panic("errors during item prototype load")
	}
	// Resolve all base templates
	var fn func(*Item)
	fn = func(i *Item) {
		// Skip resolved templates and root templates
		if i.btResolved || i.BaseTemplate == "" {
			return
		}
		// Resolve base template
		p := itemPrototypes[i.BaseTemplate]
		if p == nil {
			panic(fmt.Errorf("item template %s referenced non-existent base template %s",
				i.TemplateName, i.BaseTemplate))
		}
		fn(p)
		// Merge values
		pr := reflect.ValueOf(p)
		ir := reflect.ValueOf(i)
		for i := 0; i < pr.NumField(); i++ {
			sf := pr.Type().Field(i)
			switch sf.Name {
			case "PostCreationEvents":
				// Prepend array contents
				prf := pr.Field(i)
				irf := ir.Field(i)
				irf.Set(reflect.AppendSlice(prf, irf))
			case "Events":
				// Merge map
				prf := pr.Field(i)
				irf := ir.Field(i)
				for _, k := range prf.MapKeys() {
					if irf.MapIndex(k).IsZero() {
						irf.MapIndex(k).Set(prf.MapIndex(k))
					}
				}
			case "Flags":
				// Merge flag bits
				ir.Field(i).Set(reflect.ValueOf(ItemFlags(pr.Field(i).Int() | ir.Field(i).Int())))
			default:
				// Just copy the value
				ir.Field(i).Set(pr.Field(i))
			}
		}
		// Flag prototype as done
		i.btResolved = true
	}
}

// constructItem creates a new item from the named template.
func constructItem(which string) *Item {
	p := itemPrototypes[which]
	if p == nil {
		panic(fmt.Errorf("unknown item prototype %s", which))
	}
	i := &Item{}
	*i = *p
	for _, en := range i.PostCreationEvents {
		i.ExecuteEvent(en, nil, nil)
	}
	return i
}

// NewItem creates a new item and adds it to the world datastores.
func NewItem(which string) *Item {
	i := constructItem(which)
	World.Add(i)
	return i
}

// itemPrototypes contains all of the prototype items.
var itemPrototypes map[string]*Item

// ItemFlags encode boolean information about an item.
type ItemFlags uint8

const (
	ItemFlagsContainer ItemFlags = 0b00000001 // Item is a container
	ItemFlagsFixed     ItemFlags = 0b00000010 // Item is fixed within the world and cannot be moved normally
	ItemFlagsStatic    ItemFlags = 0b00000100 // Item is static
	ItemFlagsStackable ItemFlags = 0b00001000 // Item is stackable
)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *ItemFlags) UnmarshalJSON(in []byte) error {
	flags := []string{}
	if err := json.Unmarshal(in, &flags); err != nil {
		return err
	}
	for _, s := range flags {
		switch s {
		case "container":
			*f |= ItemFlagsContainer
		case "fixed":
			*f |= ItemFlagsFixed
		case "static":
			*f |= ItemFlagsStatic
		case "stackable":
			*f |= ItemFlagsStackable
		}
	}
	return nil
}

// Item describes any item in the world.
type Item struct {
	Object
	// Static variables
	Def         *uo.StaticDefinition // Item properties pointer
	Flags       ItemFlags            // Boolean item flags
	Graphic     uo.Graphic           // Base graphic to use for the item
	Layer       uo.Layer             // Layer the object is worn on
	Weight      float64              // Weight of the item, NOT the stack, just one of these items
	MaxWeight   float64              // Maximum weight that can be held in this container
	MaxItems    int                  // Maximum number of items that can be held in this container
	GUMPGraphic uo.GUMP              // GUMP graphic to use for containers
	Bounds      uo.Bounds            // Container GUMP bounds
	LiftSound   uo.Sound             // Sound this item makes when lifted
	// Persistent variables
	Amount   int         // Stack amount
	Contents []*Item     // Contents of the container
	LootType uo.LootType // Persistent loot type so we can bless arbitrary items
	IArg     int         // Generic int argument
	// Transient values
	Wearer          *Mobile             // Pointer to the mobile currently wearing this item if any, note this only indicates if the item is directly equipped to the mobile, not within equipped containers
	Container       *Item               // Pointer to the parent container if any
	Observers       []ContainerObserver // All observers currently observing this container
	ItemCount       int                 // Cache of the total number of contained items including all sub-containers
	Gold            int                 // Cache of the total amount of gold coins contained in this and all sub containers
	ContainedWeight float64             // Cache of the total weight held in this and all sub containers
	decayAt         uo.Time             // When this item should decay on the map
}

// Write writes the persistent data of the item to w.
func (i *Item) Write(w io.Writer) {
	util.PutUInt32(w, 0)                       // Version
	util.PutString(w, i.TemplateName)          // Template name
	util.PutUInt32(w, uint32(i.Serial))        // Serial
	util.PutPoint(w, i.Location)               // Location
	util.PutByte(w, byte(i.Facing))            // Facing
	util.PutUInt16(w, uint16(i.Hue))           // Hue
	util.PutUInt16(w, uint16(i.Amount))        // Stack amount
	util.PutByte(w, byte(i.LootType))          // Loot type
	util.PutUInt32(w, uint32(i.IArg))          // Generic int argument
	util.PutUInt32(w, uint32(len(i.Contents))) // Contents
	for _, item := range i.Contents {
		item.Write(w)
	}
}

// NewItemFromReader reads the persistent data of the item from r and returns a
// new item. This also inserts the item into the world datastore.
func NewItemFromReader(r io.Reader) *Item {
	_ = util.GetUInt32(r)   // Version
	tn := util.GetString(r) // Template name
	i := constructItem(tn)
	i.TemplateName = tn
	i.Serial = uo.Serial(util.GetUInt32(r))   // Serial
	i.Location = util.GetPoint(r)             // Location
	i.Facing = uo.Direction(util.GetByte(r))  // Facing
	i.Hue = uo.Hue(util.GetUInt16(r))         // Hue
	i.Amount = int(util.GetUInt16(r))         // Stack amount
	i.LootType = uo.LootType(util.GetByte(r)) // Loot type
	i.IArg = int(util.GetUInt32(r))           // Generic int argument
	n := int(util.GetUInt32(r))               // Contents item count
	i.Contents = make([]*Item, n)             // Contents
	for idx := 0; idx < n; idx++ {
		i.Contents[idx] = NewItemFromReader(r)
	}
	World.Insert(i)
	return i
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

// Split splits off a number of items from a stack. nil is returned if
// n < 1 || n >= item.Amount(). nil is also returned for all non-stackable
// items. In the event of an error during duplication the error will be
// logged and nil will be returned. Otherwise a new duplicate item is
// created with the remaining amount. This item is removed from its parent.
// If this remove operation fails this function returns nil. The new
// object is then force-added to the old parent in the same location.
// The parent of this item is then set to the new item. If nil is returned
// this item's amount and parent has not changed.
func (i *Item) Split(n int) *Item {
	// No new item required in these cases
	if !i.HasFlags(ItemFlagsStackable) || n < 1 || n >= i.Amount {
		return nil
	}
	// Create the new item
	item := NewItem(i.TemplateName)
	// Remove this item from its parent
	if i.Container == nil {
		World.Map().RemoveItem(i)
	} else {
		i.Container.RemoveItem(i)
	}
	item.Amount = i.Amount - n
	i.Amount = n
	// Force the remainder back where we came from
	item.Location = i.Location
	if i.Container == nil {
		World.Map().AddItem(item, false)
	} else {
		i.Container.AddItem(item, false)
	}
	i.Container = item
	i.Container.AdjustWeightAndCount(i.Weight*float64(-n), -n)
	return item
}

// RefreshDecayDeadline refreshes the decay deadline for the item.
func (i *Item) RefreshDecayDeadline() {
	if i.Spawner != nil {
		// If we are being managed by a spawner we don't decay
		i.decayAt = uo.TimeNever
		return
	}
	switch i.LootType {
	case uo.LootTypeNormal:
		i.decayAt = World.Time() + uo.DurationHour
	case uo.LootTypeBlessed:
		i.decayAt = World.Time() + uo.DurationHour
	case uo.LootTypeNewbie:
		i.decayAt = World.Time() + uo.DurationSecond*15
	case uo.LootTypeSystem:
		i.decayAt = uo.TimeNever
	}
}

// AddItem adds an item to this item's inventory.
func (i *Item) AddItem(item *Item, force bool) error {
	ai := 1 + item.ItemCount
	aw := item.Weight + item.ContainedWeight
	// Check item and weight limits
	if !force {
		if i.ItemCount+ai > i.MaxItems {
			return &UOError{
				Cliloc: 1080017, // That container cannot hold more items.
			}
		}
		if i.ContainedWeight+aw > i.MaxWeight {
			return &UOError{
				Cliloc: 1080016, // That container cannot hold more weight.
			}
		}
	}
	// Try to stack
	if item.HasFlags(ItemFlagsStackable) &&
		item.Location == uo.RandomContainerLocation {
		for _, oi := range i.Contents {
			// Same item type and enough stack capacity to accept
			if item.TemplateName == oi.TemplateName &&
				item.Hue == oi.Hue &&
				oi.Amount+item.Amount <= uo.MaxStackAmount {
				i.AdjustWeightAndCount(item.Weight, 0)
				oi.Amount = oi.Amount + item.Amount
				if item.TemplateName == "GoldCoin" {
					i.AdjustGold(item.Amount)
				}
				World.RemoveItem(item)
				World.UpdateItem(oi)
				return nil
			}
		}
	}
	// Handle drop location
	l := item.Location
	if l.X == uo.RandomDropX {
		l.X = util.Random(i.Bounds.X, i.Bounds.X+i.Bounds.W-1)
	}
	if l.Y == uo.RandomDropY {
		l.Y = util.Random(i.Bounds.Y, i.Bounds.Y+i.Bounds.H-1)
	}
	if l.X < i.Bounds.X {
		l.X = i.Bounds.X
	}
	if l.X >= i.Bounds.X+i.Bounds.W {
		l.X = i.Bounds.X + i.Bounds.W - 1
	}
	if l.Y < i.Bounds.Y {
		l.Y = i.Bounds.Y
	}
	if l.Y >= i.Bounds.Y+i.Bounds.H {
		l.Y = i.Bounds.Y + i.Bounds.H - 1
	}
	// Add item to container
	i.Contents = append(i.Contents, item)
	i.AdjustWeightAndCount(aw, ai)
	if item.TemplateName == "GoldCoin" {
		i.AdjustGold(item.Amount)
	}
	// Let all observers know about the new item
	for _, o := range i.Observers {
		o.ContainerItemAdded(i, item)
	}
	return nil
}

// RemoveItem removes an item from this item's inventory.
func (i *Item) RemoveItem(item *Item) {
	idx := -1
	for i, oi := range i.Contents {
		if oi == item {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	copy(i.Contents[idx:], i.Contents[idx+1:])
	i.Contents[len(i.Contents)-1] = nil
	i.Contents = i.Contents[:len(i.Contents)-1]
	i.AdjustWeightAndCount(-item.Weight+item.ContainedWeight, -item.ItemCount+1)
}

// AdjustWeightAndCount adjusts the contained weight and item count of the
// container.
func (i *Item) AdjustWeightAndCount(w float64, n int) {
	i.InvalidateOPL()
	i.ContainedWeight += w
	i.ItemCount += n
	if i.TemplateName == "PlayerBankBox" {
		// We are a mobile's bank box, don't propagate up
		return
	}
	if i.Container != nil {
		// We are a sub-container, propagate the adjustment up
		i.Container.AdjustWeightAndCount(w, n)
	} else if i.Wearer != nil {
		i.Wearer.InvalidateOPL()
		if i.Wearer.Cursor == i {
			// We are being held by a mobile's cursor, don't need to do anything
			return
		}
		// We are a mobile's backpack, send the weight adjustment up
		i.Wearer.AdjustWeight(w)
	}
}

// AdjustGold adjusts the cached value of how much gold the container holds
// including all sub-containers. This function propagates upward to parent
// containers.
func (i *Item) AdjustGold(n int) {
	i.Gold += n
	if i.Container != nil {
		i.Container.AdjustGold(n)
	}
}

// InvalidateOPL schedules an OPL update.
func (i *Item) InvalidateOPL() {
	i.opl = nil
	i.oplInfo = nil
	World.UpdateItemOPLInfo(i)
}

// Consume removes items of the given template name from the container and all
// sub-containers until the amount specified has been consumed. If there are not
// enough items to satisfy the amount the function returns false and no items
// are modified or removed.
func (i *Item) Consume(tn string, n int) bool {
	// TODO Stub
	return false
}

// ConsumeGold removes gold from the container and all sub-containers until the
// amount specified has been consumed. If there are not enough gold in coin
// piles and checks to satisfy the amount the function returns false and no
// items are modified or removed.
func (i *Item) ConsumeGold(n int) bool {
	var remove []*Item
	ac := 0
	var getGold func(*Item) bool
	getGold = func(c *Item) bool {
		for _, item := range c.Contents {
			if item.HasFlags(ItemFlagsContainer) {
				if getGold(item) {
					return true
				}
				continue
			}
			if item.TemplateName == "GoldCoin" {
				if ac+item.Amount >= n {
					item.Amount -= (n - ac)
					if item.Amount < 1 {
						remove = append(remove, item)
					} else {
						item.InvalidateOPL()
					}
					for _, tr := range remove {
						tr.Container.RemoveItem(tr)
						World.RemoveItem(tr)
					}
					return true
				}
				ac += item.Amount
				remove = append(remove, item)
			}
		}
		return false
	}
	var getCheck func(*Item) bool
	getCheck = func(c *Item) bool {
		for _, item := range c.Contents {
			if item.HasFlags(ItemFlagsContainer) {
				if getCheck(item) {
					return true
				}
				continue
			}
			if item.TemplateName == "Check" {
				if ac+item.IArg >= n {
					item.IArg -= (n - ac)
					if item.IArg < 1 {
						remove = append(remove, item)
					} else {
						item.InvalidateOPL()
					}
					for _, tr := range remove {
						tr.Container.RemoveItem(tr)
						World.RemoveItem(tr)
					}
					return true
				}
				ac += item.IArg
				remove = append(remove, item)
			}
		}
		return false
	}
	return getGold(i) || getCheck(i)
}

// DropInto attempts to add the item to this container at a random location.
func (i *Item) DropInto(item *Item, force bool) error {
	item.Location = uo.RandomContainerLocation
	return i.AddItem(item, force)
}

func (i *Item) StandingHeight() int {
	if !i.Surface() && !i.Wet() && !i.Impassable() {
		return 0
	}
	if i.Bridge() {
		return i.Def.Height / 2
	}
	return i.Def.Height
}
func (i *Item) Height() int             { return i.Def.Height }
func (i *Item) Highest() int            { return i.Location.Z + i.Def.Height }
func (i *Item) Z() int                  { return i.Location.Z }
func (i *Item) Background() bool        { return i.Def.TileFlags&uo.TileFlagsBackground != 0 }
func (i *Item) Weapon() bool            { return i.Def.TileFlags&uo.TileFlagsWeapon != 0 }
func (i *Item) Transparent() bool       { return i.Def.TileFlags&uo.TileFlagsTransparent != 0 }
func (i *Item) Translucent() bool       { return i.Def.TileFlags&uo.TileFlagsTranslucent != 0 }
func (i *Item) Wall() bool              { return i.Def.TileFlags&uo.TileFlagsWall != 0 }
func (i *Item) Damaging() bool          { return i.Def.TileFlags&uo.TileFlagsDamaging != 0 }
func (i *Item) Impassable() bool        { return i.Def.TileFlags&uo.TileFlagsImpassable != 0 }
func (i *Item) Wet() bool               { return i.Def.TileFlags&uo.TileFlagsWet != 0 }
func (i *Item) Surface() bool           { return i.Def.TileFlags&uo.TileFlagsSurface != 0 }
func (i *Item) Bridge() bool            { return i.Def.TileFlags&uo.TileFlagsBridge != 0 }
func (i *Item) Generic() bool           { return i.Def.TileFlags&uo.TileFlagsGeneric != 0 }
func (i *Item) Window() bool            { return i.Def.TileFlags&uo.TileFlagsWindow != 0 }
func (i *Item) NoShoot() bool           { return i.Def.TileFlags&uo.TileFlagsNoShoot != 0 }
func (i *Item) ArticleA() bool          { return i.Def.TileFlags&uo.TileFlagsArticleA != 0 }
func (i *Item) ArticleAn() bool         { return i.Def.TileFlags&uo.TileFlagsArticleAn != 0 }
func (i *Item) Internal() bool          { return i.Def.TileFlags&uo.TileFlagsInternal != 0 }
func (i *Item) Foliage() bool           { return i.Def.TileFlags&uo.TileFlagsFoliage != 0 }
func (i *Item) PartialHue() bool        { return i.Def.TileFlags&uo.TileFlagsPartialHue != 0 }
func (i *Item) NoHouse() bool           { return i.Def.TileFlags&uo.TileFlagsNoHouse != 0 }
func (i *Item) Map() bool               { return i.Def.TileFlags&uo.TileFlagsMap != 0 }
func (i *Item) StaticContainer() bool   { return i.Def.TileFlags&uo.TileFlagsContainer != 0 }
func (i *Item) Wearable() bool          { return i.Def.TileFlags&uo.TileFlagsWearable != 0 }
func (i *Item) LightSource() bool       { return i.Def.TileFlags&uo.TileFlagsLightSource != 0 }
func (i *Item) Animation() bool         { return i.Def.TileFlags&uo.TileFlagsAnimation != 0 }
func (i *Item) NoDiagonal() bool        { return i.Def.TileFlags&uo.TileFlagsNoDiagonal != 0 }
func (i *Item) Armor() bool             { return i.Def.TileFlags&uo.TileFlagsArmor != 0 }
func (i *Item) Roof() bool              { return i.Def.TileFlags&uo.TileFlagsRoof != 0 }
func (i *Item) Door() bool              { return i.Def.TileFlags&uo.TileFlagsDoor != 0 }
func (i *Item) StairBack() bool         { return i.Def.TileFlags&uo.TileFlagsStairBack != 0 }
func (i *Item) StairRight() bool        { return i.Def.TileFlags&uo.TileFlagsStairRight != 0 }
func (i *Item) AlphaBlend() bool        { return i.Def.TileFlags&uo.TileFlagsAlphaBlend != 0 }
func (i *Item) UseNewArt() bool         { return i.Def.TileFlags&uo.TileFlagsUseNewArt != 0 }
func (i *Item) ArtUsed() bool           { return i.Def.TileFlags&uo.TileFlagsArtUsed != 0 }
func (i *Item) NoShadow() bool          { return i.Def.TileFlags&uo.TileFlagsBackground != 0 }
func (i *Item) PixelBleed() bool        { return i.Def.TileFlags&uo.TileFlagsPixelBleed != 0 }
func (i *Item) PlayAnimOnce() bool      { return i.Def.TileFlags&uo.TileFlagsPlayAnimOnce != 0 }
func (i *Item) MultiMovable() bool      { return i.Def.TileFlags&uo.TileFlagsMultiMovable != 0 }
func (i *Item) BaseGraphic() uo.Graphic { return i.Graphic }
