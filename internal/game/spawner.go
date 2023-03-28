package game

import (
	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("Spawner", marshal.ObjectTypeSpawner, func() any { return &Spawner{} })
}

// SpawnerEntry describes one object to spawn.
type SpawnerEntry struct {
	// Name of the template of the object
	Template string
	// Amount of objects to spawn in the area
	Amount int
	// Pointers to the spawned objects
	Objects []Object
}

// Spawner manages one or more objects that are re-spawned after they are
// removed.
type Spawner struct {
	BaseItem
	// Radius of the spawning region
	Radius int
	// All objects to spawn
	Entries []SpawnerEntry
}

// NoRent implements the Object interface.
func (o *Spawner) NoRent() bool { return true }

// Visibility implements the Object interface.
func (o *Spawner) Visibility() uo.Visibility {
	return uo.VisibilityStaff
}
