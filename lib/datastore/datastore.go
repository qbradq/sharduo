package datastore

import (
	"log"
	"sync"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

type dsobj interface {
	Serial() uo.Serial
	SetSerial(uo.Serial)
	SerialType() uo.SerialType
	Deserialize(*template.T, bool)
	RecalculateStats()
	TemplateName() string
	SetTemplateName(string)
	marshal.Marshaler
	marshal.Unmarshaler
}

// T is a file-backed key-value store.
type T[K dsobj] struct {
	// Pool of managed objects
	objects map[uo.Serial]K
	// Pool of deserialization data for rebuilding the objects
	tagsPool map[uo.Serial]*marshal.TagCollection
	// Random number source for serials
	rng uo.RandomSource
}

// NewDataStore initializes and returns a new DataStore object.
func NewDataStore[K dsobj](rng uo.RandomSource) *T[K] {
	return &T[K]{
		objects:  make(map[uo.Serial]K),
		tagsPool: make(map[uo.Serial]*marshal.TagCollection),
		rng:      rng,
	}
}

// Add adds a new object to the datastore assigning it a new serial of the
// correct type.
func (s *T[K]) Add(k K, t uo.SerialType) {
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
			k.SetSerial(serial)
			s.objects[serial] = k
			break
		}
	}
}

// Remove blindly removes the object from the datastore that is indexed by this
// object's serial.
func (s *T[K]) Remove(o dsobj) {
	var zero K
	s.objects[o.Serial()] = zero
	delete(s.objects, o.Serial())
}

// Insert inserts the object into the datastore with its current serial and will
// overwrite existing values without warning. This is typically only used when
// rebuilding the dataset from an external data source.
func (s *T[K]) Insert(k K) {
	s.objects[k.Serial()] = k
}

// Get returns the identified object or nil.
func (s *T[K]) Get(serial uo.Serial) K {
	return s.objects[serial]
}

// Data returns the underlying data store.
func (s *T[K]) Data() map[uo.Serial]K {
	return s.objects
}

// MarshalObjects marshals objects to raw data.
func (s *T[K]) MarshalObjects(tf *marshal.TagFile, goroutines int, wg *sync.WaitGroup) {
	// We have to build a slice of all of our objects so we don't have
	// concurrency issues on the data store map during the multi-goroutine save
	objects := make([]K, len(s.objects))
	idx := 0
	for _, obj := range s.objects {
		objects[idx] = obj
		idx++
	}
	for i := 0; i < goroutines; i++ {
		// Object data
		seg := tf.Segment(marshal.SegmentObjectsStart + marshal.Segment(i))
		wg.Add(1)
		go func(seg *marshal.TagFileSegment, pool int) {
			defer wg.Done()
			for i := pool; i < len(objects); i += goroutines {
				o := objects[i]
				seg.PutInt(uint32(o.Serial()))
				seg.PutString(o.TemplateName())
				o.Marshal(seg)
				// We have to terminate the tag list outside of Marshal() due to
				// how the unmarshaling chain works.
				seg.PutTag(marshal.TagEndOfList, marshal.TagValueBool, true)
				seg.IncrementRecordCount()
			}
		}(seg, i)
	}
}

// UnmarshalObjects unmarshals objects from raw data. AfterUnmarshalObjects must
// be called after this to complete the load process and free internal memory.
// True is returned on success.
func (s *T[K]) UnmarshalObjects(seg *marshal.TagFileSegment) bool {
	for i := uint32(0); i < seg.RecordCount(); i++ {
		// Grab the object's serial
		serial := uo.Serial(seg.Int())
		k, ok := s.objects[serial]
		if !ok {
			log.Printf("failed to find object %s", serial.String())
			return false
		}
		// Load the template so we can deserialize the default and static values
		tn := seg.String()
		t := template.FindTemplate(tn)
		if t == nil {
			// The error is logged inside FindTemplate
			return false
		}
		k.Deserialize(t, false)
		k.RecalculateStats()
		tags := k.Unmarshal(seg)
		s.tagsPool[serial] = tags
	}
	return true
}

// AfterUnmarshalObjects executes the AfterUnmarshal function for all objects in
// the datastore. UnmarshalObjects must be called first.
func (s *T[K]) AfterUnmarshalObjects() {
	for serial, k := range s.objects {
		tags := s.tagsPool[serial]
		k.AfterUnmarshal(tags)
	}
	s.tagsPool = make(map[uo.Serial]*marshal.TagCollection)
}
