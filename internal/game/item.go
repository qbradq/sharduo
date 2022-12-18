package game

import (
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseItem{} })
}

// Item is the interface that all non-static items implement.
type Item interface {
	Object
	// Graphic returns the graphic of the item
	Graphic() uo.Graphic
	// Dyable returns true if the item's hue can be changed by the player
	Dyable() bool
	// Flippable returns true if the item can be flipped / turned
	Flippable() bool
	// Stackable returns true if the item can be stacked
	Stackable() bool
	// Amount of the stack
	Amount() int
}

// BaseItem provides the basic implementation of Item.
type BaseItem struct {
	BaseObject
	// Graphic of the item
	graphic uo.Graphic
	// Graphic of the item when flipped. If this is uo.ItemNone the item cannot
	// be flipped.
	flippedGraphic uo.Graphic
	// Dyable flag
	dyable bool
	// Stackable flag
	stackable bool
	// Stack amount
	amount int
}

// TypeName implements the util.Serializeable interface.
func (i *BaseItem) TypeName() string {
	return "BaseItem"
}

// Serialize implements the util.Serializeable interface.
func (i *BaseItem) Serialize(f *util.TagFileWriter) {
	i.BaseObject.Serialize(f)
	f.WriteHex("Graphic", int(i.graphic))
}

// Deserialize implements the util.Serializeable interface.
func (i *BaseItem) Deserialize(f *util.TagFileObject) {
	i.BaseObject.Deserialize(f)
	i.graphic = uo.Graphic(f.GetNumber("Graphic", 0))
}

// Graphic implements the Item interface.
func (i *BaseItem) Graphic() uo.Graphic { return i.graphic }

// Dyable implements the Item interface.
func (i *BaseItem) Dyable() bool { return i.dyable }

// Flippable implements the Item interface.
func (i *BaseItem) Flippable() bool { return i.flippedGraphic != uo.ItemNone }

// Stackable implements the Item interface.
func (i *BaseItem) Stackable() bool { return i.stackable }

// Amount implements the Item interface.
func (i *BaseItem) Amount() int { return i.amount }
