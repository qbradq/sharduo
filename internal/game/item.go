package game

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
	"golang.org/x/image/colornames"
)

// LoadItemPrototypes loads all item prototypes.
func LoadItemPrototypes() {
	// Load all item templates
	errors := false
	templates := map[string]*Template{}
	for _, err := range data.Walk("templates/items", func(s string, b []byte) []error {
		// Load prototypes
		ps := map[string]*Template{}
		if err := json.Unmarshal(b, &ps); err != nil {
			return []error{err}
		}
		// Prototype prep
		for k, p := range ps {
			// Check for duplicates
			if _, duplicate := templates[k]; duplicate {
				return []error{fmt.Errorf("duplicate item prototype %s", k)}
			}
			// Initialize non-zero default values
			p.Name = k
			p.Fields["TemplateName"] = k
			templates[k] = p
		}
		return nil
	}) {
		errors = true
		log.Printf("error: during item prototype load: %v", err)
	}
	if errors {
		panic("errors during item prototype load")
	}
	// Resolve all base templates and construct their item prototypes
	for tn, t := range templates {
		t.Resolve(templates)
		i := &Item{
			Object: Object{
				Events: map[string]string{},
			},
		}
		t.constructPrototype(i)
		// Initialize variables
		i.Def = World.ItemDefinition(i.CurrentGraphic())
		itemPrototypes[tn] = i
	}
}

// constructItem creates a new item from the named template.
func constructItem(which string) *Item {
	p := itemPrototypes[which]
	if p == nil {
		return nil
	}
	i := &Item{}
	*i = *p
	i.Def = World.ItemDefinition(i.CurrentGraphic())
	i.Durability = i.MaxDurability
	for _, e := range i.PostCreationEvents {
		if !e.Execute(i) {
			panic(fmt.Errorf("failed to execute post creation event %s", e.EventName))
		}
	}
	return i
}

// NewItem creates a new item and adds it to the world datastores.
func NewItem(which string) *Item {
	i := constructItem(which)
	if i == nil {
		return nil
	}
	World.Add(i)
	return i
}

// itemPrototypes contains all of the prototype items.
var itemPrototypes = map[string]*Item{}

// ItemFlags encode boolean information about an item.
type ItemFlags uint8

const (
	ItemFlagsContainer ItemFlags = 0b00000001 // Item is a container
	ItemFlagsFixed     ItemFlags = 0b00000010 // Item is fixed within the world and cannot be moved normally
	ItemFlagsStatic    ItemFlags = 0b00000100 // Item is static
	ItemFlagsStackable ItemFlags = 0b00001000 // Item is stackable
	ItemFlagsUses      ItemFlags = 0b00010000 // Item has uses instead of durability
	ItemFlagsDyeable   ItemFlags = 0b00100000 // Item can be dyed normally
)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *ItemFlags) AddByName(s string) {
	switch strings.ToLower(s) {
	case "container":
		*f |= ItemFlagsContainer
	case "fixed":
		*f |= ItemFlagsFixed
	case "static":
		*f |= ItemFlagsStatic
	case "stackable":
		*f |= ItemFlagsStackable
	case "uses":
		*f |= ItemFlagsUses
	case "dyeable":
		*f |= ItemFlagsDyeable
	default:
		panic(fmt.Errorf("invalid item flag name %s", s))
	}
}

