package game

import (
	"log"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &WearableContainer{} })
	objectCtors[marshal.ObjectTypeWearableContainer] = func() Object { return &WearableContainer{} }
}

// WearableContainer is a wearable item with the properties of a container, such
// as inventory backpacks and the player's bank box.
type WearableContainer struct {
	BaseContainer
	// Layer is the layer of the wearable
	layer uo.Layer
}

// TypeName implements the util.Serializeable interface.
func (o *WearableContainer) TypeName() string {
	return "WearableContainer"
}

// ObjectType implements the Object interface.
func (i *WearableContainer) ObjectType() marshal.ObjectType {
	return marshal.ObjectTypeWearableContainer
}

// Serialize implements the util.Serializeable interface.
func (s *WearableContainer) Serialize(f *util.TagFileWriter) {
	s.BaseContainer.Serialize(f)
	f.WriteNumber("Layer", int(s.layer))
}

// Marshal implements the marshal.Marshaler interface.
func (i *WearableContainer) Marshal(s *marshal.TagFileSegment) {
	i.BaseContainer.Marshal(s)
	s.PutTag(marshal.TagLayer, byte(i.layer))
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
func (i *WearableContainer) Unmarshal(to *marshal.TagObject) {
	i.BaseContainer.Unmarshal(to)
	i.layer = uo.Layer(to.Tags.Byte(marshal.TagLayer, byte(uo.LayerWeapon)))
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (s *WearableContainer) OnAfterDeserialize(f *util.TagFileObject) {
	s.BaseContainer.OnAfterDeserialize(f)
}

// Layer implements the Layerer interface.
func (c *WearableContainer) Layer() uo.Layer { return c.layer }
