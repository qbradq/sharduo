package datastore

import (
	"log"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

type dsobj interface {
	Serial() uo.Serial
	SetSerial(uo.Serial)
	SerialType() uo.SerialType
	Deserialize(*template.Template, bool)
	RecalculateStats()
	TemplateName() string
	SetTemplateName(string)
	Removed() bool
	Remove()
	NoRent() bool
	marshal.Marshaler
	marshal.Unmarshaler
}

// T is a file-backed key-value store.
type T[K dsobj] struct {
	// Pool of managed objects
	objects map[uo.Serial]K
	// Random number source for serials
	rng uo.RandomSource
}

// NewDataStore initializes and returns a new DataStore object.
func NewDataStore[K dsobj](rng uo.RandomSource) *T[K] {
	return &T[K]{
		objects: make(map[uo.Serial]K),
		rng:     rng,
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
	o.Remove()
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
