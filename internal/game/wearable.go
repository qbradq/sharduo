package game

import (
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseWearable{} })
}

// Wearable represents an item that can be worn by a humanoid mobile
type Wearable interface {
	Item
	// Layer returns the layer of the item
	Layer() uo.Layer
}

// BaseWearable provides the most common implementation of Wearable
type BaseWearable struct {
	BaseItem
	// Layer is the layer of the wearable
	layer uo.Layer
}

// TypeName implements the util.Serializeable interface.
func (i *BaseWearable) TypeName() string {
	return "BaseWearable"
}

// Serialize implements the util.Serializeable interface.
func (i *BaseWearable) Serialize(f *util.TagFileWriter) {
	i.BaseItem.Serialize(f)
	f.WriteNumber("Layer", int(i.layer))
}

// Deserialize implements the util.Serializeable interface.
func (i *BaseWearable) Deserialize(f *util.TagFileObject) {
	i.BaseItem.Deserialize(f)
	i.layer = uo.Layer(f.GetNumber("Layer", int(uo.LayerInvalid)))
}

// Layer implements the Item interface.
func (i *BaseWearable) Layer() uo.Layer { return i.layer }
