package uo

// Item is a reference number to an item graphic.
type Item uint16

// Item number values
const (
	ItemDefault Item = 0x0000
	ItemNone    Item = 0x0000
	ItemHueFlag Item = 0x8000
	ItemHueMask Item = 0x7fff
)

var items = map[string]Item{}

// GetItem returns the named item number or the default.
func GetItem(name string) Item {
	if item, ok := items[name]; ok {
		return item
	}
	return ItemDefault
}

// HasHueFlag returns true if the hue flag is present
func (i Item) HasHueFlag() bool {
	return i&ItemHueFlag != 0
}

// SetHueFlag sets the hue flag and returns the new Body value
func (i Item) SetHueFlag() Item {
	return i | ItemHueFlag
}

// RemoveHueFlag removes the hue flag and returns the new Body value
func (i Item) RemoveHueFlag() Item {
	return i & ItemHueMask
}
