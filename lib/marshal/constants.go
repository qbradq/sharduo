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
