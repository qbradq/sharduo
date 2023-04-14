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
	Template string           // Name of the template of the object
	Amount   int              // Amount of objects to spawn in the area
	Delay    uo.Time          // Delay between object disappearance and respawn
	Objects  []*SpawnedObject // Pointers to the spawned objects
}

// Spawner manages one or more objects that are re-spawned after they are
// removed.
type Spawner struct {
	BaseItem
	Radius  int             // Radius of the spawning region
	Entries []*SpawnerEntry // All objects to spawn
}

// NoRent implements the Object interface.
func (o *Spawner) NoRent() bool { return false }

// ObjectType implements the Object interface.
func (o *Spawner) ObjectType() marshal.ObjectType { return marshal.ObjectTypeSpawner }

// Marshal implements the marshal.Marshaler interface.
func (o *Spawner) Marshal(s *marshal.TagFileSegment) {
	// Prepare for the marshal
	o.deleteRemovedObjects(world.Time())
	// Marshal chain
	o.BaseItem.Marshal(s)
	// Spawner-level data
	s.PutInt(uint32(o.Radius))
	// Entry-level data
	s.PutByte(byte(len(o.Entries)))
	for _, e := range o.Entries {
		s.PutString(e.Template)
		s.PutInt(uint32(e.Amount))
		s.PutLong(uint64(e.Delay))
		// Object-level data
		s.PutInt(uint32(len(e.Objects)))
		for _, so := range e.Objects {
			if so.Object == nil {
				// An object is scheduled to spawn in the future, just record
				// the deadline.
				s.PutLong(uint64(so.NextSpawnDeadline))
			} else {
				// An object already exists. We indicate this by writing the
				// uo.TimeNever value. This flags to the unmarshaler that we
				// also have a location following.
				s.PutLong(uint64(uo.TimeNever))
				s.PutLocation(so.Object.Location())
			}
		}
	}
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (o *Spawner) Unmarshal(s *marshal.TagFileSegment) {
	o.BaseItem.Unmarshal(s)
	// Spawner-level values
	o.Radius = int(s.Int())
	// Entity-level values
	count := int(s.Byte())
	for i := 0; i < count; i++ {
		e := &SpawnerEntry{
			Template: s.String(),
			Amount:   int(s.Int()),
			Delay:    uo.Time(s.Long()),
		}
		// Object-level values
		objCount := int(s.Int())
		for iObj := 0; iObj < objCount; iObj++ {
			so := &SpawnedObject{
				NextSpawnDeadline: uo.Time(s.Long()),
			}
			if so.NextSpawnDeadline == uo.TimeNever {
				// The object was spawned when we saved so create a replacement
				// now.
				l := s.Location()
				obj, ok := template.Create(e.Template).(Object)
				if !ok {
					log.Printf("warning: failed to create object from template %s", e.Template)
					continue
				}
				obj.SetLocation(l)
				world.Map().ForceAddObject(obj)
				so.Object = obj
			} // Else deadline is in the future so we don't need an object
		}
	}
}

// Visibility implements the Object interface.
func (o *Spawner) Visibility() uo.Visibility {
	return uo.VisibilityStaff
}

// deleteRemovedObjects deletes all of the removed objects from the object
// pools.
func (o *Spawner) deleteRemovedObjects(t uo.Time) {
	for _, e := range o.Entries {
		for _, so := range e.Objects {
			// TODO account for objects that have a new controller, like items picked up off the floor or tamed animals
			if so.Object != nil && so.Object.Removed() {
				so.Object = nil
				so.NextSpawnDeadline = t + e.Delay
			}
		}
	}
}

// Update implements the Object interface.
func (o *Spawner) Update(t uo.Time) {
	o.deleteRemovedObjects(t)
	// Spawn new objects when needed
	for _, e := range o.Entries {
		for _, so := range e.Objects {
			if so.Object == nil && t >= so.NextSpawnDeadline {
				so.Object = o.Spawn(e.Template)
				if so.Object != nil {
					so.NextSpawnDeadline = uo.TimeZero
				} // Else we will try to spawn again on next Update() call
			}
		}
	}
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
			e.Objects = append(e.Objects, &SpawnedObject{
				Object:            nil,
				NextSpawnDeadline: uo.TimeZero,
			})
		}
	}
	// Scan the object descriptors and spawn objects
	for _, so := range e.Objects {
		Remove(so.Object)
		so.Object = nil
		so.NextSpawnDeadline = world.Time() + e.Delay
		if len(e.Template) == 0 {
			continue
		}
		so.Object = o.Spawn(e.Template)
		if so.Object != nil {
			so.NextSpawnDeadline = uo.TimeZero
		} else {
			so.NextSpawnDeadline = world.Time() + e.Delay
		}
	}
}

func (o *Spawner) Spawn(which string) Object {
	so := template.Create(which).(Object)
	if so == nil {
		log.Printf("warning: template %s not found in Spawner.Spawn()", which)
		return nil
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
