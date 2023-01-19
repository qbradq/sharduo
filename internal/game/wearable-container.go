package game

import (
	"log"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	objctors["WearableContainer"] = func() Object { return &WearableContainer{} }
	marshal.RegisterCtor(marshal.ObjectTypeWearableContainer, func() interface{} { return &WearableContainer{} })
}

// WearableContainer is a wearable item with the properties of a container, such
// as inventory backpacks and the player's bank box.
type WearableContainer struct {
	BaseContainer
	// Layer is the layer of the wearable
	layer uo.Layer
}

// ObjectType implements the Object interface.
func (i *WearableContainer) ObjectType() marshal.ObjectType {
	return marshal.ObjectTypeWearableContainer
}

// Marshal implements the marshal.Marshaler interface.
func (i *WearableContainer) Marshal(s *marshal.TagFileSegment) {
	i.BaseContainer.Marshal(s)
	s.PutTag(marshal.TagLayer, marshal.TagValueByte, byte(i.layer))
}

// Deserialize implements the util.Serializeable interface.
func (s *WearableContainer) Deserialize(f *util.TagFileObject) {
	s.BaseContainer.Deserialize(f)
	s.layer = uo.Layer(f.GetNumber("Layer", int(uo.LayerInvalid)))
	if s.layer == uo.LayerInvalid {
		log.Printf("error: wearable container %s no layer given", s.Serial().String())
	}
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *WearableContainer) Unmarshal(s *marshal.TagFileSegment) *marshal.TagCollection {
	tags := i.BaseContainer.Unmarshal(s)
	i.layer = uo.Layer(tags.Byte(marshal.TagLayer))
	return tags
}

// Layer implements the Layerer interface.
func (c *WearableContainer) Layer() uo.Layer { return c.layer }
