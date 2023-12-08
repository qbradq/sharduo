package game

import (
	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/template"
)

func init() {
	reg("WearableContainer", marshal.ObjectTypeWearableContainer, func() any { return &WearableContainer{} })
}

// WearableContainer is a wearable item with the properties of a container, such
// as inventory backpacks and the player's bank box.
type WearableContainer struct {
	BaseContainer
	BaseWearableImplementation
}

// ObjectType implements the Object interface.
func (i *WearableContainer) ObjectType() marshal.ObjectType {
	return marshal.ObjectTypeWearableContainer
}

// Deserialize implements the util.Serializeable interface.
func (i *WearableContainer) Deserialize(t *template.Template, create bool) {
	i.BaseContainer.Deserialize(t, create)
	i.BaseWearableImplementation.Deserialize(t, create)
}

// Marshal implements the marshal.Marshaler interface.
func (i *WearableContainer) Marshal(s *marshal.TagFileSegment) {
	i.BaseContainer.Marshal(s)
	i.BaseWearableImplementation.Marshal(s)
	s.PutInt(0) // version
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *WearableContainer) Unmarshal(s *marshal.TagFileSegment) {
	i.BaseContainer.Unmarshal(s)
	i.BaseWearableImplementation.Unmarshal(s)
	_ = s.Int() // version
}
