package uo

import (
	"fmt"
	"strconv"
	"strings"
)

func flexNum(in []byte) int {
	if len(in) < 1 {
		return 0
	}
	if in[0] != '"' {
		v, err := strconv.ParseInt(string(in), 0, 32)
		if err != nil {
			return 0
		}
		return int(v)
	}
	v, err := strconv.ParseInt(string(in[1:len(in)-1]), 0, 32)
	if err != nil {
		return 0
	}
	return int(v)
}

// Random constants
const (
	MinStackAmount            int  = 1
	MaxStackAmount            int  = 60000
	MinViewRange              int  = 5
	MaxViewRange              int  = 18
	MaxUpdateRange            int  = 24
	ChunkWidth                int  = 8
	ChunkHeight               int  = 8
	MapWidth                  int  = 7168
	MapHeight                 int  = 4096
	MapOverworldWidth         int  = 5120
	MapChunksWidth            int  = MapWidth / ChunkWidth
	MapChunksHeight           int  = MapHeight / ChunkHeight
	MapMinZ                   int  = -128
	MapMaxZ                   int  = 127
	StatsCapDefault           int  = 225
	MaxFollowers              int  = 5
	MaxUseRange               int  = 3
	MaxLiftRange              int  = 3
	MaxDropRange              int  = 3
	MaxContainerViewRange     int  = 3
	MaxItemStackHeight        int  = 18
	DefaultMaxContainerWeight int  = 400
	DefaultMaxContainerItems  int  = 125
	RandomDropX               int  = -1
	RandomDropY               int  = -1
	TargetCanceledX           int  = -1
	TargetCanceledY           int  = -1
	ContainerOpenLowerLimit   int  = -8
	ContainerOpenUpperLimit   int  = 16
	PlayerHeight              int  = 16
	StepHeight                int  = 2
	SpeechWhisperRange        int  = 1
	SpeechNormalRange         int  = 12
	SpeechEmoteRange          int  = SpeechNormalRange
	SpeechYellRange           int  = MaxViewRange
	WalkDelay                 Time = 6
	RunDelay                  Time = 3
	MountedWalkDelay          Time = 4
	MountedRunDelay           Time = 2
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

// BoundUpdateRange bounds an update range between MinViewRange and
// MaxUpdateRange
func BoundUpdateRange(r int) int {
	if r < MinViewRange {
		return MinViewRange
	} else if r > MaxUpdateRange {
		return MaxUpdateRange
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
	LayerLastVisible              Layer = LayerLegArmor
	LayerCount                    Layer = LayerLastValid + 1
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
	NotorietyInvalid      Notoriety = 0 // Invalid value
	NotorietyInnocent     Notoriety = 1 // Blue - not attackable
	NotorietyFriend       Notoriety = 2 // Green - guild or faction ally, attackable
	NotorietyAttackable   Notoriety = 3 // Gray - attackable but not criminal
	NotorietyCriminal     Notoriety = 4 // Gray - attackable, criminal
	NotorietyEnemy        Notoriety = 5 // Orange - guild or faction enemy
	NotorietyMurderer     Notoriety = 6 // Red - attackable, murderer
	NotorietyInvulnerable Notoriety = 7 // Yellow - invulnerable, vendors etc
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
	BodyGhost       Body = 970
	BodyCounselor   Body = 987 // GM body
	BodyDefault     Body = 991 // Blackthorne
	BodySystem      Body = 0x7fff
)

// UnmarshalJSON implements json.Unmarshaler.
func (b *Body) UnmarshalJSON(in []byte) error {
	*b = Body(flexNum(in))
	return nil
}

// MoveSpeed represents one of the available movement speeds.
type MoveSpeed byte

// All valid values for MovementSpeed
const (
	MoveSpeedNormal   MoveSpeed = 0
	MoveSpeedFast     MoveSpeed = 1
	MoveSpeedSlow     MoveSpeed = 2
	MoveSpeedTeleport MoveSpeed = 3
)

// Texture represents a texture ID
type Texture uint16

// All known values for Texture
const (
	TextureNone Texture = 0x0000
)

// Light represents a light graphic
type Light uint16

// MoveItemRejectReason represents the reasons that an item move request might
// be rejected.
type MoveItemRejectReason byte

// All known values for MoveItemRejectReason
const (
	MoveItemRejectReasonCannotLift         MoveItemRejectReason = 0
	MoveItemRejectReasonOutOfRange         MoveItemRejectReason = 1
	MoveItemRejectReasonOutOfSight         MoveItemRejectReason = 2
	MoveItemRejectReasonBelongsToAnother   MoveItemRejectReason = 3
	MoveItemRejectReasonAlreadyHoldingItem MoveItemRejectReason = 4
	MoveItemRejectReasonUnspecified        MoveItemRejectReason = 5
)

// GUMP represents a gump graphic.
type GUMP uint16

// Constant values for Gump
const (
	GUMPNone             GUMP = 0x0000
	GUMPDefault          GUMP = 0x0046 // Partial skull with glowing eyes
	GUMPContainerDefault GUMP = 0x03E8 // Huge chest, old login gump
)

// UnmarshalJSON implements json.Unmarshaler.
func (g *GUMP) UnmarshalJSON(in []byte) error {
	*g = GUMP(flexNum(in))
	return nil
}

// Protocol extension request types
type ProtocolExtensionRequest byte

const (
	ProtocolExtensionRequestPartyLocations ProtocolExtensionRequest = 0x00 // Send all party member locations for tracking
	ProtocolExtensionRequestGuildLocations ProtocolExtensionRequest = 0x01 // Send all guild member locations for tracking
)

// Stat is a numeric code to refer to a mobile stat
type Stat byte

const (
	StatStrength     Stat = 0 // Strength
	StatDexterity    Stat = 1 // Dexterity
	StatIntelligence Stat = 2 // Intelligence
)

// SkillLock is a numeric code for the states of skill lock
type SkillLock byte

const (
	SkillLockUp     SkillLock = 0 // Skill is marked for gains
	SkillLockDown   SkillLock = 1 // Skill is marked for atrophy
	SkillLockLocked SkillLock = 2 // Skill is locked
)

// SkillUpdate is a code for the types of skill updates
type SkillUpdate byte

// SkillUpdate values
const (
	SkillUpdateLegacyAll    SkillUpdate = 0x00
	SkillUpdateAll          SkillUpdate = 0x02
	SkillUpdateSingle       SkillUpdate = 0xDF
	SkillUpdateLegacySingle SkillUpdate = 0xFF
)

// Cliloc is a code for the localized client messages
type Cliloc uint32

// Sound is a code referencing a sound effect on the client side
type Sound uint16

const (
	SoundInvalidDrop Sound = 0
	SoundDefaultLift Sound = 0x57
	SoundDefaultDrop Sound = 0x42
	SoundBagDrop     Sound = 0x48
)

// UnmarshalJSON implements json.Unmarshaler.
func (s *Sound) UnmarshalJSON(in []byte) error {
	*s = Sound(flexNum(in))
	return nil
}

// AnimationType indicates which animation type to play on the client side
type AnimationType uint16

// AnimationType values
const (
	AnimationTypeAttack      AnimationType = 0
	AnimationTypeParry       AnimationType = 1
	AnimationTypeBlock       AnimationType = 2
	AnimationTypeDie         AnimationType = 3
	AnimationTypeImpact      AnimationType = 4
	AnimationTypeFidget      AnimationType = 5
	AnimationTypeEat         AnimationType = 6
	AnimationTypeEmote       AnimationType = 7
	AnimationTypeAlert       AnimationType = 8
	AnimationTypeTakeOff     AnimationType = 9
	AnimationTypeLand        AnimationType = 10
	AnimationTypeSpell       AnimationType = 11
	AnimationTypeStartCombat AnimationType = 12
	AnimationTypeEndCombat   AnimationType = 13
	AnimationTypePillage     AnimationType = 14
	AnimationTypeSpawn       AnimationType = 15
)

// AnimationAction is a second parameter to animations which selects between
// different sub-animations.
type AnimationAction uint16

// AnimationAction values for weapon animations
const (
	AnimationActionSlash1H   AnimationAction = 9
	AnimationActionPierce1H  AnimationAction = 10
	AnimationActionBash1H    AnimationAction = 11
	AnimationActionBash2H    AnimationAction = 12
	AnimationActionSlash2H   AnimationAction = 13
	AnimationActionPierce2H  AnimationAction = 14
	AnimationActionShootBow  AnimationAction = 18
	AnimationActionShootXBow AnimationAction = 19
	AnimationActionWrestle   AnimationAction = 31
	AnimationActionThrowing  AnimationAction = 32
)

// LightLevel indicates how bright a light is.
type LightLevel byte

const (
	LightLevelDay   LightLevel = 0    // Brightest light level
	LightLevelNight LightLevel = 9    // OSI night
	LightLevelBlack LightLevel = 0x1F // Lowest light level
)

// GFXType is a code that indicates how a graphical effect behaves.
type GFXType byte

const (
	GFXTypeMoving    GFXType = 0 // Moves from source to target
	GFXTypeLightning GFXType = 1 // Lightning strike at source
	GFXTypeFixed     GFXType = 2 // Moves toward the absolute target location
	GFXTypeTrack     GFXType = 3 // Moves toward and tracks the source object
)

// GFXBlendMode is a code that indicates how a graphical effect is rendered.
type GFXBlendMode uint32

const (
	GFXBlendModeNormal          GFXBlendMode = 0 // Normal blending, black is transparent
	GFXBlendModeMultiply        GFXBlendMode = 1 // Darken
	GFXBlendModeScreen          GFXBlendMode = 2 // Lighten
	GFXBlendModeScreenMore      GFXBlendMode = 3 // Lighten more
	GFXBlendModeScreenLess      GFXBlendMode = 4 // Lighten less
	GFXBlendModeHalfTransparent GFXBlendMode = 5 // Transparent with black edges
	GFXBlendModeShadowBlue      GFXBlendMode = 6 // Complete shadow with blue edges
	GFXBlendModeScreenRed       GFXBlendMode = 7 // Transparent but more red?
)

// Visibility is a code that indicates the visibility state of an object.
type Visibility uint8

const (
	VisibilityVisible   Visibility = 0 // Normal visibility, everyone can see it
	VisibilityInvisible Visibility = 1 // Magical invisibility, the kind certain AI and spells can see through
	VisibilityHidden    Visibility = 2 // Hidden as in the hiding skill, certain AI and spells can see through this
	VisibilityStaff     Visibility = 3 // Only staff can see this object
	VisibilityNone      Visibility = 4 // This object is never shown to the client
)

// LootType is a code that indicates what happens to items when a player dies
// with them in their inventory as well as how quickly the item decays.
type LootType uint8

const (
	LootTypeNormal  LootType = 0 // Drops on death, decays in 1 hour
	LootTypeBlessed LootType = 1 // Does not drop on death, decays in 1 hour
	LootTypeNewbie  LootType = 2 // Does not drop on death, decays in 15 seconds
	LootTypeSystem  LootType = 3 // Does not drop on death, never decays
)

// UnmarshalJSON implements json.Unmarshaler.
func (t *LootType) UnmarshalJSON(in []byte) error {
	if len(in) < 1 {
		*t = LootTypeNormal
	} else if in[0] == '"' {
		s := strings.ToLower(string(in[1 : len(in)-1]))
		switch s {
		case "blessed":
			*t = LootTypeBlessed
		case "newbie":
			*t = LootTypeNewbie
		case "system":
			*t = LootTypeSystem
		default:
			panic(fmt.Errorf("unsupported loot type %s", s))
		}
	}
	return nil
}

// Door location offsets
var DoorOffsets = []Point{
	{X: -1, Y: 1},
	{X: 1, Y: 1},
	{X: -1, Y: 0},
	{X: 1, Y: -1},
	{X: 1, Y: 1},
	{X: 1, Y: -1},
	{X: 0, Y: 0},
	{X: 0, Y: -1},
}

// MacroType is a code indicating what type of macro is requested in a client
// packet 0x12.
type MacroType uint8

const (
	MacroTypeSkill    MacroType = 0 // Skill use request
	MacroTypeSpell    MacroType = 1 // Spell cast request
	MacroTypeOpenDoor MacroType = 2 // Open door request
	MacroTypeAction   MacroType = 3 // 0 = bow, 1 = salute
	MacroTypeInvalid  MacroType = 4 // Parsing error
)

// Music is a code that describes which music track to play on the client side.
type Music uint16

const (
	MusicApproach  Music = 0  // approach
	MusicBritain1  Music = 1  // britain1
	MusicBritain2  Music = 2  // britain2
	MusicBtcastle  Music = 3  // btcastle
	MusicBucsden   Music = 4  // bucsden
	MusicCave01    Music = 5  // cave01
	MusicCombat1   Music = 6  // combat1
	MusicCombat2   Music = 7  // combat2
	MusicCombat3   Music = 8  // combat3
	MusicCove      Music = 9  // cove
	MusicCreate1   Music = 10 // create1
	MusicDeath     Music = 11 // death
	MusicDragflit  Music = 12 // dragflit
	MusicDungeon2  Music = 13 // dungeon2
	MusicDungeon9  Music = 14 // dungeon9
	MusicForest_a  Music = 15 // forest_a
	MusicIntown01  Music = 16 // intown01
	MusicJhelom    Music = 17 // jhelom
	MusicJungle_a  Music = 18 // jungle_a
	MusicLBCastle  Music = 19 // lbcastle
	MusicLinelle   Music = 20 // linelle
	MusicMagincia  Music = 21 // magincia
	MusicMinoc     Music = 22 // minoc
	MusicMoonglow  Music = 23 // moonglow
	MusicMountn_a  Music = 24 // mountn_a
	MusicNujelm    Music = 25 // nujelm
	MusicOcllo     Music = 26 // ocllo
	MusicOldult01  Music = 27 // oldult01
	MusicOldult02  Music = 28 // oldult02
	MusicOldult03  Music = 29 // oldult03
	MusicOldult04  Music = 30 // oldult04
	MusicOldult05  Music = 31 // oldult05
	MusicOldult06  Music = 32 // oldult06
	MusicPlains_a  Music = 33 // plains_a
	MusicSailing   Music = 34 // sailing
	MusicSamlethe  Music = 35 // samlethe
	MusicSerpents  Music = 36 // serpents
	MusicSkarabra  Music = 37 // skarabra
	MusicStones2   Music = 38 // stones2
	MusicSwamp_a   Music = 39 // swamp_a
	MusicTavern01  Music = 40 // tavern01
	MusicTavern02  Music = 41 // tavern02
	MusicTavern03  Music = 42 // tavern03
	MusicTavern04  Music = 43 // tavern04
	MusicTrinsic   Music = 44 // trinsic
	MusicVesper    Music = 45 // vesper
	MusicVictory   Music = 46 // victory
	MusicWind      Music = 47 // wind
	MusicYew       Music = 48 // yew
	MusicLastValid Music = MusicYew
)
