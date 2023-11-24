package game

import (
	"fmt"
	"strconv"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("BaseItem", marshal.ObjectTypeItem, func() any { return &BaseItem{} })
}

// Item is the interface that all non-static items implement.
type Item interface {
	Object
	// BaseGraphic returns the base graphic of the item
	BaseGraphic() uo.Graphic
	// SetBaseGraphic sets the base graphic of the item
	SetBaseGraphic(uo.Graphic)
	// Graphic returns the current effective graphic of the item without the
	// graphic offset
	Graphic() uo.Graphic
	// GraphicOffset returns the current graphic offset of the item, this will
	// range 0-255 inclusive.
	GraphicOffset() int
	// Dyable returns true if the item's hue can be changed by the player
	Dyable() bool
	// Flippable returns true if the item can be flipped / turned
	Flippable() bool
	// Flipped returns true if the item is currently flipped / turned
	Flipped() bool
	// Flip toggles the flipped / turned state
	Flip()
	// FlippedGraphic returns the flipped graphic for this item if any.
	FlippedGraphic() uo.Graphic
	// SetFlippedGraphic sets the flipped graphic for this item.
	SetFlippedGraphic(uo.Graphic)
	// Stackable returns true if the item can be stacked
	Stackable() bool
	// Movable returns true if the item can be moved
	Movable() bool
	// Amount of the stack
	Amount() int
	// SetAmount sets the amount of the stack. If this is out of range it will
	// be bounded to a sane value
	SetAmount(int)
	// Value returns the base sale price of the item at a vendor
	Value() int
	// Consume attempts to remove n from the number of items in this stack and
	// returns true if successful. This function takes care of removing the
	// object if amount reaches zero and updating the object otherwise.
	Consume(int) bool
	// Split splits off a number of items from a stack. nil is returned if
	// n < 1 || n >= item.Amount(). nil is also returned for all non-stackable
	// items. In the event of an error during duplication the error will be
	// logged and nil will be returned. Otherwise a new duplicate item is
	// created with the remaining amount. This item is removed from its parent.
	// If this remove operation fails this function returns nil. The new
	// object is then force-added to the old parent in the same location.
	// The parent of this item is then set to the new item. If nil is returned
	// this item's amount and parent has not changed.
	Split(int) Item
	// Combine adds the amount of this item to the amount of the other item,
	// then replaces itself with that other item. Returns false on failure.
	// Failure can happen if this item does not support stacking, the items are
	// not from the same template object, or the combined amounts would be
	// greater than uo.MaxStackAmount. If this function returns true this item
	// has been totally removed from the world and data stores.
	Combine(Item) bool
	// CanCombineWith returns true if the given item can be combined with this
	// one. This function does not consider max stack count.
	CanCombineWith(Item) bool
	// Height returns the height of the item
	Height() int8
	// Highest returns the highest elevation of the object.
	Highest() int8
	// StandingHeight returns the standing height offset of the item, usually 0
	StandingHeight() int8
	// Z returns the permanent Z location of the object
	Z() int8
	// DropLocation returns the requested drop location of an item
	DropLocation() uo.Location
	// SetDropLocation sets the requested drop location of an item
	SetDropLocation(uo.Location)
	// RefreshDecayDeadline updates the decay deadline for this item based on
	// the item's internal state.
	RefreshDecayDeadline()
	// LiftSound returns the sound to play when this item is lifted.
	LiftSound() uo.Sound
	// DropSoundOverride returns the sound to play considering the item's
	// override and the given sound.
	DropSoundOverride(uo.Sound) uo.Sound
	// Uses returns the number of generic uses / charges left on an item.
	Uses() int
	// ConsumeUse returns true if a generic use / charge was able to be consumed
	// from the item.
	ConsumeUse() bool
	// SetDefForGraphic forces a different static definition to be used for this
	// item instead of the one associated with the base graphic. This is really
	// hacky and was put in for southward counter-clockwise-opening doors.
	SetDefForGraphic(uo.Graphic)

	//
	// Flag accessors
	//

	Background() bool
	Weapon() bool
	Transparent() bool
	Translucent() bool
	Wall() bool
	Damaging() bool
	Impassable() bool
	Wet() bool
	Surface() bool
	Bridge() bool
	Generic() bool
	Window() bool
	NoShoot() bool
	ArticleA() bool
	ArticleAn() bool
	Internal() bool
	Foliage() bool
	PartialHue() bool
	NoHouse() bool
	Map() bool
	Container() bool
	Wearable() bool
	LightSource() bool
	Animation() bool
	NoDiagonal() bool
	Armor() bool
	Roof() bool
	Door() bool
	StairBack() bool
	StairRight() bool
	AlphaBlend() bool
	UseNewArt() bool
	ArtUsed() bool
	NoShadow() bool
	PixelBleed() bool
	PlayAnimOnce() bool
	MultiMovable() bool
}

