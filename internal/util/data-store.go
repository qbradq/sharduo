package util

import (
	"sync"

	"github.com/qbradq/sharduo/lib/uo"
)

// DataStore is a file-backed key-value store.
type DataStore struct {
	// Collection of all objects in the data store
	Objects map[uo.Serial]Serializeable
	// Flag that tells us whether or not to build a string index
	BuildIndex bool
	// String index
	Index  map[string]uo.Serial
	dbpath string
	lock   sync.RWMutex
}

// NewDataStore initializes and returns a new DataStore object.
func NewDataStore(dbpath string, index bool) *DataStore {
	return &DataStore{
		Objects:    make(map[uo.Serial]Serializeable),
		BuildIndex: index,
		Index:      make(map[string]uo.Serial),
		dbpath:     dbpath,
	}
}

// Get returns the named object or nil.
func (s *DataStore) Get(which uo.Serial) Serializeable {
	s.lock.RLock()
	defer s.lock.RUnlock()
	o, ok := s.Objects[which]
	if !ok {
		return nil
	}
	return o
}

// GetByIndex uses the string index, if any, to fetch the object. Returns nil
// if the object could not be found.
func (s *DataStore) GetByIndex(name string) Serializeable {
	if !s.BuildIndex {
		return nil
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	ds, ok := s.Index[name]
	if !ok {
		return nil
	}
	return s.Objects[ds]
}

// Set adds the object to the data store.
func (s *DataStore) Set(o Serializeable, name string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Objects[o.GetSerial()] = o
	if s.BuildIndex {
		s.Index[name] = o.GetSerial()
	}
}
