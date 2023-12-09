package game

import (
	"strconv"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/template"
)

func init() {
	reg("Check", marshal.ObjectTypeCheck, func() any { return &Check{} })
}

// Check implements a bank check
type Check struct {
	BaseItem
	checkAmount int // The amount of gold left on the check.
}

// ObjectType implements marshal.Marshaler.
func (i *Check) ObjectType() marshal.ObjectType { return marshal.ObjectTypeCheck }

// Marshal implements the marshal.Marshaler interface.
func (i *Check) Marshal(s *marshal.TagFileSegment) {
	i.BaseItem.Marshal(s)
	s.PutInt(0)                     // version
	s.PutInt(uint32(i.checkAmount)) // check amount
}

// Deserialize implements the util.Serializeable interface.
func (i *Check) Deserialize(t *template.Template, create bool) {
	i.BaseItem.Deserialize(t, create)
	i.checkAmount = i.amount
	i.amount = 1
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *Check) Unmarshal(s *marshal.TagFileSegment) {
	i.BaseItem.Unmarshal(s)
	_ = s.Int() // version
	i.checkAmount = int(s.Int())
}

// DisplayName implements the Object interface
func (i *Check) DisplayName() string {
	return "a check for " + strconv.FormatInt(int64(i.checkAmount), 10) + " gold"
}

// CheckAmount returns the amount of gold left on the check.
func (i *Check) CheckAmount() int { return i.checkAmount }

// SetCheckAmount sets the amount on the check.
func (i *Check) SetCheckAmount(v int) { i.checkAmount = v }

// ConsumeGold removes the given amount of gold from the check if possible and
// returns true if completed successfully.
func (i *Check) ConsumeGold(v int) bool {
	if v > i.checkAmount {
		// Requested amount exceeds check amount
		return false
	}
	i.checkAmount -= v
	if i.checkAmount < 1 {
		Remove(i)
	} else {
		i.InvalidateOPL()
	}
	return true
}
