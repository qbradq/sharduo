package game

import (
	"io"

	"github.com/qbradq/sharduo/internal/util"
	"github.com/qbradq/sharduo/lib/uo"
)

// ObjectManager manages all of the game objects on the server.
type ObjectManager struct {
	// Database of all objects
	ds *util.DataStore
}

// NewObjectManager creates and returns a new ObjetManager object.
func NewObjectManager() *ObjectManager {
	return &ObjectManager{
		ds: util.OpenOrCreateDataStore(false),
	}
}

// Add adds the object to the manager, which will then assign a unique ID
// to the object.
func (m *ObjectManager) Add(o util.Serializeable, serialType uo.SerialType) {
	m.ds.Add(o, "", serialType)
}

// NewItem adds the newly-created item to the object manager and returns the
// item. This method has the side-effect of setting the ID of the item.
func (m *ObjectManager) NewItem(item Item) Item {
	s := item.(util.Serializeable)
	m.ds.Add(s, "", uo.SerialTypeItem)
	return item
}

// NewMobile adds the newly-created mobile to the object manager and returns the
// mobile. This method has the side-effect of setting the ID of the mobile.
func (m *ObjectManager) NewMobile(mob Mobile) Mobile {
	s := mob.(util.Serializeable)
	m.ds.Add(s, "", uo.SerialTypeMobile)
	return mob
}

// Save writes the object data store in TagFile format.
func (m *ObjectManager) Save(w io.Writer) []error {
	return m.ds.Save("objects", w)
}
