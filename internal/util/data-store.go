package util

import (
	"io"
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
	Index map[string]uo.Serial
	// SerialManager for objects in this store
	sm uo.SerialManager
	// Read/Write lock
	lock sync.RWMutex
}

// NewDataStore initializes and returns a new DataStore object.
func NewDataStore(index bool) *DataStore {
	return &DataStore{
		Objects:    make(map[uo.Serial]Serializeable),
		BuildIndex: index,
		Index:      make(map[string]uo.Serial),
		sm:         *uo.NewSerialManager(),
	}
}

// OpenOrCreateDataStore opens the named DataStore or creates a new, initialized
// one.
func OpenOrCreateDataStore(index bool) *DataStore {
	ds := NewDataStore(index)

	// Rebuild serial manager
	for serial := range ds.Objects {
		ds.sm.Add(serial)
	}

	return ds
}

// Save writes all objects to the given io.Writer. It is a wrapper for Write().
func (s *DataStore) Save(dataStoreName string, w io.Writer) []error {
	tfw := NewTagFileWriter(w)
	tfw.WriteCommentLine(dataStoreName + " data store")
	s.Write(tfw)
	tfw.WriteCommentLine("END OF FILE")
	return tfw.Errors()
}

// Write writes all objects in the the data store to the given tag file.
func (s *DataStore) Write(f *TagFileWriter) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, s := range s.Objects {
		f.WriteObject(s)
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

// Add adds the object to the store, assigning it a unique ID.
func (s *DataStore) Add(o Serializeable, name string, serialType uo.SerialType) {
	s.lock.Lock()
	defer s.lock.Unlock()
	o.SetSerial(s.UniqueSerial(serialType))
	if s.BuildIndex {
		s.Index[name] = o.GetSerial()
	}
	s.Objects[o.GetSerial()] = o
	s.sm.Add(o.GetSerial())
}

// UniqueSerial returns a uo.Serial that is unique to this dataset.
func (s *DataStore) UniqueSerial(serialType uo.SerialType) uo.Serial {
	return s.sm.New(serialType)
}