// Item describes any item in the world.
type Item struct {
	Object
	// Static variables
	Def                *uo.StaticDefinition // Item properties pointer
	Flags              ItemFlags            // Boolean item flags
	Graphic            uo.Graphic           // Base graphic to use for the item
	FlippedGraphic     uo.Graphic           // Flipped graphic
	Layer              uo.Layer             // Layer the object is worn on
	Weight             float64              // Weight of the item, NOT the stack, just one of these items
	MaxContainerWeight float64              // Maximum weight that can be held in this container
	MaxContainerItems  int                  // Maximum number of items that can be held in this container
	Gump               uo.GUMP              // GUMP graphic to use for containers
	Bounds             uo.Bounds            // Container GUMP bounds
	LiftSound          uo.Sound             // Sound this item makes when lifted
	DropSound          uo.Sound             // Sound this container makes by default when dropping an item into it
	Value              int                  // Purchase price of the item if it can be bought
	DropSoundOverride  uo.Sound             // Sound to override the normal drop sound for this item
	Plural             string               // String to use for stacks greater than 1
	UseSkill           uo.Skill             // Skill the item checks on use / attack / parry etc
	AnimationAction    uo.AnimationAction   // Animation action this item plays
	// Persistent variables
	Flipped       bool        // If true the item is currently flipped
	Amount        int         // Stack amount
	Contents      []*Item     // Contents of the container
	LootType      uo.LootType // Persistent loot type so we can bless arbitrary items
	MaxDurability int         // Maximum durability, 0 means this item does not use durability
	Durability    int         // Current durability
	IArg          int         // Generic int argument
	MArg          *Mobile     // Generic mobile argument
	// Transient values
	Wearer          *Mobile                        // Pointer to the mobile currently wearing this item if any, note this only indicates if the item is directly equipped to the mobile, not within equipped containers
	Container       *Item                          // Pointer to the parent container if any
	Observers       map[ContainerObserver]struct{} // All observers currently observing this container
	ItemCount       int                            // Cache of the total number of contained items including all sub-containers
	Gold            int                            // Cache of the total amount of gold coins contained in this and all sub containers
	ContainedWeight float64                        // Cache of the total weight held in this and all sub containers
	decayAt         uo.Time                        // When this item should decay on the map
}

// Write writes the persistent data of the item to w.
func (i *Item) Write(w io.Writer) {
	util.PutUInt32(w, 0)                       // Version
	util.PutString(w, i.TemplateName)          // Template name
	util.PutUInt32(w, uint32(i.Serial))        // Serial
	util.PutPoint(w, i.Location)               // Location
	util.PutByte(w, byte(i.Facing))            // Facing
	util.PutUInt16(w, uint16(i.Hue))           // Hue
	util.PutBool(w, i.Flipped)                 // Flipped flag
	util.PutUInt16(w, uint16(i.Amount))        // Stack amount
	util.PutByte(w, byte(i.LootType))          // Loot type
	util.PutUInt16(w, uint16(i.MaxDurability)) // Maximum durability
	util.PutUInt16(w, uint16(i.Durability))    // Current durability
	util.PutUInt32(w, uint32(i.IArg))          // Generic int argument
	if i.MArg != nil {                         // Generic mobile argument
		util.PutBool(w, true)
		i.MArg.Write(w)
	} else {
		util.PutBool(w, false)
	}
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
	i.Flipped = util.GetBool(r)               // Flipped flag
	i.Amount = int(util.GetUInt16(r))         // Stack amount
	i.LootType = uo.LootType(util.GetByte(r)) // Loot type
	i.MaxDurability = int(util.GetUInt16(r))  // Maximum durability
	i.Durability = int(util.GetUInt16(r))     // Current durability
	i.IArg = int(util.GetUInt32(r))           // Generic int argument
	if util.GetBool(r) {                      // Generic mobile argument
		i.MArg = NewMobileFromReader(r)
	}
	n := int(util.GetUInt32(r))   // Contents item count
	i.Contents = make([]*Item, n) // Contents
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
	for co := range i.Observers {
		if co == o {
			return
		}
	}
	i.Observers[o] = struct{}{}
}

// RemoveObserver removes the ContainerObserver from the list of current
// observers.
func (i *Item) RemoveObserver(o ContainerObserver) {
	delete(i.Observers, o)
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
				i.ItemCount, i.MaxContainerItems,
				int(i.ContainedWeight), int(i.MaxContainerWeight)),
				false)
		}
		if i.MaxDurability > 0 {
			if i.HasFlags(ItemFlagsUses) {
				i.opl.AppendColor(colornames.Aqua, fmt.Sprintf(
					"Uses: %d/%d",
					i.Durability, i.MaxDurability,
				), false)
			} else {
				i.opl.AppendColor(colornames.Gray, fmt.Sprintf(
					"Durability: %d/%d",
					i.Durability, i.MaxDurability,
				), false)
			}
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
	for o := range i.Observers {
		o.ContainerItemAdded(i, item)
	}
}

