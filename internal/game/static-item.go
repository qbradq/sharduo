package game

import (
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

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
func (i *StaticItem) LinkEvent(s string, fn *func(Object, Object))      {}
func (i *StaticItem) GetEventHandler(s string) *func(Object, Object)    { return nil }
func (i *StaticItem) RecalculateStats()                                 {}
func (i *StaticItem) RemoveObject(o Object) bool                        { return false }
func (i *StaticItem) AddObject(o Object) bool                           { return false }
func (i *StaticItem) ForceAddObject(o Object)                           {}
func (i *StaticItem) ForceRemoveObject(o Object)                        {}
func (i *StaticItem) DropObject(o Object, l uo.Location, m Mobile) bool { return false }
func (i *StaticItem) SingleClick(m Mobile)                              {}
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
