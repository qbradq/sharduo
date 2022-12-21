package uo

// StaticDefinition holds the static data of a static object
type StaticDefinition struct {
	// Graphic of the item
	Graphic Graphic
	// Flags
	TileFlags TileFlags
	// Weight
	Weight int
	// Layer
	Layer Layer
	// Stack amount
	Count int
	// Animation of the static
	Animation Animation
	// Hue
	Hue Hue
	// Light index
	Light Light
	// Z position
	Z int
	// Name of the static, truncated to 20 characters
	Name string
}

// Static represents a static object on the map
type Static struct {
	// Static definition
	def *StaticDefinition
	// Absolute location
	Location Location
}

// NewStatic returns a new Static object with the given definition and location
func NewStatic(l Location, def *StaticDefinition) Static {
	return Static{
		def:      def,
		Location: l,
	}
}

// Graphic returns the graphic of the static
func (s Static) Graphic() Graphic {
	return s.def.Graphic
}
