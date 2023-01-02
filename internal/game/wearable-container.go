package game

import (
	"log"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &WearableContainer{} })
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

// Serialize implements the util.Serializeable interface.
func (s *WearableContainer) Serialize(f *util.TagFileWriter) {
	s.BaseContainer.Serialize(f)
	f.WriteNumber("Layer", int(s.layer))
}

// Deserialize implements the util.Serializeable interface.
func (s *WearableContainer) Deserialize(f *util.TagFileObject) {
	s.BaseContainer.Deserialize(f)
	s.layer = uo.Layer(f.GetNumber("Layer", int(uo.LayerInvalid)))
	if s.layer == uo.LayerInvalid {
		log.Printf("error: wearable container %s no layer given", s.Serial().String())
	}
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (s *WearableContainer) OnAfterDeserialize(f *util.TagFileObject) {
	s.BaseContainer.OnAfterDeserialize(f)
}

// Layer implements the Layerer interface.
func (c *WearableContainer) Layer() uo.Layer { return c.layer }
