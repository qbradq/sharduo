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
	Animation AnimationType
	// Hue
	Hue Hue
	// Light index
	Light Light
	// Height
	Height int
	// Name of the static, truncated to 20 characters
	Name string
}

// Static represents a static object on the map
type Static struct {
	// Static definition
	Def *StaticDefinition
	// Absolute location
	Location Point
}

// NewStatic returns a new Static object with the given definition and location
func NewStatic(l Point, def *StaticDefinition) Static {
	return Static{
		Def:      def,
		Location: l,
	}
}

// BaseGraphic returns the graphic of the static
func (s Static) BaseGraphic() Graphic {
	return s.Def.Graphic
}

// Height returns the height of the static with regards to internal flags
func (s Static) Height() int {
	return s.Def.Height
}

// Highest returns the highest elevation of the object
func (s Static) Highest() int {
	return s.Location.Z + s.Def.Height
}

// StandingHeight returns the standing height based on the object's flags.
func (s Static) StandingHeight() int {
	if !s.Surface() && !s.Wet() && !s.Impassable() {
		return s.Location.Z
	}
	if s.Bridge() {
		return s.Location.Z + s.Def.Height/2
	}
	return s.Location.Z + s.Def.Height
}

// Z returns the permanent Z location of the tile
func (s Static) Z() int { return s.Location.Z }

func (s Static) Background() bool      { return s.Def.TileFlags&TileFlagsBackground != 0 }
func (s Static) Weapon() bool          { return s.Def.TileFlags&TileFlagsWeapon != 0 }
func (s Static) Transparent() bool     { return s.Def.TileFlags&TileFlagsTransparent != 0 }
func (s Static) Translucent() bool     { return s.Def.TileFlags&TileFlagsTranslucent != 0 }
func (s Static) Wall() bool            { return s.Def.TileFlags&TileFlagsWall != 0 }
func (s Static) Damaging() bool        { return s.Def.TileFlags&TileFlagsDamaging != 0 }
func (s Static) Impassable() bool      { return s.Def.TileFlags&TileFlagsImpassable != 0 }
func (s Static) Wet() bool             { return s.Def.TileFlags&TileFlagsWet != 0 }
func (s Static) Surface() bool         { return s.Def.TileFlags&TileFlagsSurface != 0 }
func (s Static) Bridge() bool          { return s.Def.TileFlags&TileFlagsBridge != 0 }
func (s Static) Generic() bool         { return s.Def.TileFlags&TileFlagsGeneric != 0 }
func (s Static) Window() bool          { return s.Def.TileFlags&TileFlagsWindow != 0 }
func (s Static) NoShoot() bool         { return s.Def.TileFlags&TileFlagsNoShoot != 0 }
func (s Static) ArticleA() bool        { return s.Def.TileFlags&TileFlagsArticleA != 0 }
func (s Static) ArticleAn() bool       { return s.Def.TileFlags&TileFlagsArticleAn != 0 }
func (s Static) Internal() bool        { return s.Def.TileFlags&TileFlagsInternal != 0 }
func (s Static) Foliage() bool         { return s.Def.TileFlags&TileFlagsFoliage != 0 }
func (s Static) PartialHue() bool      { return s.Def.TileFlags&TileFlagsPartialHue != 0 }
func (s Static) NoHouse() bool         { return s.Def.TileFlags&TileFlagsNoHouse != 0 }
func (s Static) Map() bool             { return s.Def.TileFlags&TileFlagsMap != 0 }
func (s Static) StaticContainer() bool { return s.Def.TileFlags&TileFlagsContainer != 0 }
func (s Static) Wearable() bool        { return s.Def.TileFlags&TileFlagsWearable != 0 }
func (s Static) LightSource() bool     { return s.Def.TileFlags&TileFlagsLightSource != 0 }
func (s Static) Animation() bool       { return s.Def.TileFlags&TileFlagsAnimation != 0 }
func (s Static) NoDiagonal() bool      { return s.Def.TileFlags&TileFlagsNoDiagonal != 0 }
func (s Static) Armor() bool           { return s.Def.TileFlags&TileFlagsArmor != 0 }
func (s Static) Roof() bool            { return s.Def.TileFlags&TileFlagsRoof != 0 }
func (s Static) Door() bool            { return s.Def.TileFlags&TileFlagsDoor != 0 }
func (s Static) StairBack() bool       { return s.Def.TileFlags&TileFlagsStairBack != 0 }
func (s Static) StairRight() bool      { return s.Def.TileFlags&TileFlagsStairRight != 0 }
func (s Static) AlphaBlend() bool      { return s.Def.TileFlags&TileFlagsAlphaBlend != 0 }
func (s Static) UseNewArt() bool       { return s.Def.TileFlags&TileFlagsUseNewArt != 0 }
func (s Static) ArtUsed() bool         { return s.Def.TileFlags&TileFlagsArtUsed != 0 }
func (s Static) NoShadow() bool        { return s.Def.TileFlags&TileFlagsBackground != 0 }
func (s Static) PixelBleed() bool      { return s.Def.TileFlags&TileFlagsPixelBleed != 0 }
func (s Static) PlayAnimOnce() bool    { return s.Def.TileFlags&TileFlagsPlayAnimOnce != 0 }
func (s Static) MultiMovable() bool    { return s.Def.TileFlags&TileFlagsMultiMovable != 0 }