// BaseItem provides the basic implementation of Item.
type BaseItem struct {
	BaseObject

	//
	// Persistant values
	//

	// Flipped is true if the item is currently flipped.
	flipped bool
	// Stack amount
	amount int
	// Uses left
	uses int
	// Loot type
	lootType uo.LootType

	//
	// Non-persistent values
	//

	// Drop request location
	dropLocation uo.Location
	// Static definition that holds the static data for the item
	def *uo.StaticDefinition
	// Graphic of the item
	graphic uo.Graphic
	// Graphic of the item when flipped. If this is uo.GraphicNone the item cannot
	// be flipped.
	flippedGraphic uo.Graphic
	// Dyable flag
	dyable bool
	// Base weight of the item
	weight float32
	// Stackable flag
	stackable bool
	// Name when there is more than one in the stack
	plural string
	// Base sale value of the item at a vendor
	value int
	// Time at which this item decays on the ground.
	decayDeadline uo.Time
	// Lift sound of the item
	liftSound uo.Sound
	// Override of the container's drop sound if any
	dropSoundOverride uo.Sound
	// No-rent flag
	noRent bool
	// Non-movable flag
	fixed bool
}

// ObjectType implements the Object interface.
func (i *BaseItem) ObjectType() marshal.ObjectType { return marshal.ObjectTypeItem }

// Marshal implements the marshal.Marshaler interface.
func (i *BaseItem) Marshal(s *marshal.TagFileSegment) {
	i.BaseObject.Marshal(s)
	s.PutBool(i.flipped)
	s.PutShort(uint16(i.amount))
	s.PutShort(uint16(i.uses))
	s.PutByte(byte(i.lootType))
}

// Deserialize implements the util.Serializeable interface.
func (i *BaseItem) Deserialize(t *template.Template, create bool) {
	i.BaseObject.Deserialize(t, create)
	i.graphic = uo.Graphic(t.GetNumber("Graphic", int(uo.GraphicDefault)))
	i.def = world.GetItemDefinition(i.graphic)
	i.flippedGraphic = uo.Graphic(t.GetNumber("FlippedGraphic", int(uo.GraphicNone)))
	i.dyable = t.GetBool("Dyable", false)
	i.weight = t.GetFloat("Weight", 255.0)
	i.stackable = t.GetBool("Stackable", false)
	i.amount = t.GetNumber("Amount", 1)
	i.value = t.GetNumber("Value", 0)
	i.plural = t.GetString("Plural", "")
	i.liftSound = uo.Sound(t.GetNumber("LiftSound", int(uo.SoundDefaultLift)))
	i.dropSoundOverride = uo.Sound(t.GetNumber("DropSoundOverride", int(uo.SoundInvalidDrop)))
	i.uses = t.GetNumber("Uses", 0)
	i.lootType = uo.LootType(t.GetNumber("LootType", int(uo.LootTypeNormal)))
	i.noRent = t.GetBool("NoRent", false)
	i.fixed = t.GetBool("Fixed", false)
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *BaseItem) Unmarshal(s *marshal.TagFileSegment) {
	i.BaseObject.Unmarshal(s)
	i.flipped = s.Bool()
	i.amount = int(s.Short())
	if i.amount < 1 {
		i.amount = 1
	}
	i.uses = int(s.Short())
	i.lootType = uo.LootType(s.Byte())
	i.def = world.GetItemDefinition(i.graphic)
	// Instead of storing the decay deadline we just refresh everything on a
	// world load.
	i.RefreshDecayDeadline()
}

// BaseGraphic implements the Item interface.
func (i *BaseItem) BaseGraphic() uo.Graphic { return i.graphic }

