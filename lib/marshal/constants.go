package marshal

// Segment represents a segment name within a file
type Segment byte

// Segment values
const (
	SegmentAccounts     Segment = 0
	SegmentMap          Segment = 1
	SegmentTimers       Segment = 2
	SegmentWorld        Segment = 3
	SegmentObjectList   Segment = 4
	SegmentObjectsStart Segment = 0x7F // THIS MUST BE THE LAST ENTRY!
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
	TagEndOfList         Tag = 0
	TagArticleA          Tag = 1
	TagArticleAn         Tag = 2
	TagFacing            Tag = 3
	TagDoubleClick       Tag = 4
	TagGraphic           Tag = 5
	TagFlippedGraphic    Tag = 6
	TagFlipped           Tag = 7
	TagDyable            Tag = 8
	TagWeight            Tag = 9
	TagStackable         Tag = 10
	TagAmount            Tag = 11
	TagPlural            Tag = 12
	TagIsPlayerCharacter Tag = 13
	TagIsFemale          Tag = 14
	TagBody              Tag = 15
	TagNotoriety         Tag = 16
	TagCursor            Tag = 17
	TagEquipment         Tag = 18
	TagStrength          Tag = 19
	TagDexterity         Tag = 20
	TagIntelligence      Tag = 21
	TagHitPoints         Tag = 22
	TagMana              Tag = 23
	TagStamina           Tag = 24
	TagSkills            Tag = 25
	TagLayer             Tag = 26
	TagRequiredSkill     Tag = 27
	TagManagedObject     Tag = 28
	TagContents          Tag = 29
	TagGump              Tag = 30
	TagMaxWeight         Tag = 31
	TagMaxItems          Tag = 32
	TagBounds            Tag = 33
	TagViewRange         Tag = 34
	TagLastValidValue    Tag = TagViewRange
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

// ObjectType are the concrete Go types in the game package.
type ObjectType byte

const (
	ObjectTypeObject            ObjectType = 0  // BaseObject
	ObjectTypeStatic            ObjectType = 1  // StaticItem
	ObjectTypeItem              ObjectType = 2  // BaseItem
	ObjectTypeWearable          ObjectType = 3  // BaseWearable
	ObjectTypeWearableContainer ObjectType = 4  // WearableContainer
	ObjectTypeWeapon            ObjectType = 5  // BaseWeapon
	ObjectTypeContainer         ObjectType = 6  // BaseContainer
	ObjectTypeMountItem         ObjectType = 7  // MountItem
	ObjectTypeMobile            ObjectType = 8  // BaseMobile
	ObjectTypeAccount           ObjectType = 9  // Account
	ObjectTypeSpawner           ObjectType = 10 // Spawner
)
