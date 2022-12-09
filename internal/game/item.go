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
	// Layer returns the layer of the item
	Layer() uo.Layer
	// Graphic returns the graphic of the item
	Graphic() uo.Item
}

// BaseItem provides the basic implementation of Item.
type BaseItem struct {
	BaseObject
	// Item graphic of the item, if any
	graphic uo.Item
	// Wearable is true if the item is wearable
	wearable bool
	// Layer is the layer of the wearable
	layer uo.Layer
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
	f.WriteBool("Wearable", i.wearable)
	f.WriteNumber("Layer", int(i.layer))
}

// Deserialize implements the util.Serializeable interface.
func (i *BaseItem) Deserialize(f *util.TagFileObject) {
	i.BaseObject.Deserialize(f)
	i.graphic = uo.Item(f.GetNumber("Graphic", 0))
	i.wearable = f.GetBool("Wearable", false)
	i.layer = uo.Layer(f.GetNumber("Layer", int(uo.LayerInvalid)))
}

// Layer implements the Item interface.
func (i *BaseItem) Layer() uo.Layer { return i.layer }

// Graphic implements the Item interface.
func (i *BaseItem) Graphic() uo.Item { return i.graphic }
