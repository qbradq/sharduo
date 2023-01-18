package game

import (
	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.Add("StaticItem", func() Object { return &StaticItem{} })
	marshal.RegisterCtor(marshal.ObjectTypeStatic, func() interface{} { return &StaticItem{} })
}

// StaticItem is a light-weight Item implementation intended to be used for
// non-functional decorative items.
type StaticItem struct {
	util.BaseSerializeable
	// Graphic of the item
	graphic uo.Graphic
	// Item definition
	def *uo.StaticDefinition
	// Location
	location uo.Location
	// Hue
	hue uo.Hue
}

// TypeName implements the util.Serializeable interface.
func (o *StaticItem) TypeName() string {
	return "StaticItem"
}

// ObjectType implements the Object interface.
func (o *StaticItem) ObjectType() marshal.ObjectType { return marshal.ObjectTypeStatic }

// SetObjectType implements the Object interface.
func (o *StaticItem) SetObjectType(t marshal.ObjectType) {}

// SerialType implements the util.Serializeable interface.
func (o *StaticItem) SerialType() uo.SerialType {
	return uo.SerialTypeItem
}

// Item interface
// BaseGraphic implements the Item interface.
func (i *StaticItem) BaseGraphic() uo.Graphic { return i.graphic }

// SetBaseGraphic implements the Item interface.
func (i *StaticItem) SetBaseGraphic(g uo.Graphic) {
	i.graphic = g
	i.def = world.GetItemDefinition(g)
}
func (i *StaticItem) GraphicOffset() int            { return 0 }
func (i *StaticItem) Dyable() bool                  { return false }
func (i *StaticItem) Flippable() bool               { return false }
func (i *StaticItem) Stackable() bool               { return false }
func (i *StaticItem) Movable() bool                 { return false }
func (i *StaticItem) Amount() int                   { return i.def.Count }
func (i *StaticItem) SetAmount(int)                 {}
func (i *StaticItem) Consume(n int) bool            { return false }
func (i *StaticItem) Split(n int) Item              { return nil }
func (i *StaticItem) Combine(item Item) bool        { return false }
func (i *StaticItem) CanCombineWith(Item) bool      { return false }
func (i *StaticItem) Height() int                   { return i.def.Height }
func (i *StaticItem) Z() int                        { return i.location.Z }
func (i *StaticItem) DropLocation() uo.Location     { return uo.RandomContainerLocation }
func (i *StaticItem) SetDropLocation(l uo.Location) {}

// Object interface
func (i *StaticItem) Parent() Object                                    { return nil }
func (i *StaticItem) HasParent(o Object) bool                           { return o == nil }
func (i *StaticItem) SetParent(o Object)                                {}
func (i *StaticItem) TemplateName() string                              { return "StaticItem" }
func (i *StaticItem) LinkEvent(event, handler string)                   {}
func (i *StaticItem) GetEventHandler(s string) *func(Object, Object)    { return nil }
func (i *StaticItem) RecalculateStats()                                 {}
func (i *StaticItem) RemoveObject(o Object) bool                        { return false }
func (i *StaticItem) AddObject(o Object) bool                           { return false }
func (i *StaticItem) ForceAddObject(o Object)                           {}
func (i *StaticItem) ForceRemoveObject(o Object)                        {}
func (i *StaticItem) DropObject(o Object, l uo.Location, m Mobile) bool { return false }
func (i *StaticItem) SingleClick(m Mobile)                              { defaultSingleClickHandler(i, m) }
func (i *StaticItem) Location() uo.Location                             { return i.location }
func (i *StaticItem) SetLocation(l uo.Location)                         { i.location = l }
func (i *StaticItem) Hue() uo.Hue                                       { return i.hue }
func (i *StaticItem) Facing() uo.Direction                              { return uo.DirectionNorth }
func (i *StaticItem) SetFacing(d uo.Direction)                          {}
func (i *StaticItem) DisplayName() string                               { return i.def.Name }
func (i *StaticItem) Weight() float32                                   { return 255.0 }

// Serialize implements the util.Serializeable interface.
func (i *StaticItem) Serialize(f *util.TagFileWriter) {
	i.BaseSerializeable.Serialize(f)
	// Owned properties
	f.WriteNumber("Graphic", int(i.graphic))
	f.WriteNumber("Hue", int(i.hue))
	f.WriteLocation("Location", i.location)
}

// An empty event map so we don't have to allocate one every time.
var _emptyEventHandlers = make(map[string]string)

// Marshal implements the marshal.Marshaler interface.
func (i *StaticItem) Marshal(s *marshal.TagFileSegment) {
	s.PutObjectHeader(
		marshal.ObjectTypeStatic,
		i.Serial(),
		"StaticItem",
		uo.SerialSystem,
		"",
		i.hue,
		i.location,
		_emptyEventHandlers)
	s.PutShort(uint16(i.graphic))
}

