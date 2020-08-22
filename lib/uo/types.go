package uo

import "math/rand"

// A Serial is a 31-bit value with the following characteristics:
// The zero value is also the "invalid value" value
// No Serial will have a value greater than 2^31-1
type Serial uint32

// Pre-defined values of Serial
const (
	SerialSystem Serial = 0xffffffff
)

// A Dir is a 3-bit value indicating the direction a mobile is facing
type Dir byte

// Dir value meanings
const (
	DirNorth       Dir = 0
	DirNorthEast   Dir = 1
	DirEast        Dir = 2
	DirSouthEast   Dir = 3
	DirSouth       Dir = 4
	DirSouthWest   Dir = 5
	DirWest        Dir = 6
	DirNorthWest   Dir = 7
	DirRunningFlag Dir = 0x80
)

// A Body is a 16-bit value that describes the set of animations to use for a
// mobile. Body values used by UO range 1-999.
type Body uint16

// Pre-defined values for Body
const (
	BodySystem Body = 0xffff
)

// A Hue is a 16-bit value that describes the rendering mode of an object.
// Hues have the following characteristics:
// The zero value means "default rendering mode"
// Values 1-3000 inclusive select a set of 16 colors from the file "hues.mul"
//   that replace the first 16 color indicies (the grayscales).
// The special value -1 (0xffff) will do the shadow dragon alpha effect.
type Hue uint16

// Important hue values
const (
	HueDefault Hue = 0
	HueMin     Hue = 1
	HueBlack   Hue = 1
	HueDieMin  Hue = 2
	HueDieMax  Hue = 1001
	HueSkinMin Hue = 1002
	HueSkinMax Hue = 1058
	HueMax     Hue = 3000
)

// RandomSkinHue returns a random skin hue
func RandomSkinHue() Hue {
	return Hue(rand.Intn(int(HueSkinMax-HueSkinMin))) + HueSkinMin
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
)

// A StatusFlag describes the status of a mobile
type StatusFlag byte

// StatusFlag constants
const (
	StatusNormal StatusFlag = 0
)

// A Noto is a 3-bit value describing the notoriety status of a mobile
// The zero-value is invalid
type Noto byte

// Notoriety constants
const (
	NotoInvalid      Noto = 0
	NotoInnocent     Noto = 1
	NotoFriend       Noto = 2
	NotoCriminal     Noto = 3
	NotoEnemy        Noto = 4
	NotoMurderer     Noto = 5
	NotoInvulnerable Noto = 6
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

// All meaningfull LoginDeniedReason values
const (
	LoginDeniedReasonBadPass        LoginDeniedReason = 0 // Password invalid for user
	LoginDeniedReasonAccountInUse   LoginDeniedReason = 1 // The account already has an active season
	LoginDeniedReasonAccountBlocked LoginDeniedReason = 2 // The account has been blocked for some reason
)

// SpeechType represents the type of speech being requested or sent.
type SpeechType byte

// All meaningfull SpeechType values
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

// All meaningfull Font values
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
