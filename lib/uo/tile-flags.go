package uo

// TileFlags represents the bit field of flags for a tile
// definition.
type TileFlags uint64

// Bit definitions for TileFlags
const (
	TileFlagsNone         TileFlags = 0x000000000000
	TileFlagsBackground   TileFlags = 0x000000000001
	TileFlagsWeapon       TileFlags = 0x000000000002
	TileFlagsTransparent  TileFlags = 0x000000000004
	TileFlagsTranslucent  TileFlags = 0x000000000008
	TileFlagsWall         TileFlags = 0x000000000010
	TileFlagsDamaging     TileFlags = 0x000000000020
	TileFlagsImpassable   TileFlags = 0x000000000040
	TileFlagsWet          TileFlags = 0x000000000008
	TileFlagsUnknown1     TileFlags = 0x000000000100
	TileFlagsSurface      TileFlags = 0x000000000200
	TileFlagsBridge       TileFlags = 0x000000000400
	TileFlagsGeneric      TileFlags = 0x000000000800
	TileFlagsWindow       TileFlags = 0x000000001000
	TileFlagsNoShoot      TileFlags = 0x000000002000
	TileFlagsArticleA     TileFlags = 0x000000004000
	TileFlagsArticleAn    TileFlags = 0x000000008000
	TileFlagsInternal     TileFlags = 0x000000010000
	TileFlagsFoliage      TileFlags = 0x000000020000
	TileFlagsPartialHue   TileFlags = 0x000000040000
	TileFlagsNoHouse      TileFlags = 0x000000080000
	TileFlagsMap          TileFlags = 0x000000100000
	TileFlagsContainer    TileFlags = 0x000000200000
	TileFlagsWearable     TileFlags = 0x000000400000
	TileFlagsLightSource  TileFlags = 0x000000800000
	TileFlagsAnimation    TileFlags = 0x000001000000
	TileFlagsNoDiagonal   TileFlags = 0x000002000000
	TileFlagsUnknown2     TileFlags = 0x000004000000
	TileFlagsArmor        TileFlags = 0x000008000000
	TileFlagsRoof         TileFlags = 0x000010000000
	TileFlagsDoor         TileFlags = 0x000020000000
	TileFlagsStairBack    TileFlags = 0x000040000000
	TileFlagsStairRight   TileFlags = 0x000080000000
	TileFlagsAlphaBlend   TileFlags = 0x000100000000
	TileFlagsUseNewArt    TileFlags = 0x000200000000
	TileFlagsArtUsed      TileFlags = 0x000400000000
	TileFlagsNoShadow     TileFlags = 0x001000000000
	TileFlagsPixelBleed   TileFlags = 0x002000000000
	TileFlagsPlayAnimOnce TileFlags = 0x004000000000
	TileFlagsMultiMovable TileFlags = 0x010000000000
)

// Flag reading functions
func (f TileFlags) Background() bool   { return f&TileFlagsBackground != 0 }
func (f TileFlags) Weapon() bool       { return f&TileFlagsWeapon != 0 }
func (f TileFlags) Transparent() bool  { return f&TileFlagsTransparent != 0 }
func (f TileFlags) Translucent() bool  { return f&TileFlagsTranslucent != 0 }
func (f TileFlags) Wall() bool         { return f&TileFlagsWall != 0 }
func (f TileFlags) Damaging() bool     { return f&TileFlagsDamaging != 0 }
func (f TileFlags) Impassable() bool   { return f&TileFlagsImpassable != 0 }
func (f TileFlags) Wet() bool          { return f&TileFlagsWet != 0 }
func (f TileFlags) Surface() bool      { return f&TileFlagsSurface != 0 }
func (f TileFlags) Bridge() bool       { return f&TileFlagsBridge != 0 }
func (f TileFlags) Generic() bool      { return f&TileFlagsGeneric != 0 }
func (f TileFlags) Window() bool       { return f&TileFlagsWindow != 0 }
func (f TileFlags) NoShoot() bool      { return f&TileFlagsNoShoot != 0 }
func (f TileFlags) ArticleA() bool     { return f&TileFlagsArticleA != 0 }
func (f TileFlags) ArticleAn() bool    { return f&TileFlagsArticleAn != 0 }
func (f TileFlags) Internal() bool     { return f&TileFlagsInternal != 0 }
func (f TileFlags) Foliage() bool      { return f&TileFlagsFoliage != 0 }
func (f TileFlags) PartialHue() bool   { return f&TileFlagsPartialHue != 0 }
func (f TileFlags) NoHouse() bool      { return f&TileFlagsNoHouse != 0 }
func (f TileFlags) Map() bool          { return f&TileFlagsMap != 0 }
func (f TileFlags) Container() bool    { return f&TileFlagsContainer != 0 }
func (f TileFlags) Wearable() bool     { return f&TileFlagsWearable != 0 }
func (f TileFlags) LightSource() bool  { return f&TileFlagsLightSource != 0 }
func (f TileFlags) Animation() bool    { return f&TileFlagsAnimation != 0 }
func (f TileFlags) NoDiagonal() bool   { return f&TileFlagsNoDiagonal != 0 }
func (f TileFlags) Armor() bool        { return f&TileFlagsArmor != 0 }
func (f TileFlags) Roof() bool         { return f&TileFlagsRoof != 0 }
func (f TileFlags) Door() bool         { return f&TileFlagsDoor != 0 }
func (f TileFlags) StairBack() bool    { return f&TileFlagsStairBack != 0 }
func (f TileFlags) StairRight() bool   { return f&TileFlagsStairRight != 0 }
func (f TileFlags) AlphaBlend() bool   { return f&TileFlagsAlphaBlend != 0 }
func (f TileFlags) UseNewArt() bool    { return f&TileFlagsUseNewArt != 0 }
func (f TileFlags) ArtUsed() bool      { return f&TileFlagsArtUsed != 0 }
func (f TileFlags) NoShadow() bool     { return f&TileFlagsBackground != 0 }
func (f TileFlags) PixelBleed() bool   { return f&TileFlagsPixelBleed != 0 }
func (f TileFlags) PlayAnimOnce() bool { return f&TileFlagsPlayAnimOnce != 0 }
func (f TileFlags) MultiMovable() bool { return f&TileFlagsMultiMovable != 0 }
