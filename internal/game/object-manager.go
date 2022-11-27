package game

import (
	"github.com/qbradq/sharduo/internal/util"
	"github.com/qbradq/sharduo/lib/uo"
)

// ObjectManager manages all of the game objects on the server.
type ObjectManager struct {
	// Database of all objects
	ds *util.DataStore
}

// NewObjectManager creates and returns a new ObjetManager object.
func NewObjectManager(dbpath string) *ObjectManager {
	return &ObjectManager{
		ds: util.OpenOrCreateDataStore(dbpath, false),
	}
}

// Add adds the object to the manager, which will then assign a unique ID
// to the object.
func (m *ObjectManager) Add(o util.Serializeable, serialType uo.SerialType) {
	m.ds.Add(o, "", serialType)
}
