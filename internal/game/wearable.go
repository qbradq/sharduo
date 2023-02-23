package game

import (
	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	objctors["BaseWearable"] = func() Object { return &BaseWearable{} }
	marshal.RegisterCtor(marshal.ObjectTypeWearable, func() interface{} { return &BaseWearable{} })
}

// Layerer represents an item that can be layered onto an equippable mobile.
type Layerer interface {
	// Layer returns the layer of the object
	Layer() uo.Layer
}

// Wearable represents an item that can be worn by a humanoid mobile
type Wearable interface {
	Item
	Layerer
}

// BaseWearable provides the most common implementation of Wearable
type BaseWearable struct {
	BaseItem
	// Layer is the layer of the wearable
	layer uo.Layer
}

// ObjectType implements the Object interface.
func (i *BaseWearable) ObjectType() marshal.ObjectType { return marshal.ObjectTypeWearable }

// Deserialize implements the util.Serializeable interface.
func (i *BaseWearable) Deserialize(f *util.TagFileObject) {
	i.BaseItem.Deserialize(f)
	i.layer = uo.Layer(f.GetNumber("Layer", int(uo.LayerInvalid)))
}

// Layer implements the Item interface.
func (i *BaseWearable) Layer() uo.Layer { return i.layer }
