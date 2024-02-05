package uo

// Graphic is a reference number to an item graphic.
type Graphic uint16

// Item number values
const (
	GraphicDefault      Graphic = 0x0000
	GraphicNone         Graphic = 0x0000
	GraphicNoDraw       Graphic = 0x0002
	GraphicCaveEntrance Graphic = 0x01DB
	GraphicNoDrawStart  Graphic = 0x01AE
	GraphicNoDrawEnd    Graphic = 0x01B5
	GraphicHueFlag      Graphic = 0x8000
	GraphicHueMask      Graphic = 0x7fff
)

// UnmarshalJSON implements json.Unmarshaler.
func (i *Graphic) UnmarshalJSON(in []byte) error {
	*i = Graphic(flexNum(in))
	return nil
}

// IsNoDraw tries to determine if this is one of the common no-draw tiles.
// Please note that this is NOT 100% accurate.
func (i Graphic) IsNoDraw() bool {
	return i == GraphicNoDraw ||
		(i >= GraphicNoDrawStart && i <= GraphicNoDrawEnd)
}

// HasHueFlag returns true if the hue flag is present
func (i Graphic) HasHueFlag() bool {
	return i&GraphicHueFlag != 0
}

// SetHueFlag sets the hue flag and returns the new Body value
func (i Graphic) SetHueFlag() Graphic {
	return i | GraphicHueFlag
}

// RemoveHueFlag removes the hue flag and returns the new Body value
func (i Graphic) RemoveHueFlag() Graphic {
	return i & GraphicHueMask
}
