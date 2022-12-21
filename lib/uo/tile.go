package uo

// TileDefinition holds the static data about a tile. Valid tile indexes are
// 0x0000-0x3FFF.
type TileDefinition struct {
	// Tile graphic
	Graphic Graphic
	// Tile flags
	TileFlags TileFlags
	// Texture for the tile, if any
	Texture Texture
	// Name of the tile truncated to 20 characters
	Name string
}

// Tile represents one landscape tile on the map
type Tile struct {
	// Pointer to the tile definition for this tile
	def *TileDefinition
	// Altitude of the tile
	Z int
}

// NewTile returns a Tile value with the given properties
func NewTile(z int, def *TileDefinition) Tile {
	return Tile{
		def: def,
		Z:   z,
	}
}

// Graphic returns the graphic of the tile
func (t Tile) Graphic() Graphic {
	return t.def.Graphic
}
