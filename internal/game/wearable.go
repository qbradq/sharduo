package game

import (
	"fmt"
	"image/color"
	"log"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("BaseWearable", marshal.ObjectTypeWearable, func() any { return &BaseWearable{} })
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
	// DamageDurability handles durability loss of a wearable. The first
	// parameter must be an object back-reference.
	DamageDurability(Object, float64)
}

// BaseWearableImplementation provides the most common implementation of the
// Wearable interface and associated functionality to be mixed into other
// structs.
type BaseWearableImplementation struct {
	// Layer this wearable equips to
	layer uo.Layer
	// Current durability
	durability float64
	// Maximum durability
	maxDurability float64
}

// Deserialize implements the util.Serializeable interface.
func (i *BaseWearableImplementation) Deserialize(t *template.Template, create bool) {
	i.layer = uo.Layer(t.GetNumber("Layer", int(uo.LayerInvalid)))
	if i.layer == uo.LayerInvalid {
		log.Fatalf("wearable template %s no layer given", t.TemplateName)
	}
	i.durability = float64(t.GetFloat("Durability", 10))
	i.maxDurability = i.durability
}

// Layer implements the Wearable interface.
func (i *BaseWearableImplementation) Layer() uo.Layer { return i.layer }

// Marshal implements the marshal.Marshaler interface.
func (i *BaseWearableImplementation) Marshal(s *marshal.TagFileSegment) {
	s.PutFloat(i.durability)
	s.PutFloat(i.maxDurability)
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *BaseWearableImplementation) Unmarshal(s *marshal.TagFileSegment) {
	i.durability = s.Float()
	i.maxDurability = s.Float()
}

// DamageDurability implements the Wearable interface.
func (i *BaseWearableImplementation) DamageDurability(r Object, v float64) {
	i.durability -= v
	if i.durability < 0 {
		i.maxDurability += i.durability
		i.durability = 0
	}
	if i.maxDurability <= 0 {
		Remove(r)
	} else {
		r.InvalidateOPL()
	}
}

// BaseWearable provides the most common implementation of Wearable
type BaseWearable struct {
	BaseItem
	BaseWearableImplementation
}

// ObjectType implements the Object interface.
func (i *BaseWearable) ObjectType() marshal.ObjectType { return marshal.ObjectTypeWearable }

// Deserialize implements the util.Serializeable interface.
func (i *BaseWearable) Deserialize(t *template.Template, create bool) {
	i.BaseItem.Deserialize(t, create)
	i.BaseWearableImplementation.Deserialize(t, create)
}

// Marshal implements the marshal.Marshaler interface.
func (i *BaseWearable) Marshal(s *marshal.TagFileSegment) {
	i.BaseItem.Marshal(s)
	i.BaseWearableImplementation.Marshal(s)
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *BaseWearable) Unmarshal(s *marshal.TagFileSegment) {
	i.BaseItem.Unmarshal(s)
	i.BaseWearableImplementation.Unmarshal(s)
}

// AppendOPLEntries implements the Object interface.
func (i *BaseWearable) AppendOPLEntires(r Object, p *serverpacket.OPLPacket) {
	i.BaseItem.AppendOPLEntires(r, p)
	p.AppendColor(color.RGBA{
		R: 127,
		G: 127,
		B: 127,
		A: 255,
	}, fmt.Sprintf("Durability %d/%d",
		int(i.durability),
		int(i.maxDurability),
	), true)
}