// SetBaseGraphic implements the Item interface.
func (i *BaseItem) SetBaseGraphic(g uo.Graphic) {
	i.graphic = g
	i.def = world.GetItemDefinition(g)
	i.InvalidateOPL()
	world.Update(i)
}

// Graphic implements the Item interface.
func (i *BaseItem) Graphic() uo.Graphic {
	if i.flipped && i.flippedGraphic != uo.GraphicNone {
		return i.flippedGraphic
	}
	return i.graphic
}

// GraphicOffset implements the Item interface.
func (i *BaseItem) GraphicOffset() int {
	return 0
}

// Dyable implements the Item interface.
func (i *BaseItem) Dyable() bool { return i.dyable }

// Flippable implements the Item interface.
func (i *BaseItem) Flippable() bool { return i.flippedGraphic != uo.GraphicNone }

// Flipped implements the Item interface.
func (i *BaseItem) Flipped() bool { return i.flipped }

// Flip implements the Item interface.
func (i *BaseItem) Flip() {
	i.flipped = !i.flipped
	world.Update(i)
}

// Stackable implements the Item interface.
func (i *BaseItem) Stackable() bool { return i.stackable }

// Movable implements the Item interface
func (i *BaseItem) Movable() bool { return !i.fixed }

// Amount implements the Item interface.
func (i *BaseItem) Amount() int { return i.amount }

// SetAmount implements the Item interface.
func (i *BaseItem) SetAmount(n int) {
	if i.amount == n {
		return
	}
	i.InvalidateOPL()
	if n < 1 {
		i.amount = 1
	} else if n > int(uo.MaxStackAmount) {
		i.amount = int(uo.MaxStackAmount)
	} else {
		i.amount = n
	}
}

// Value implements the Item interface.
func (i *BaseItem) Value() int {
	return i.value
}

// Consume implements the Item interface.
func (i *BaseItem) Consume(n int) bool {
	if n < 1 {
		return true
	}
	if n > i.amount {
		return false
	}
	i.amount -= n
	if i.amount == 0 {
		Remove(i)
	} else {
		i.InvalidateOPL()
	}
	return true
}

// DisplayName implements the Object interface.
func (i *BaseItem) DisplayName() string {
	if i.amount > 1 {
		if i.plural != "" {
			return strconv.Itoa(i.amount) + " " + i.plural
		}
		return strconv.Itoa(i.amount) + " " + i.name
	}
	return i.BaseObject.DisplayName()
}

// SingleClick implements the Object interface
func (o *BaseItem) SingleClick(from Mobile) {
	// Default action is to send the name as over-head text
	if from.NetState() != nil {
		from.NetState().Speech(o, o.DisplayName())
	}
}

// Split implements the Item interface.
func (i *BaseItem) Split(n int) Item {
	// No new item required in these cases
	if !i.stackable || n < 1 || n >= i.amount {
		return nil
	}
	// Create the new item
	item := template.Create[Item](i.templateName)
	// Remove this item from its parent
	failed := false
	if i.parent == nil {
		failed = !world.Map().RemoveObject(i)
	} else {
		failed = !i.parent.RemoveObject(i)
	}
	if failed {
		return nil
	}
	item.SetAmount(i.amount - n)
	i.SetAmount(n)
	// Force the remainder back where we came from
	item.SetLocation(i.location)
	item.SetDropLocation(i.location)
	if i.parent == nil {
		world.Map().ForceAddObject(item)
	} else {
		i.parent.ForceAddObject(item)
	}
	i.parent = item
	// Update parents if needed
	if container, ok := i.parent.(Container); ok {
		container.AdjustWeightAndCount(i.weight*float32(-n), -n)
	}
	return item
}

// Combine implements the Item interface.
func (i *BaseItem) Combine(other Item) bool {
	if !i.CanCombineWith(other) {
		return false
	}
	if i.amount+other.Amount() > int(uo.MaxStackAmount) {
		return false
	}
	// Update stack amounts
	other.SetAmount(other.Amount() + i.amount)
	other.SetLocation(i.location)
	other.SetDropLocation(i.location)
	iparent := i.parent
	if iparent == nil {
		world.Map().ForceRemoveObject(i)
		world.Map().ForceAddObject(other)
	} else {
		iparent.ForceRemoveObject(i)
		iparent.ForceAddObject(other)
	}
	Remove(i)
	i.InvalidateOPL()
	return true
}

