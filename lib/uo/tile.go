package uo

// CommonObject represents an object that can be queried for it's item graphic,
// it's permanent Z location, and tile flags.
type CommonObject interface {
	// BaseGraphic returns the item graphic of the object
	BaseGraphic() Graphic
	// Z returns the permanent Z location of the object
	Z() int
	// Height returns the height of the object.
	Height() int
	// StandingHeight returns the height at which other objects rest above this
	// object's position. Absolute elevation is Z()+StandingHeight(). For solid
	// objects this is equal to Z()+Height(). For Bridge() type objects -
	// typically stairs - this is Z()+Height()/2 rounded down. For all non-solid
	// objects the return value will be Z().
	StandingHeight() int
	// Flag accessors
	Background() bool
	Weapon() bool
	Transparent() bool
	Translucent() bool
	Wall() bool
	Damaging() bool
	Impassable() bool
	Wet() bool
	Surface() bool
	Bridge() bool
	Generic() bool
	Window() bool
	NoShoot() bool
	ArticleA() bool
	ArticleAn() bool
	Internal() bool
	Foliage() bool
	PartialHue() bool
	NoHouse() bool
	Map() bool
	Container() bool
	Wearable() bool
	LightSource() bool
	Animation() bool
	NoDiagonal() bool
	Armor() bool
	Roof() bool
	Door() bool
	StairBack() bool
	StairRight() bool
	AlphaBlend() bool
	UseNewArt() bool
	ArtUsed() bool
	NoShadow() bool
	PixelBleed() bool
	PlayAnimOnce() bool
	MultiMovable() bool
}

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
	z int
}

// NewTile returns a Tile value with the given properties
func NewTile(z int, def *TileDefinition) Tile {
	return Tile{
		def: def,
		z:   z,
	}
}

// Ignore returns true if this tile should not used when calculating Z locations
func (t Tile) Ignore() bool {
	return t.def.Graphic == GraphicNoDraw ||
		t.def.Graphic == GraphicCaveEntrance ||
		(t.def.Graphic >= GraphicNoDrawStart && t.def.Graphic <= GraphicNoDrawEnd)
}

// BaseGraphic returns the graphic of the tile
func (t Tile) BaseGraphic() Graphic { return t.def.Graphic }

// Z returns the permanent Z location of the tile
func (t Tile) Z() int { return BoundZ(t.z) }

// Height returns the height of the tile, which is always 0
func (t Tile) Height() int { return 0 }

// StandingHeight returns the standing height of the tile, which is always 0
func (t Tile) StandingHeight() int { return 0 }

func (t Tile) Background() bool   { return t.def.TileFlags&TileFlagsBackground != 0 }
func (t Tile) Weapon() bool       { return t.def.TileFlags&TileFlagsWeapon != 0 }
func (t Tile) Transparent() bool  { return t.def.TileFlags&TileFlagsTransparent != 0 }
func (t Tile) Translucent() bool  { return t.def.TileFlags&TileFlagsTranslucent != 0 }
func (t Tile) Wall() bool         { return t.def.TileFlags&TileFlagsWall != 0 }
func (t Tile) Damaging() bool     { return t.def.TileFlags&TileFlagsDamaging != 0 }
func (t Tile) Impassable() bool   { return t.def.TileFlags&TileFlagsImpassable != 0 }
func (t Tile) Wet() bool          { return t.def.TileFlags&TileFlagsWet != 0 }
func (t Tile) Surface() bool      { return t.def.TileFlags&TileFlagsSurface != 0 }
func (t Tile) Bridge() bool       { return t.def.TileFlags&TileFlagsBridge != 0 }
func (t Tile) Generic() bool      { return t.def.TileFlags&TileFlagsGeneric != 0 }
func (t Tile) Window() bool       { return t.def.TileFlags&TileFlagsWindow != 0 }
func (t Tile) NoShoot() bool      { return t.def.TileFlags&TileFlagsNoShoot != 0 }
func (t Tile) ArticleA() bool     { return t.def.TileFlags&TileFlagsArticleA != 0 }
func (t Tile) ArticleAn() bool    { return t.def.TileFlags&TileFlagsArticleAn != 0 }
func (t Tile) Internal() bool     { return t.def.TileFlags&TileFlagsInternal != 0 }
func (t Tile) Foliage() bool      { return t.def.TileFlags&TileFlagsFoliage != 0 }
func (t Tile) PartialHue() bool   { return t.def.TileFlags&TileFlagsPartialHue != 0 }
func (t Tile) NoHouse() bool      { return t.def.TileFlags&TileFlagsNoHouse != 0 }
func (t Tile) Map() bool          { return t.def.TileFlags&TileFlagsMap != 0 }
func (t Tile) Container() bool    { return t.def.TileFlags&TileFlagsContainer != 0 }
func (t Tile) Wearable() bool     { return t.def.TileFlags&TileFlagsWearable != 0 }
func (t Tile) LightSource() bool  { return t.def.TileFlags&TileFlagsLightSource != 0 }
func (t Tile) Animation() bool    { return t.def.TileFlags&TileFlagsAnimation != 0 }
func (t Tile) NoDiagonal() bool   { return t.def.TileFlags&TileFlagsNoDiagonal != 0 }
func (t Tile) Armor() bool        { return t.def.TileFlags&TileFlagsArmor != 0 }
func (t Tile) Roof() bool         { return t.def.TileFlags&TileFlagsRoof != 0 }
func (t Tile) Door() bool         { return t.def.TileFlags&TileFlagsDoor != 0 }
func (t Tile) StairBack() bool    { return t.def.TileFlags&TileFlagsStairBack != 0 }
func (t Tile) StairRight() bool   { return t.def.TileFlags&TileFlagsStairRight != 0 }
func (t Tile) AlphaBlend() bool   { return t.def.TileFlags&TileFlagsAlphaBlend != 0 }
func (t Tile) UseNewArt() bool    { return t.def.TileFlags&TileFlagsUseNewArt != 0 }
func (t Tile) ArtUsed() bool      { return t.def.TileFlags&TileFlagsArtUsed != 0 }
func (t Tile) NoShadow() bool     { return t.def.TileFlags&TileFlagsBackground != 0 }
func (t Tile) PixelBleed() bool   { return t.def.TileFlags&TileFlagsPixelBleed != 0 }
func (t Tile) PlayAnimOnce() bool { return t.def.TileFlags&TileFlagsPlayAnimOnce != 0 }
func (t Tile) MultiMovable() bool { return t.def.TileFlags&TileFlagsMultiMovable != 0 }
