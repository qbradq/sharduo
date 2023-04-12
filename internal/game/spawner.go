package game

import (
	"log"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("Spawner", marshal.ObjectTypeSpawner, func() any { return &Spawner{} })
}

// SpawnedObject describes one object that was spawned.
type SpawnedObject struct {
	Object            Object  // Pointer to the object that was spawned
	NextSpawnDeadline uo.Time // When should the object be spawned again
}

// SpawnerEntry describes one object to spawn.
type SpawnerEntry struct {
	Template string          // Name of the template of the object
	Amount   int             // Amount of objects to spawn in the area
	Delay    uo.Time         // Delay between object disappearance and respawn
	Objects  []SpawnedObject // Pointers to the spawned objects
}

// Spawner manages one or more objects that are re-spawned after they are
// removed.
type Spawner struct {
	BaseItem
	Radius  int            // Radius of the spawning region
	Entries []SpawnerEntry // All objects to spawn
}

// NoRent implements the Object interface.
func (o *Spawner) NoRent() bool { return false }

// ObjectType implements the Object interface.
func (o *Spawner) ObjectType() marshal.ObjectType { return marshal.ObjectTypeSpawner }

// Marshal implements the marshal.Marshaler interface.
func (o *Spawner) Marshal(s *marshal.TagFileSegment) {
	o.BaseItem.Marshal(s)
	s.PutInt(uint32(o.Radius))
	s.PutByte(byte(len(o.Entries)))
	for _, e := range o.Entries {
		s.PutString(e.Template)
		s.PutInt(uint32(e.Amount))
		s.PutLong(uint64(e.Delay))
	}
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (o *Spawner) Unmarshal(s *marshal.TagFileSegment) {
	o.BaseItem.Unmarshal(s)
	o.Radius = int(s.Int())
	count := int(s.Byte())
	for i := 0; i < count; i++ {
		o.Entries = append(o.Entries, SpawnerEntry{
			Template: s.String(),
			Amount:   int(s.Int()),
			Delay:    uo.Time(s.Long()),
		})
	}
}

// Visibility implements the Object interface.
func (o *Spawner) Visibility() uo.Visibility {
	return uo.VisibilityStaff
}

// Update implements the Object interface.
func (o *Spawner) Update(t uo.Time) {
}

// Weight implements the Object interface
func (o *Spawner) Weight() float32 {
	return 0
}

// AddObject implements the Object interface
func (o *Spawner) AddObject(c Object) bool {
	o.SetParent(c)
	return false
}

// FullRespawn respawns all objects
func (o *Spawner) FullRespawn() {
	for i := range o.Entries {
		o.RespawnEntry(i)
	}
}

// RespawnEntry respawns all objects for entry n
func (o *Spawner) RespawnEntry(n int) {
	if n < 0 || n >= len(o.Entries) {
		return
	}
	e := o.Entries[n]
	// Initialize the object descriptors if needed
	if len(e.Objects) == 0 && e.Amount != 0 {
		for i := 0; i < e.Amount; i++ {
			e.Objects = append(e.Objects, SpawnedObject{
				Object:            nil,
				NextSpawnDeadline: uo.TimeZero,
			})
		}
	}
	// Scan the object descriptors and spawn objects
	for i, so := range e.Objects {
		if so.Object != nil && !so.Object.Removed() {
			Remove(so.Object)
		}
		so.Object = nil
		so.NextSpawnDeadline = world.Time() + e.Delay
		if len(e.Template) == 0 {
			continue
		}
		so.Object = o.Spawn(e.Template)
		so.NextSpawnDeadline = uo.TimeZero
		e.Objects[i] = so
	}
}

func (o *Spawner) Spawn(which string) Object {
	so := template.Create(which).(Object)
	if so == nil {
		log.Printf("warning: template %s not found in Spawner.Spawn()", which)
	}
	for tries := 0; tries < 8; tries++ {
		nl := uo.Location{
			X: o.location.X + int16(world.Random().Random(-o.Radius, o.Radius)),
			Y: o.location.Y + int16(world.Random().Random(-o.Radius, o.Radius)),
			Z: o.location.Z,
		}
		floor, ceiling := world.Map().GetFloorAndCeiling(nl, false)
		if floor == nil {
			if ceiling == nil {
				continue
			}
			floor, _ = world.Map().GetFloorAndCeiling(uo.Location{
				X: nl.X,
				Y: nl.Y,
				Z: ceiling.Z() + ceiling.Highest(),
			}, false)
			if floor == nil {
				continue
			}
		}
		nl.Z = floor.Z() + floor.StandingHeight()
		so.SetLocation(nl)
		if world.Map().AddObject(so) {
			break
		}
	}
	return so
}
