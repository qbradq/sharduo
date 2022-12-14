package uo

// Graphic is a reference number to an item graphic.
type Graphic uint16

// Item number values
const (
	ItemDefault Graphic = 0x0000
	ItemNone    Graphic = 0x0000
	ItemHueFlag Graphic = 0x8000
	ItemHueMask Graphic = 0x7fff
)

// HasHueFlag returns true if the hue flag is present
func (i Graphic) HasHueFlag() bool {
	return i&ItemHueFlag != 0
}

// SetHueFlag sets the hue flag and returns the new Body value
func (i Graphic) SetHueFlag() Graphic {
	return i | ItemHueFlag
}

// RemoveHueFlag removes the hue flag and returns the new Body value
func (i Graphic) RemoveHueFlag() Graphic {
	return i & ItemHueMask
}