// CanCombineWith implements the Item interface.
func (i *BaseItem) CanCombineWith(item Item) bool {
	if !i.stackable {
		return false
	}
	if i.templateName != item.TemplateName() {
		return false
	}
	if i.amount+item.Amount() > int(uo.MaxStackAmount) {
		return false
	}
	return true
}

// Height implements the Item interface.
func (i *BaseItem) Height() int8 { return i.def.Height }

// Highest returns the highest elevation of the object
func (i *BaseItem) Highest() int8 {
	return i.location.Z + i.def.Height
}

// StandingHeight returns the standing height based on the object's flags.
func (i *BaseItem) StandingHeight() int8 {
	if !i.Surface() && !i.Wet() && !i.Impassable() {
		return 0
	}
	if i.Bridge() {
		return i.def.Height / 2
	}
	return i.def.Height
}

// Weight implements the Object interface
func (i *BaseItem) Weight() float32 { return i.weight * float32(i.amount) }

// Z returns the permanent Z location of the tile
func (i *BaseItem) Z() int8 { return i.location.Z }

// DropLocation implements the Item interface
func (i *BaseItem) DropLocation() uo.Location { return i.dropLocation }

// SetDropLocation implements the Item interface
func (i *BaseItem) SetDropLocation(l uo.Location) { i.dropLocation = l }

// RefreshDecayDeadline implements the Item interface
func (i *BaseItem) RefreshDecayDeadline() {
	if i.BaseObject.spawnerRegion != nil {
		// If we are being managed by a spawner we don't decay
		i.decayDeadline = uo.TimeNever
		return
	}
	switch i.lootType {
	case uo.LootTypeNormal:
		i.decayDeadline = world.Time() + uo.DurationHour
	case uo.LootTypeBlessed:
		i.decayDeadline = world.Time() + uo.DurationHour
	case uo.LootTypeNewbied:
		i.decayDeadline = world.Time() + uo.DurationSecond*15
	case uo.LootTypeSystem:
		i.decayDeadline = uo.TimeNever
	}
}

// Update implements the Object interface.
func (i *BaseItem) Update(t uo.Time) {
	if t >= i.decayDeadline {
		if i.BaseObject.spawnerRegion != nil {
			// If we are being managed by a spawner we don't decay
			return
		}
		Remove(i)
	}
}

// LiftSound implements the Item interface.
func (i *BaseItem) LiftSound() uo.Sound { return i.liftSound }

// DropSoundOverride implements the Item interface.
func (i *BaseItem) DropSoundOverride(s uo.Sound) uo.Sound {
	if i.dropSoundOverride != uo.SoundInvalidDrop {
		return i.dropSoundOverride
	}
	return s
}

// Uses implements the Item interface.
func (i *BaseItem) Uses() int { return i.uses }

// ConsumeUse implements the Item interface.
func (i *BaseItem) ConsumeUse() bool {
	if i.uses < 1 {
		return false
	}
	i.uses--
	i.InvalidateOPL()
	return true
}

// AppendOPLEntries implements the Object interface.
func (i *BaseItem) AppendOPLEntires(r Object, p *serverpacket.OPLPacket) {
	i.BaseObject.AppendOPLEntires(r, p)
	if i.uses > 0 {
		p.Append(fmt.Sprintf("Uses: %d", i.uses), true)
	}
}

// FlippedGraphic implements the Item interface.
func (i *BaseItem) FlippedGraphic() uo.Graphic { return i.flippedGraphic }

// SetFlippedGraphic implements the Item interface.
func (i *BaseItem) SetFlippedGraphic(g uo.Graphic) {
	i.flippedGraphic = g
	world.Update(i)
}

// SetDefForGraphic implements the Item interface.
func (i *BaseItem) SetDefForGraphic(g uo.Graphic) {
	i.def = world.GetItemDefinition(g)
	world.Update(i)
}

// NoRent implements the Object interface.
func (i *BaseItem) NoRent() bool { return i.noRent }

