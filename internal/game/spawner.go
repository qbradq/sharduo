package game

import (
	"github.com/qbradq/sharduo/lib/marshal"
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
func (o *Spawner) NoRent() bool { return true }

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
