package game

import "github.com/qbradq/sharduo/lib/uo"

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

// GetLayer implements the Item interface.
func (i *BaseItem) GetLayer() uo.Layer { return i.Layer }

// GetGraphic implements the Item interface.
func (i *BaseItem) GetGraphic() uo.Item { return i.Graphic }
