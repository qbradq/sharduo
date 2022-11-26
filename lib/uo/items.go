package uo

// Item is a reference number to an item graphic.
type Item uint16

// Item number values
const (
	ItemDefault Item = 0x0000
	ItemNone    Item = 0x0000
)

var items = map[string]Item{}

// GetItem returns the named item number or the default.
func GetItem(name string) Item {
	if item, ok := items[name]; ok {
		return item
	}
	return ItemDefault
}