// UpdateItemOPL updates the OPL information for the given item for every
// observer currently observing this container.
func (i *Item) UpdateItemOPL(item *Item) {
	for o := range i.Observers {
		o.ContainerItemOPLChanged(i, item)
	}
}

// ContextMenuPacket returns a new context menu packet.
func (i *Item) ContextMenuPacket(p *ContextMenu, m *Mobile) {
	if i.HasFlags(ItemFlagsContainer) {
		p.Append("OpenContainer", 3000362) // Open
	}
	for _, e := range i.ContextMenu {
		p.Append(e.Event, e.Cliloc)
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
	aw := item.Weight*float64(item.Amount) + item.ContainedWeight
	// Check item and weight limits
	if !force {
		if i.ItemCount+ai > i.MaxContainerItems {
			return &UOError{
				Cliloc: 1080017, // That container cannot hold more items.
			}
		}
		if i.ContainedWeight+aw > i.MaxContainerWeight {
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
				i.AdjustWeightAndCount(aw, 0)
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
	item.Container = i
	i.AdjustWeightAndCount(aw, ai)
	if item.TemplateName == "GoldCoin" {
		i.AdjustGold(item.Amount)
	}
	// Let all observers know about the new item
	for o := range i.Observers {
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
	item.Container = nil
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

// Remove cleanly removes this item from its parent and the world datastores.
func (i *Item) Remove() {
	if i.Container != nil {
		i.Container.RemoveItem(i)
	} else {
		World.Map().RemoveItem(i)
	}
	World.RemoveItem(i)
}

// Consume removes n from the amount and removes the item if none are left.
// Returns true on success.
func (i *Item) Consume(n int) bool {
	if n > i.Amount {
		return false
	}
	i.Amount -= n
	if i.Amount == 0 {
		i.Remove()
	}
	return true
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

// CurrentGraphic returns the current graphic in use by the item.
func (i *Item) CurrentGraphic() uo.Graphic {
	if i.Flipped {
		return i.FlippedGraphic
	}
	return i.Graphic
}

// Open implements the object interface.
func (i *Item) Open(m *Mobile) {
	if m.NetState == nil {
		return
	}
	observer, ok := m.NetState.(ContainerObserver)
	if !ok {
		return
	}
	// TODO access calculations
	if i.Observers == nil {
		i.Observers = make(map[ContainerObserver]struct{})
	}
	i.Observers[observer] = struct{}{}
	observer.ContainerOpen(i)
}

// SetBaseGraphicForBody sets the base graphic of the item correctly for the
// given mount body.
func (i *Item) SetBaseGraphicForBody(body uo.Body) {
	switch body {
	case 0xC8:
		i.Graphic = 0x3E9F
	case 0xCC:
		i.Graphic = 0x3EA2
	case 0xDC:
		i.Graphic = 0x3EA6
	case 0xE2:
		i.Graphic = 0x3EA0
	case 0xE4:
		i.Graphic = 0x3EA1
	}
}

// ConsumeUse consumes a use off of this item returning true if successful.
func (i *Item) ConsumeUse() bool {
	if i.Durability < 1 {
		return false
	}
	i.Durability--
	return true
}

// Update implements the Object interface.
func (i *Item) Update(t uo.Time) {
	if t >= i.decayAt {
		if i.Spawner != nil {
			// If we are being managed by a spawner we don't decay
			return
		}
		i.Remove()
	}
}

// DropSoundOverride returns the drop sound override or s if there is none.
func (i *Item) GetDropSoundOverride(s uo.Sound) uo.Sound {
	if i.DropSoundOverride != uo.SoundInvalidDrop {
		return i.DropSoundOverride
	}
	return s
}

// DisplayName returns the normalized displayable name of the object.
func (i *Item) DisplayName() string {
	if i.Amount > 1 {
		return i.Plural
	} else {
		if i.Object.ArticleA {
			return "a " + i.Name
		}
		if i.Object.ArticleAn {
			return "an " + i.Name
		}
	}
	return i.Name
}

// ExecuteEvent executes the named event handler if any is configured. Returns
// true if the handler was found and also returned true.
func (i *Item) ExecuteEvent(which string, s, v any) bool {
	hn, ok := i.Events[which]
	if !ok {
		return false
	}
	return ExecuteEventHandler(hn, i, s, v)
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
