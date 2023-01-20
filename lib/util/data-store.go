package util

import (
	"log"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
)

type dsobj interface {
	Serial() uo.Serial
	SetSerial(uo.Serial)
	SerialType() uo.SerialType
	marshal.Marshaler
	marshal.Unmarshaler
}

// DataStore is a file-backed key-value store.
type DataStore[K dsobj] struct {
	// Pool of managed objects
	objects map[uo.Serial]K
	// Pool of deserialization data for rebuilding the objects
	tagsPool map[uo.Serial]*marshal.TagCollection
	// Random number source for serials
	rng uo.RandomSource
}

// NewDataStore initializes and returns a new DataStore object.
func NewDataStore[K dsobj](rng uo.RandomSource) *DataStore[K] {
	return &DataStore[K]{
		objects:  make(map[uo.Serial]K),
		tagsPool: make(map[uo.Serial]*marshal.TagCollection),
		rng:      rng,
	}
}

// Add adds a new object to the datastore assigning it a new serial of the
// correct type.
func (s *DataStore[K]) Add(k K, t uo.SerialType) {
	for {
		var serial uo.Serial
		switch t {
		case uo.SerialTypeItem:
			serial = uo.RandomItemSerial(s.rng)
		case uo.SerialTypeMobile:
			serial = uo.RandomMobileSerial(s.rng)
		default:
			log.Println("error: DataStore.Add(K, SerialType) SerialType was an invalid value")
			return
		}
		_, used := s.objects[serial]
		if !used {
			s.objects[serial] = k
			break
		}
	}
}

// Insert inserts the object into the datastore with its current serial and will
// overwrite existing values without warning. This is typically only used when
// rebuilding the dataset from an external data source.
func (s *DataStore[K]) Insert(k K) {
	s.objects[k.Serial()] = k
}

// Data returns the underlying data store.
func (s *DataStore[K]) Data() map[uo.Serial]K {
	return s.objects
}

// UnmarshalObjects unmarshals objects from raw data. AfterUnmarshalObjects must
// be called after this to complete the load process and free internal memory.
func (s *DataStore[K]) UnmarshalObjects(seg *marshal.TagFileSegment) {
	for i := uint32(0); i < seg.RecordCount(); i++ {
		serial := uo.Serial(seg.Int())
		if k, ok := s.objects[serial]; ok {
			tags := k.Unmarshal(seg)
			s.tagsPool[serial] = tags
		} else {
			log.Printf("failed to find object %s", serial.String())
		}
	}
}

// AfterUnmarshalObjects executes the AfterUnmarshal function for all objects in
// the datastore. UnmarshalObjects must be called first.
func (s *DataStore[K]) AfterUnmarshalObjects() {
	for serial, k := range s.objects {
		tags := s.tagsPool[serial]
		k.AfterUnmarshal(tags)
	}
	s.tagsPool = make(map[uo.Serial]*marshal.TagCollection)
}