// Flag accessors
func (i *BaseItem) Background() bool   { return i.def.TileFlags&uo.TileFlagsBackground != 0 }
func (i *BaseItem) Weapon() bool       { return i.def.TileFlags&uo.TileFlagsWeapon != 0 }
func (i *BaseItem) Transparent() bool  { return i.def.TileFlags&uo.TileFlagsTransparent != 0 }
func (i *BaseItem) Translucent() bool  { return i.def.TileFlags&uo.TileFlagsTranslucent != 0 }
func (i *BaseItem) Wall() bool         { return i.def.TileFlags&uo.TileFlagsWall != 0 }
func (i *BaseItem) Damaging() bool     { return i.def.TileFlags&uo.TileFlagsDamaging != 0 }
func (i *BaseItem) Impassable() bool   { return i.def.TileFlags&uo.TileFlagsImpassable != 0 }
func (i *BaseItem) Wet() bool          { return i.def.TileFlags&uo.TileFlagsWet != 0 }
func (i *BaseItem) Surface() bool      { return i.def.TileFlags&uo.TileFlagsSurface != 0 }
func (i *BaseItem) Bridge() bool       { return i.def.TileFlags&uo.TileFlagsBridge != 0 }
func (i *BaseItem) Generic() bool      { return i.def.TileFlags&uo.TileFlagsGeneric != 0 }
func (i *BaseItem) Window() bool       { return i.def.TileFlags&uo.TileFlagsWindow != 0 }
func (i *BaseItem) NoShoot() bool      { return i.def.TileFlags&uo.TileFlagsNoShoot != 0 }
func (i *BaseItem) ArticleA() bool     { return i.def.TileFlags&uo.TileFlagsArticleA != 0 }
func (i *BaseItem) ArticleAn() bool    { return i.def.TileFlags&uo.TileFlagsArticleAn != 0 }
func (i *BaseItem) Internal() bool     { return i.def.TileFlags&uo.TileFlagsInternal != 0 }
func (i *BaseItem) Foliage() bool      { return i.def.TileFlags&uo.TileFlagsFoliage != 0 }
func (i *BaseItem) PartialHue() bool   { return i.def.TileFlags&uo.TileFlagsPartialHue != 0 }
func (i *BaseItem) NoHouse() bool      { return i.def.TileFlags&uo.TileFlagsNoHouse != 0 }
func (i *BaseItem) Map() bool          { return i.def.TileFlags&uo.TileFlagsMap != 0 }
func (i *BaseItem) Container() bool    { return i.def.TileFlags&uo.TileFlagsContainer != 0 }
func (i *BaseItem) Wearable() bool     { return i.def.TileFlags&uo.TileFlagsWearable != 0 }
func (i *BaseItem) LightSource() bool  { return i.def.TileFlags&uo.TileFlagsLightSource != 0 }
func (i *BaseItem) Animation() bool    { return i.def.TileFlags&uo.TileFlagsAnimation != 0 }
func (i *BaseItem) NoDiagonal() bool   { return i.def.TileFlags&uo.TileFlagsNoDiagonal != 0 }
func (i *BaseItem) Armor() bool        { return i.def.TileFlags&uo.TileFlagsArmor != 0 }
func (i *BaseItem) Roof() bool         { return i.def.TileFlags&uo.TileFlagsRoof != 0 }
func (i *BaseItem) Door() bool         { return i.def.TileFlags&uo.TileFlagsDoor != 0 }
func (i *BaseItem) StairBack() bool    { return i.def.TileFlags&uo.TileFlagsStairBack != 0 }
func (i *BaseItem) StairRight() bool   { return i.def.TileFlags&uo.TileFlagsStairRight != 0 }
func (i *BaseItem) AlphaBlend() bool   { return i.def.TileFlags&uo.TileFlagsAlphaBlend != 0 }
func (i *BaseItem) UseNewArt() bool    { return i.def.TileFlags&uo.TileFlagsUseNewArt != 0 }
func (i *BaseItem) ArtUsed() bool      { return i.def.TileFlags&uo.TileFlagsArtUsed != 0 }
func (i *BaseItem) NoShadow() bool     { return i.def.TileFlags&uo.TileFlagsBackground != 0 }
func (i *BaseItem) PixelBleed() bool   { return i.def.TileFlags&uo.TileFlagsPixelBleed != 0 }
func (i *BaseItem) PlayAnimOnce() bool { return i.def.TileFlags&uo.TileFlagsPlayAnimOnce != 0 }
func (i *BaseItem) MultiMovable() bool { return i.def.TileFlags&uo.TileFlagsMultiMovable != 0 }
