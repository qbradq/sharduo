package uo

// CommonObject represents an object that can be queried for it's item graphic,
// it's permanent Z location, and tile flags.
type CommonObject interface {
	// BaseGraphic returns the item graphic of the object
	BaseGraphic() Graphic
	// Z returns the permanent Z location of the lowest point of the object
	Z() int
	// Height returns the height of the object.
	Height() int
	// Highest returns the highest elevation of the object.
	Highest() int
	// StandingHeight returns the height at which other objects rest above this
	// object's position. For solid objects this is equal to Z()+Height(). For
	// Bridge() type objects - typically stairs - this is Z()+Height()/2 rounded
	// down. For all non-solid objects the return value will be Z().
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
	StaticContainer() bool
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
	Def *TileDefinition
	// Altitude of the tile's North West corner
	NWZ int
	// Altitude of the tile's LowestZ point
	LowestZ int
	// Height of the tile
	HighestZ int
	// Height of the standing point
	AvgZ int
}

// NewTile returns a Tile value with the given properties
func NewTile(z int, def *TileDefinition) Tile {
	return Tile{
		Def: def,
		NWZ: z,
	}
}

// Ignore returns true if this tile should not used when calculating Z locations
func (t Tile) Ignore() bool {
	return t.Def.Graphic == GraphicNoDraw ||
		t.Def.Graphic == GraphicCaveEntrance ||
		(t.Def.Graphic >= GraphicNoDrawStart && t.Def.Graphic <= GraphicNoDrawEnd)
}

// BaseGraphic returns the graphic of the tile
func (t Tile) BaseGraphic() Graphic { return t.Def.Graphic }

// RawZ returns the elevation of the tile from map0.mul
func (t Tile) RawZ() int { return t.NWZ }

// Z returns the elevation of the lowest corder of the tile
func (t Tile) Z() int { return t.LowestZ }

// Height returns the height of the tile
func (t Tile) Height() int { return t.HighestZ - t.LowestZ }

// Highest returns the highest point of the tile
func (t Tile) Highest() int { return t.HighestZ }

// StandingHeight returns the standing height of the tile, which is always 0
func (t Tile) StandingHeight() int { return t.AvgZ }

// SetElevations sets the three pre-calculated elevation parameters
func (t Tile) SetElevations(lowest, avg, height int) Tile {
	t.LowestZ = lowest
	t.AvgZ = avg
	t.HighestZ = height
	return t
}

func (t Tile) Background() bool      { return t.Def.TileFlags&TileFlagsBackground != 0 }
func (t Tile) Weapon() bool          { return t.Def.TileFlags&TileFlagsWeapon != 0 }
func (t Tile) Transparent() bool     { return t.Def.TileFlags&TileFlagsTransparent != 0 }
func (t Tile) Translucent() bool     { return t.Def.TileFlags&TileFlagsTranslucent != 0 }
func (t Tile) Wall() bool            { return t.Def.TileFlags&TileFlagsWall != 0 }
func (t Tile) Damaging() bool        { return t.Def.TileFlags&TileFlagsDamaging != 0 }
func (t Tile) Impassable() bool      { return t.Def.TileFlags&TileFlagsImpassable != 0 }
func (t Tile) Wet() bool             { return t.Def.TileFlags&TileFlagsWet != 0 }
func (t Tile) Surface() bool         { return !t.Impassable() }
func (t Tile) Bridge() bool          { return t.Def.TileFlags&TileFlagsBridge != 0 }
func (t Tile) Generic() bool         { return t.Def.TileFlags&TileFlagsGeneric != 0 }
func (t Tile) Window() bool          { return t.Def.TileFlags&TileFlagsWindow != 0 }
func (t Tile) NoShoot() bool         { return t.Def.TileFlags&TileFlagsNoShoot != 0 }
func (t Tile) ArticleA() bool        { return t.Def.TileFlags&TileFlagsArticleA != 0 }
func (t Tile) ArticleAn() bool       { return t.Def.TileFlags&TileFlagsArticleAn != 0 }
func (t Tile) Internal() bool        { return t.Def.TileFlags&TileFlagsInternal != 0 }
func (t Tile) Foliage() bool         { return t.Def.TileFlags&TileFlagsFoliage != 0 }
func (t Tile) PartialHue() bool      { return t.Def.TileFlags&TileFlagsPartialHue != 0 }
func (t Tile) NoHouse() bool         { return t.Def.TileFlags&TileFlagsNoHouse != 0 }
func (t Tile) Map() bool             { return t.Def.TileFlags&TileFlagsMap != 0 }
func (t Tile) StaticContainer() bool { return t.Def.TileFlags&TileFlagsContainer != 0 }
func (t Tile) Wearable() bool        { return t.Def.TileFlags&TileFlagsWearable != 0 }
func (t Tile) LightSource() bool     { return t.Def.TileFlags&TileFlagsLightSource != 0 }
func (t Tile) Animation() bool       { return t.Def.TileFlags&TileFlagsAnimation != 0 }
func (t Tile) NoDiagonal() bool      { return t.Def.TileFlags&TileFlagsNoDiagonal != 0 }
func (t Tile) Armor() bool           { return t.Def.TileFlags&TileFlagsArmor != 0 }
func (t Tile) Roof() bool            { return t.Def.TileFlags&TileFlagsRoof != 0 }
func (t Tile) Door() bool            { return t.Def.TileFlags&TileFlagsDoor != 0 }
func (t Tile) StairBack() bool       { return t.Def.TileFlags&TileFlagsStairBack != 0 }
func (t Tile) StairRight() bool      { return t.Def.TileFlags&TileFlagsStairRight != 0 }
func (t Tile) AlphaBlend() bool      { return t.Def.TileFlags&TileFlagsAlphaBlend != 0 }
func (t Tile) UseNewArt() bool       { return t.Def.TileFlags&TileFlagsUseNewArt != 0 }
func (t Tile) ArtUsed() bool         { return t.Def.TileFlags&TileFlagsArtUsed != 0 }
func (t Tile) NoShadow() bool        { return t.Def.TileFlags&TileFlagsBackground != 0 }
func (t Tile) PixelBleed() bool      { return t.Def.TileFlags&TileFlagsPixelBleed != 0 }
func (t Tile) PlayAnimOnce() bool    { return t.Def.TileFlags&TileFlagsPlayAnimOnce != 0 }
func (t Tile) MultiMovable() bool    { return t.Def.TileFlags&TileFlagsMultiMovable != 0 }
