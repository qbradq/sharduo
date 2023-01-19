package util

import (
	"log"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
)

type dsobj interface {
	Serial() uo.Serial
	SetSerial(uo.Serial)
	marshal.Marshaler
	marshal.Unmarshaler
}

// DataStore is a file-backed key-value store.
type DataStore[K dsobj] struct {
	SerialPool[K]
	// Pool of deserialization data for rebuilding the objects
	tfoPool map[uo.Serial]*TagFileObject
	// Pool of unmarshaled object data for rebuilding the objects
	toPool map[uo.Serial]*marshal.TagObject
	// Debug name of the data store
	name string
}

// NewDataStore initializes and returns a new DataStore object.
func NewDataStore[K dsobj](name string, rng uo.RandomSource) *DataStore[K] {
	return &DataStore[K]{
		SerialPool: *NewSerialPool[K](name, rng),
		tfoPool:    make(map[uo.Serial]*TagFileObject),
		toPool:     make(map[uo.Serial]*marshal.TagObject),
		name:       name,
	}
}

// Data returns the underlying data store.
func (s *DataStore[K]) Data() map[uo.Serial]K {
	return s.SerialPool.objects
}

// LoadMarshalData loads all of the object data from the given segment and
// rebuilds the database with those objects. UnmarsahlObjects and
// AfterUnmarshalObjects should be called afterwards to complete the load and
// free internal resources.
func (s *DataStore[K]) LoadMarshalData(seg *marshal.TagFileSegment) {
	for i := 0; i < int(seg.RecordCount()); i++ {
		to := seg.TagObject()
		ctor := marshal.Constructor(to.Type)
		if ctor == nil {
			log.Printf("warning: object %s ctor not found for type %d, object leaked: %+v", to.Serial, to.Type, to)
			continue
		}
		iface := ctor()
		k, ok := iface.(K)
		if !ok {
			log.Printf("warning: object %s did not implement the expected interface, object leaked", to.Serial)
			continue
		}
		s.toPool[to.Serial] = to
		s.objects[to.Serial] = k
	}
}

// UnmarshalObjects executes the Unmarshal function for all objects in the
// datastore. LoadMarshalData must be called before this function, and
// AfterUnmarshalObjects should be called after.
func (s *DataStore[K]) UnmarshalObjects() {
	for serial, k := range s.objects {
		to := s.toPool[serial]
		k.Unmarshal(to)
	}
}

// AfterUnmarshalObjects executes the AfterUnmarshal function for all objects in
// the datastore. LoadMarshalData and UnmarshalData should be called first.
func (s *DataStore[K]) AfterUnmarshalObjects() {
	for serial, k := range s.objects {
		to := s.toPool[serial]
		k.AfterUnmarshal(to)
	}
	s.toPool = make(map[uo.Serial]*marshal.TagObject)
}
