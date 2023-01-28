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
	// Height
	Height int8
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

// BaseGraphic returns the graphic of the static
func (s Static) BaseGraphic() Graphic {
	return s.def.Graphic
}

// Height returns the height of the static with regards to internal flags
func (s Static) Height() int8 {
	if s.def.TileFlags.Bridge() {
		return s.def.Height / 2
	}
	return s.def.Height
}

// StandingHeight returns the standing height based on the object's flags.
func (s Static) StandingHeight() int8 {
	if !s.Surface() && !s.Wet() && !s.Impassable() {
		return 0
	}
	if s.Bridge() {
		return s.def.Height / 2
	}
	return s.def.Height
}

// Z returns the permanent Z location of the tile
func (s Static) Z() int8 { return s.Location.Z }

func (s Static) Background() bool   { return s.def.TileFlags&TileFlagsBackground != 0 }
func (s Static) Weapon() bool       { return s.def.TileFlags&TileFlagsWeapon != 0 }
func (s Static) Transparent() bool  { return s.def.TileFlags&TileFlagsTransparent != 0 }
func (s Static) Translucent() bool  { return s.def.TileFlags&TileFlagsTranslucent != 0 }
func (s Static) Wall() bool         { return s.def.TileFlags&TileFlagsWall != 0 }
func (s Static) Damaging() bool     { return s.def.TileFlags&TileFlagsDamaging != 0 }
func (s Static) Impassable() bool   { return s.def.TileFlags&TileFlagsImpassable != 0 }
func (s Static) Wet() bool          { return s.def.TileFlags&TileFlagsWet != 0 }
func (s Static) Surface() bool      { return s.def.TileFlags&TileFlagsSurface != 0 }
func (s Static) Bridge() bool       { return s.def.TileFlags&TileFlagsBridge != 0 }
func (s Static) Generic() bool      { return s.def.TileFlags&TileFlagsGeneric != 0 }
func (s Static) Window() bool       { return s.def.TileFlags&TileFlagsWindow != 0 }
func (s Static) NoShoot() bool      { return s.def.TileFlags&TileFlagsNoShoot != 0 }
func (s Static) ArticleA() bool     { return s.def.TileFlags&TileFlagsArticleA != 0 }
func (s Static) ArticleAn() bool    { return s.def.TileFlags&TileFlagsArticleAn != 0 }
func (s Static) Internal() bool     { return s.def.TileFlags&TileFlagsInternal != 0 }
func (s Static) Foliage() bool      { return s.def.TileFlags&TileFlagsFoliage != 0 }
func (s Static) PartialHue() bool   { return s.def.TileFlags&TileFlagsPartialHue != 0 }
func (s Static) NoHouse() bool      { return s.def.TileFlags&TileFlagsNoHouse != 0 }
func (s Static) Map() bool          { return s.def.TileFlags&TileFlagsMap != 0 }
func (s Static) Container() bool    { return s.def.TileFlags&TileFlagsContainer != 0 }
func (s Static) Wearable() bool     { return s.def.TileFlags&TileFlagsWearable != 0 }
func (s Static) LightSource() bool  { return s.def.TileFlags&TileFlagsLightSource != 0 }
func (s Static) Animation() bool    { return s.def.TileFlags&TileFlagsAnimation != 0 }
func (s Static) NoDiagonal() bool   { return s.def.TileFlags&TileFlagsNoDiagonal != 0 }
func (s Static) Armor() bool        { return s.def.TileFlags&TileFlagsArmor != 0 }
func (s Static) Roof() bool         { return s.def.TileFlags&TileFlagsRoof != 0 }
func (s Static) Door() bool         { return s.def.TileFlags&TileFlagsDoor != 0 }
func (s Static) StairBack() bool    { return s.def.TileFlags&TileFlagsStairBack != 0 }
func (s Static) StairRight() bool   { return s.def.TileFlags&TileFlagsStairRight != 0 }
func (s Static) AlphaBlend() bool   { return s.def.TileFlags&TileFlagsAlphaBlend != 0 }
func (s Static) UseNewArt() bool    { return s.def.TileFlags&TileFlagsUseNewArt != 0 }
func (s Static) ArtUsed() bool      { return s.def.TileFlags&TileFlagsArtUsed != 0 }
func (s Static) NoShadow() bool     { return s.def.TileFlags&TileFlagsBackground != 0 }
func (s Static) PixelBleed() bool   { return s.def.TileFlags&TileFlagsPixelBleed != 0 }
func (s Static) PlayAnimOnce() bool { return s.def.TileFlags&TileFlagsPlayAnimOnce != 0 }
func (s Static) MultiMovable() bool { return s.def.TileFlags&TileFlagsMultiMovable != 0 }