// Deserialize implements the util.Serializeable interface.
func (i *StaticItem) Deserialize(f *util.TagFileObject) {
	i.BaseSerializeable.Deserialize(f)
	// Owned properties
	i.graphic = uo.Graphic(f.GetNumber("Graphic", int(uo.GraphicDefault)))
	i.def = world.GetItemDefinition(i.graphic)
	i.hue = uo.Hue(f.GetNumber("Hue", int(uo.HueDefault)))
	i.location = f.GetLocation("Location", uo.Location{
		X: 1324,
		Y: 1624,
		Z: 55,
	})
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *StaticItem) Unmarshal(to *marshal.TagObject) {
	i.SetSerial(to.Serial)
	i.hue = to.Hue
	i.location = to.Location
	i.graphic = uo.Graphic(to.Tags.Short(marshal.TagGraphic))
	i.def = world.GetItemDefinition(i.graphic)
}

// AfterUnmarshal implements the marshal.Unmarshaler interface.
func (i *StaticItem) AfterUnmarshal(to *marshal.TagObject) {}

// Flag accessors
func (i *StaticItem) Background() bool   { return i.def.TileFlags&uo.TileFlagsBackground != 0 }
func (i *StaticItem) Weapon() bool       { return i.def.TileFlags&uo.TileFlagsWeapon != 0 }
func (i *StaticItem) Transparent() bool  { return i.def.TileFlags&uo.TileFlagsTransparent != 0 }
func (i *StaticItem) Translucent() bool  { return i.def.TileFlags&uo.TileFlagsTranslucent != 0 }
func (i *StaticItem) Wall() bool         { return i.def.TileFlags&uo.TileFlagsWall != 0 }
func (i *StaticItem) Damaging() bool     { return i.def.TileFlags&uo.TileFlagsDamaging != 0 }
func (i *StaticItem) Impassable() bool   { return i.def.TileFlags&uo.TileFlagsImpassable != 0 }
func (i *StaticItem) Wet() bool          { return i.def.TileFlags&uo.TileFlagsWet != 0 }
func (i *StaticItem) Surface() bool      { return i.def.TileFlags&uo.TileFlagsSurface != 0 }
func (i *StaticItem) Bridge() bool       { return i.def.TileFlags&uo.TileFlagsBridge != 0 }
func (i *StaticItem) Generic() bool      { return i.def.TileFlags&uo.TileFlagsGeneric != 0 }
func (i *StaticItem) Window() bool       { return i.def.TileFlags&uo.TileFlagsWindow != 0 }
func (i *StaticItem) NoShoot() bool      { return i.def.TileFlags&uo.TileFlagsNoShoot != 0 }
func (i *StaticItem) ArticleA() bool     { return i.def.TileFlags&uo.TileFlagsArticleA != 0 }
func (i *StaticItem) ArticleAn() bool    { return i.def.TileFlags&uo.TileFlagsArticleAn != 0 }
func (i *StaticItem) Internal() bool     { return i.def.TileFlags&uo.TileFlagsInternal != 0 }
func (i *StaticItem) Foliage() bool      { return i.def.TileFlags&uo.TileFlagsFoliage != 0 }
func (i *StaticItem) PartialHue() bool   { return i.def.TileFlags&uo.TileFlagsPartialHue != 0 }
func (i *StaticItem) NoHouse() bool      { return i.def.TileFlags&uo.TileFlagsNoHouse != 0 }
func (i *StaticItem) Map() bool          { return i.def.TileFlags&uo.TileFlagsMap != 0 }
func (i *StaticItem) Container() bool    { return i.def.TileFlags&uo.TileFlagsContainer != 0 }
func (i *StaticItem) Wearable() bool     { return i.def.TileFlags&uo.TileFlagsWearable != 0 }
func (i *StaticItem) LightSource() bool  { return i.def.TileFlags&uo.TileFlagsLightSource != 0 }
func (i *StaticItem) Animation() bool    { return i.def.TileFlags&uo.TileFlagsAnimation != 0 }
func (i *StaticItem) NoDiagonal() bool   { return i.def.TileFlags&uo.TileFlagsNoDiagonal != 0 }
func (i *StaticItem) Armor() bool        { return i.def.TileFlags&uo.TileFlagsArmor != 0 }
func (i *StaticItem) Roof() bool         { return i.def.TileFlags&uo.TileFlagsRoof != 0 }
func (i *StaticItem) Door() bool         { return i.def.TileFlags&uo.TileFlagsDoor != 0 }
func (i *StaticItem) StairBack() bool    { return i.def.TileFlags&uo.TileFlagsStairBack != 0 }
func (i *StaticItem) StairRight() bool   { return i.def.TileFlags&uo.TileFlagsStairRight != 0 }
func (i *StaticItem) AlphaBlend() bool   { return i.def.TileFlags&uo.TileFlagsAlphaBlend != 0 }
func (i *StaticItem) UseNewArt() bool    { return i.def.TileFlags&uo.TileFlagsUseNewArt != 0 }
func (i *StaticItem) ArtUsed() bool      { return i.def.TileFlags&uo.TileFlagsArtUsed != 0 }
func (i *StaticItem) NoShadow() bool     { return i.def.TileFlags&uo.TileFlagsBackground != 0 }
func (i *StaticItem) PixelBleed() bool   { return i.def.TileFlags&uo.TileFlagsPixelBleed != 0 }
func (i *StaticItem) PlayAnimOnce() bool { return i.def.TileFlags&uo.TileFlagsPlayAnimOnce != 0 }
func (i *StaticItem) MultiMovable() bool { return i.def.TileFlags&uo.TileFlagsMultiMovable != 0 }
