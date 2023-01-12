package marshal

// Segment represents a segment name within a file
type Segment byte

// Segment values
const (
	SegmentStringDictionary Segment = 0
	SegmentAccounts         Segment = 1
	SegmentObjects          Segment = 2
	SegmentMap              Segment = 3
	SegmentTimers           Segment = 4
	SegmentWorld            Segment = 5
)

// Tag represents a property name within an object segment as opposed to a raw
// segment.
type Tag byte

// Tag values
// WARNING: DO NOT CHANGE ANY CONSTANT VALUES IN THIS BLOCK OR IT WILL BREAK
//          COMPATIBILITY WITH PREVIOUS SAVES!!!
// Note: The following properties are held within the object header and are not
//       preceded by a Tag value.
// object type code
// serial
// parent
// templateName
// name
// hue
// location
const (
	TagArticleA          Tag = 0
	TagArticleAn         Tag = 1
	TagFacing            Tag = 2
	TagOnDoubleClick     Tag = 3
	TagGraphic           Tag = 4
	TagFlippedGraphic    Tag = 5
	TagFlipped           Tag = 6
	TagDyable            Tag = 7
	TagWeight            Tag = 8
	TagStackable         Tag = 9
	TagAmount            Tag = 10
	TagPlural            Tag = 11
	TagIsPlayerCharacter Tag = 12
	TagIsFemale          Tag = 13
	TagBody              Tag = 14
	TagNotoriety         Tag = 15
	TagCursor            Tag = 16
	TagEquipment         Tag = 17
	TagStrength          Tag = 18
	TagDexterity         Tag = 19
	TagIntelligence      Tag = 20
	TagHitPoints         Tag = 21
	TagMana              Tag = 22
	TagStamina           Tag = 23
	TagSkills            Tag = 24
	TagLayer             Tag = 25
	TagRequiredSkill     Tag = 26
	TagManagedObject     Tag = 27
	TagContents          Tag = 28
	TagGump              Tag = 29
	TagMaxWeight         Tag = 30
	TagMaxItems          Tag = 31
	TagBounds            Tag = 32
	TagUsername          Tag = 33
	TagPasswordHash      Tag = 34
	TagPlayerMobile      Tag = 35
)

// TagValue is a code to indicate what kind of data a value contains
type TagValue byte

// TagValue values
const (
	TagValueBool           TagValue = 0 // If the tag is present the value is true, otherwise false
	TagValueByte           TagValue = 1 // 8-bit number
	TagValueShort          TagValue = 2 // 16-bit number
	TagValueInt            TagValue = 3 // 32-bit number
	TagValueLong           TagValue = 4 // 64-bit number
	TagValueString         TagValue = 5 // 32-bit string reference
	TagValueReferenceSlice TagValue = 6 // Slice of 32-bit object references
	TagValueLocation       TagValue = 7 // Tuple of int16,int16,int8
	TagValueBounds         TagValue = 8 // Tuple of int16,int16,int16,int16
	TagValueShortSlice     TagValue = 9 // Slice of 16-bit numbers
)

// Mapping of TagValue codes to Tag codes
var tagValueMapping = []TagValue{
	TagValueBool,           // TagArticleA
	TagValueBool,           // TagArticleAn
	TagValueByte,           // TagFacing,
	TagValueString,         // TagOnDoubleClick,
	TagValueShort,          // TagGraphic,
	TagValueShort,          // TagFlippedGraphic,
	TagValueBool,           // TagFlipped,
	TagValueBool,           // TagDyable,
	TagValueInt,            // TagWeight = weight*1000
	TagValueBool,           // TagStackable
	TagValueShort,          // TagAmount
	TagValueString,         // TagPlural
	TagValueBool,           // TagIsPlayerCharacter
	TagValueBool,           // TagIsFemale
	TagValueShort,          // TagBody
	TagValueByte,           // TagNotoriety
	TagValueInt,            // TagCursor
	TagValueReferenceSlice, // TagEquipment
	TagValueShort,          // TagStrength
	TagValueShort,          // TagDexterity
	TagValueShort,          // TagIntelligence
	TagValueShort,          // TagHitPoints
	TagValueShort,          // TagMana
	TagValueShort,          // TagStamina
	TagValueShortSlice,     // TagSkills
	TagValueByte,           // TagLayer
	TagValueByte,           // TagRequiredSkill
	TagValueInt,            // TagManagedObject
	TagValueReferenceSlice, // TagContents Tag
	TagValueShort,          // TagGump
	TagValueShort,          // TagMaxWeight
	TagValueShort,          // TagMaxItems
	TagValueBounds,         // TagBounds
	TagValueString,         // TagUsername
	TagValueString,         // TagPasswordHash
	TagValueInt,            // TagPlayerMobile
}
