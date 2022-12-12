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
	Graphic() uo.Item
	// Dyable returns true if the item's hue can be changed by the player
	Dyable() bool
	// Flippable returns true if the item can be flipped / turned
	Flippable() bool
}

// BaseItem provides the basic implementation of Item.
type BaseItem struct {
	BaseObject
	// Graphic of the item
	graphic uo.Item
	// Graphic of the item when flipped. If this is uo.ItemNone the item cannot
	// be flipped.
	flippedGraphic uo.Item
	// Dyable flag
	dyable bool
}

// TypeName implements the util.Serializeable interface.
func (i *BaseItem) TypeName() string {
	return "BaseItem"
}

// SerialType implements the util.Serializeable interface.
func (i *BaseItem) SerialType() uo.SerialType {
	return uo.SerialTypeItem
}

// Serialize implements the util.Serializeable interface.
func (i *BaseItem) Serialize(f *util.TagFileWriter) {
	i.BaseObject.Serialize(f)
	f.WriteHex("Graphic", int(i.graphic))
}

// Deserialize implements the util.Serializeable interface.
func (i *BaseItem) Deserialize(f *util.TagFileObject) {
	i.BaseObject.Deserialize(f)
	i.graphic = uo.Item(f.GetNumber("Graphic", 0))
}

// Graphic implements the Item interface.
func (i *BaseItem) Graphic() uo.Item { return i.graphic }

// Dyable implements the Item interface.
func (i *BaseItem) Dyable() bool { return i.dyable }

// Flippable implements the Item interface.
func (i *BaseItem) Flippable() bool { return i.flippedGraphic != uo.ItemNone }
