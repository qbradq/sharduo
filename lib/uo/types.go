package uo

// Random constants
const (
	MinStackAmount    uint16 = 1
	MaxStackAmount    uint16 = 60000
	MinViewRange      int    = 5
	MaxViewRange      int    = 18
	ChunkWidth        int    = 8
	ChunkHeight       int    = 8
	MapWidth          int    = 6144
	MapHeight         int    = 4096
	MapOverworldWidth int    = MapHeight
	MapChunksWidth    int    = MapWidth / ChunkWidth
	MapChunksHeight   int    = MapHeight / ChunkHeight
	MapMinZ           int    = -127
	MapMaxZ           int    = 128
	StatsCapDefault   int    = 225
	MaxFollowers      int    = 5
)

// BoundViewRange bounds a view range value
func BoundViewRange(r int) int {
	if r < MinViewRange {
		return MinViewRange
	} else if r > MaxViewRange {
		return MaxViewRange
	}
	return r
}

// A Layer is a 6-bit value that indicates at which rendering layer a given
// animation body is rendered. It is also used in paper doll gumps to layer the
// equipment gump images. The zero value is invalid.
type Layer byte

// Layer constants
const (
	LayerInvalid                  Layer = 0
	LayerWeapon                   Layer = 1
	LayerShield                   Layer = 2
	LayerShoes                    Layer = 3
	LayerPants                    Layer = 4
	LayerShirt                    Layer = 5
	LayerHat                      Layer = 6
	LayerGloves                   Layer = 7
	LayerRing                     Layer = 8
	LayerNeckArmor                Layer = 10
	LayerHair                     Layer = 11
	LayerBelt                     Layer = 12
	LayerChestArmor               Layer = 13
	LayerBracelet                 Layer = 14
	LayerBeard                    Layer = 16
	LayerCoat                     Layer = 17
	LayerEarrings                 Layer = 18
	LayerArmArmor                 Layer = 19
	LayerCloak                    Layer = 20
	LayerBackpack                 Layer = 21
	LayerRobe                     Layer = 22
	LayerSkirt                    Layer = 23
	LayerLegArmor                 Layer = 24
	LayerMount                    Layer = 25
	LayerNPCBuyRestockContainer   Layer = 26
	LayerNPCBuyNoRestockContainer Layer = 27
	LayerNPCSellContainer         Layer = 28
	LayerBankBox                  Layer = 29
	LayerFirstValid               Layer = LayerWeapon
	LayerLastValid                Layer = LayerBankBox
)

// A StatusFlag describes the status of a mobile
type StatusFlag byte

// StatusFlag constants
const (
	StatusNormal StatusFlag = 0
)

// A Noto is a 3-bit value describing the notoriety status of a mobile
// The zero-value is invalid
type Notoriety byte

// Notoriety constants
const (
	NotorietyInvalid      Notoriety = 0
	NotorietyInnocent     Notoriety = 1
	NotorietyFriend       Notoriety = 2
	NotorietyAttackable   Notoriety = 3
	NotorietyCriminal     Notoriety = 4
	NotorietyEnemy        Notoriety = 5
	NotorietyMurderer     Notoriety = 6
	NotorietyInvulnerable Notoriety = 7
)

// FeatureFlag represents the client features flags sent in packet 0xA9
type FeatureFlag uint32

// All documented flags
const (
	FeatureFlagNone                 FeatureFlag = 0x00000000
	FeatureFlagSiege                FeatureFlag = 0x00000004
	FeatureFlagLeftClickMenus       FeatureFlag = 0x00000008
	FeatureFlagAOS                  FeatureFlag = 0x00000020
	FeatureFlagSixthCharacterSlot   FeatureFlag = 0x00000040
	FeatureFlagAOSProfessions       FeatureFlag = 0x00000080
	FeatureFlagElvenRace            FeatureFlag = 0x00000100
	FeatureFlagSeventhCharacterSlot FeatureFlag = 0x00001000
	FeatureFlagNewMovementPackets   FeatureFlag = 0x00004000
	FeatureFlagNewFeluccaAreas      FeatureFlag = 0x00008000
)

// LoginDeniedReason represents the reason for refusing login
type LoginDeniedReason byte

