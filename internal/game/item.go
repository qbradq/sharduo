package game

import (
	"log"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseItem{} })
}

// Item is the interface that all non-static items implement.
type Item interface {
	Object
	// BaseGraphic returns the graphic of the item
	BaseGraphic() uo.Graphic
	// GraphicOffset returns the current graphic offset of the item, this will
	// range 0-255 inclusive.
	GraphicOffset() int
	// Dyable returns true if the item's hue can be changed by the player
	Dyable() bool
	// Flippable returns true if the item can be flipped / turned
	Flippable() bool
	// Stackable returns true if the item can be stacked
	Stackable() bool
	// Amount of the stack
	Amount() int
	// SetAmount sets the amount of the stack. If this is out of range it will
	// be bounded to a sane value
	SetAmount(int)
	// Split splits off a number of the items in the stack and returns a new
	// copy of the item with that amount. The amount of this item is adjusted
	// accordingly. If the amount is less than one nil is returned. If the
	// amount is greater than or equal to the number of items in this stack,
	// this item is returned an a new one is not created. For non-stackable
	// items this will always return this item. New objects created by this
	// function have their parent set as this item upon return.
	Split(int) Item
	// Combine adds the stack count of the given item to this item, then
	// destroys the given item. Returns false on failure. Failure can happen
	// if this item does not support stacking - see Stackable() - the items are
	// not from the same template object, or the combined amounts would be
	// greater than uo.MaxStackAmount. If this function returns true the other
	// object has been totally removed from the world and data stores.
	Combine(Item) bool
	// Height returns the height of the item
	Height() int
	// Z returns the permanent Z location of the object
	Z() int
	// DropLocation returns the requested drop location of an item
	DropLocation() uo.Location
	// SetDropLocation sets the requested drop location of an item
	SetDropLocation(uo.Location)
	// Flag accessors
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
	// Static definition that holds the static data for the item
	def *uo.StaticDefinition
	// Graphic of the item
	graphic uo.Graphic
	// Graphic of the item when flipped. If this is uo.GraphicNone the item cannot
	// be flipped.
	flippedGraphic uo.Graphic
	// Flipped is true if the item is currently flipped.
	flipped bool
	// Dyable flag
	dyable bool
	// Base weight of the item
	weight float32
	// Stackable flag
	stackable bool
	// Stack amount
	amount int

	//
	// Non-persistent values
	//

	// Drop request location
	dropLocation uo.Location
}

// TypeName implements the util.Serializeable interface.
func (i *BaseItem) TypeName() string {
	return "BaseItem"
}

// Serialize implements the util.Serializeable interface.
func (i *BaseItem) Serialize(f *util.TagFileWriter) {
	i.BaseObject.Serialize(f)
	f.WriteHex("Graphic", uint32(i.graphic))
	f.WriteHex("FlippedGraphic", uint32(i.flippedGraphic))
	f.WriteBool("Flipped", i.flipped)
	f.WriteBool("Dyable", i.dyable)
	f.WriteFloat("Weight", i.weight)
	f.WriteBool("Stackable", i.stackable)
	f.WriteNumber("Amount", i.amount)
}

// Deserialize implements the util.Serializeable interface.
func (i *BaseItem) Deserialize(f *util.TagFileObject) {
	i.BaseObject.Deserialize(f)
	i.graphic = uo.Graphic(f.GetNumber("Graphic", int(uo.GraphicDefault)))
	i.def = world.GetItemDefinition(i.graphic)
	i.flippedGraphic = uo.Graphic(f.GetNumber("FlippedGraphic", int(uo.GraphicNone)))
	i.flipped = f.GetBool("Flipped", false)
	i.dyable = f.GetBool("Dyable", false)
	i.weight = f.GetFloat("Weight", 255.0)
	i.stackable = f.GetBool("Stackable", false)
	i.amount = f.GetNumber("Amount", 1)
}

// BaseGraphic implements the Item interface.
func (i *BaseItem) BaseGraphic() uo.Graphic { return i.graphic }

// GraphicOffset implements the Item interface.
func (i *BaseItem) GraphicOffset() int {
	return 0
}

// Dyable implements the Item interface.
func (i *BaseItem) Dyable() bool { return i.dyable }

// Flippable implements the Item interface.
func (i *BaseItem) Flippable() bool { return i.flippedGraphic != uo.GraphicNone }

// Stackable implements the Item interface.
func (i *BaseItem) Stackable() bool { return i.stackable }

// Amount implements the Item interface.
func (i *BaseItem) Amount() int { return i.amount }

// SetAmount implements the Item interface.
func (i *BaseItem) SetAmount(n int) {
	if n < 1 {
		i.amount = 1
	} else if n > int(uo.MaxStackAmount) {
		i.amount = int(uo.MaxStackAmount)
	} else {
		i.amount = n
	}
}

// Split implements the Item interface.
func (i *BaseItem) Split(n int) Item {
	// Sanity checking
	if !i.stackable {
		return i
	}
	// Bounds checking
	if n < 1 {
		return nil
	}
	if n >= i.amount {
		return i
	}
	// Create the new item
	ni := world.New(i.templateName)
	if ni == nil {
		log.Println("error: Item.Split() failed to create duplicate item")
		return nil
	}
	item, ok := ni.(Item)
	if !ok {
		return nil
	}
	// Update stack amounts
	i.amount -= n
	item.SetAmount(n)

	return item
}

// Combine implements the Item interface.
func (i *BaseItem) Combine(other Item) bool {
	// Sanity checking
	if !i.stackable {
		return false
	}
	if i.templateName != other.TemplateName() {
		return false
	}
	if i.amount+other.Amount() > int(uo.MaxStackAmount) {
		return false
	}
	// Update stack amounts
	i.amount += other.Amount()
	world.Remove(other)
	return true
}

// Height implements the Item interface.
func (i *BaseItem) Height() int { return i.def.Height }

// Weight implements the Object interface
func (i *BaseItem) Weight() float32 {
	return i.weight * float32(i.amount)
}

// Z returns the permanent Z location of the tile
func (i *BaseItem) Z() int {
	return uo.BoundZ(i.location.Z)
}

// DropLocation implements the Item interface
func (i *BaseItem) DropLocation() uo.Location { return i.dropLocation }

// SetDropLocation implements the Item interface
func (i *BaseItem) SetDropLocation(l uo.Location) { i.dropLocation = l }

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
