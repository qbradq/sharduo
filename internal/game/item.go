package game

import (
	"encoding/gob"

	"github.com/qbradq/sharduo/internal/util"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	gob.Register(BaseItem{})
}

// Item is the interface that all non-static items implement.
type Item interface {
	Object
	// GetLayer returns the layer of the item
	GetLayer() uo.Layer
	// GetGraphic returns the graphic of the item
	GetGraphic() uo.Item
}

// BaseItem provides the basic implementation of Item.
type BaseItem struct {
	BaseObject
	// Item graphic of the item, if any
	Graphic uo.Item
	// Wearable is true if the item is wearable
	Wearable bool
	// Layer is the layer of the wearable
	Layer uo.Layer
}

// GetTypeName implements the util.Serializeable interface.
func (i *BaseItem) GetTypeName() string {
	return "BaseItem"
}

// Serialize implements the util.Serializeable interface.
func (i *BaseItem) Serialize(f *util.TagFileWriter) {
	i.BaseObject.Serialize(f)
}

// Deserialize implements the util.Serializeable interface.
func (i *BaseItem) Deserialize(f *util.TagFileObject) {
	i.BaseObject.Deserialize(f)
}

// GetLayer implements the Item interface.
func (i *BaseItem) GetLayer() uo.Layer { return i.Layer }

// GetGraphic implements the Item interface.
func (i *BaseItem) GetGraphic() uo.Item { return i.Graphic }