// All LoginDeniedReason values
const (
	LoginDeniedReasonBadPass        LoginDeniedReason = 0 // Password invalid for user
	LoginDeniedReasonAccountInUse   LoginDeniedReason = 1 // The account already has an active season
	LoginDeniedReasonAccountBlocked LoginDeniedReason = 2 // The account has been blocked for some reason
)

// SpeechType represents the type of speech being requested or sent.
type SpeechType byte

// All SpeechType values
const (
	SpeechTypeNormal       SpeechType = 0    // Overhead speech
	SpeechTypeBroadcast    SpeechType = 1    // System broadcast
	SpeechTypeEmote        SpeechType = 2    // Overhead emote
	SpeechTypeSystem       SpeechType = 6    // System message in corner
	SpeechTypeMessage      SpeechType = 7    // Message in corner with name
	SpeechTypeWhisper      SpeechType = 8    // Overhead whisper
	SpeechTypeYell         SpeechType = 9    // Overhead yell
	SpeechTypeSpell        SpeechType = 10   // Overhead spell words
	SpeechTypeGuild        SpeechType = 13   // Guild chat in corner
	SpeechTypeAlliance     SpeechType = 14   // Guild alliance chat in corner
	SpeechTypePrompt       SpeechType = 15   // Prompt for user input
	SpeechTypeClientParsed SpeechType = 0xc0 // Contains client-side parsed keywords
)

// Font represents one of the built-in fonts in the client.
type Font uint16

// All Font values
const (
	FontBold         Font = 0
	FontShadow       Font = 1
	FontBoldShadow   Font = 2
	FontNormal       Font = 3
	FontGothic       Font = 4
	FontScript       Font = 5
	FontSmallScript  Font = 6
	FontScriptShadow Font = 7
	FontRune         Font = 8
	FontSmallNormal  Font = 9
)

// StatusRequestType represents the types of status requests
type StatusRequestType byte

// All StatusRequestType values
const (
	StatusRequestTypeGod    StatusRequestType = 0
	StatusRequestTypeBasic  StatusRequestType = 4
	StatusRequestTypeSkills StatusRequestType = 5
)

// MobileFlags represent the flags for a mobile
type MobileFlags uint8

// All MobileFlags flag values
const (
	MobileFlagNone          MobileFlags = 0x00
	MobileFlagFrozen        MobileFlags = 0x01
	MobileFlagFemale        MobileFlags = 0x02
	MobileFlagPoisoned      MobileFlags = 0x04
	MobileFlagFlying        MobileFlags = 0x04
	MobileFlagYellowBar     MobileFlags = 0x08
	MobileFlagIgnoreMobiles MobileFlags = 0x10
	MobileFlagMovable       MobileFlags = 0x20
	MobileFlagWarMode       MobileFlags = 0x40
	MobileFlagHidden        MobileFlags = 0x80
)

// TargetType describes which type of targeting to use
type TargetType uint8

// All TargetType values
const (
	TargetTypeObject   TargetType = 0
	TargetTypeLocation TargetType = 1
)

// CursorType describes what the cursor should look like
type CursorType uint8

// All CursorType values
const (
	CursorTypeNeutral CursorType = 0
	CursorTypeHarmful CursorType = 1
	CursorTypeHelpful CursorType = 2
	CursorTypeCancel  CursorType = 3
)

// A Body is a 16-bit value that describes the set of animations to use for a
// mobile. Body values used by UO range 1-999.
type Body uint16

// Pre-defined values for Body
const (
	BodyNone        Body = 0
	BodyHuman       Body = 400 // Human male body
	BodyHumanMale   Body = 400
	BodyHumanFemale Body = 401
	BodyDefault     Body = 991 // Blackthorne
	BodySystem      Body = 0x7fff
)

// MoveSpeed represents one of the available movement speeds.
type MoveSpeed byte

// All valid values for MovementSpeed
const (
	MoveSpeedNormal   MoveSpeed = 0
	MoveSpeedFast     MoveSpeed = 1
	MoveSpeedSlow     MoveSpeed = 2
	MoveSpeedTeleport MoveSpeed = 3
)

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
	TileFlagsRoof         TileFlags = 0x00000010000000
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

// Texture represents a texture ID
type Texture uint16

// All known values for Texture
const (
	TextureNone Texture = 0x0000
)

// Animation represents an animation ID
type Animation uint16

// Light represents a light graphic
type Light uint16
