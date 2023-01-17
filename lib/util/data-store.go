package util

import (
	"fmt"
	"io"
	"log"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
)

type dsobj interface {
	Serializeable
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
	// Factory to use to create new objects during the indexing phase
	f *SerializeableFactory
	// Debug name of the data store
	name string
}

// NewDataStore initializes and returns a new DataStore object.
func NewDataStore[K dsobj](name string, rng uo.RandomSource, f *SerializeableFactory) *DataStore[K] {
	return &DataStore[K]{
		SerialPool: *NewSerialPool[K](name, rng),
		tfoPool:    make(map[uo.Serial]*TagFileObject),
		toPool:     make(map[uo.Serial]*marshal.TagObject),
		f:          f,
		name:       name,
	}
}

// Data returns the underlying data store.
func (s *DataStore[K]) Data() map[uo.Serial]K {
	return s.SerialPool.objects
}

// SetData throws away the underlying data store and replaces it.
func (s *DataStore[K]) SetData(data map[uo.Serial]K) {
	s.SerialPool.SetData(data)
}

// Read reads the data store from the tag file. This creates the objects of the
// data store but does not deserialize them. Deserialization should be done
// after all data stores are loaded so pointers can be resolved. This has the
// side-effect of populating the save data cache. This cache can be cleared with
// InvalidateCache or as a side-effect of calling Deserialize.
func (s *DataStore[K]) Read(r io.Reader) []error {
	var errs []error
	tfr := &TagFileReader{}
	tfr.StartReading(r)
	for {
		tfo := tfr.ReadObject()
		if tfo == nil {
			// No more objects in the file
			return append(errs, tfr.Errors()...)
		}
		if tfo.HasErrors() {
			errs = append(errs, tfr.Errors()...)
			continue
		}
		o := s.f.New(tfo.TypeName(), nil)
		if o == nil {
			errs = append(errs, fmt.Errorf("failed to create object of type %s", tfo.TypeName()))
			continue
		} else {
			serial := uo.Serial(tfo.GetNumber("Serial", int(uo.SerialZero)))
			if serial == uo.SerialZero {
				errs = append(errs, fmt.Errorf("object of type %s did not contain a Serial", tfo.TypeName()))
				continue
			}
			s.objects[serial] = o.(K)
			s.tfoPool[serial] = tfo
		}
	}
}

// InvalidateCache invalidates the object deserialization cache. This should be
// done to conserve memory if Deserialize is not called.
func (s *DataStore[K]) InvalidateCache() {
	s.tfoPool = make(map[uo.Serial]*TagFileObject)
}

// Deserialize executes the Deserialize method on every object in the data store
// in a non-deterministic order. This should be called after every data store
// has been loaded with a Read call.
func (s *DataStore[K]) Deserialize() []error {
	var errs []error
	for k, o := range s.objects {
		tfo := s.tfoPool[k]
		o.Deserialize(tfo)
		errs = append(errs, tfo.Errors()...)
	}
	return errs
}

// OnAfterDeserialize executes the OnAfterDeserialize method on every object in
// the data store in a non-deterministic order. This should be called after
// every data store has been deserialized with a Deserialize call. As a side
// effect this function calls InvalidateCache to free memory that is no longer
// needed.
func (s *DataStore[K]) OnAfterDeserialize() []error {
	var errs []error
	objs := make(map[uo.Serial]Serializeable, len(s.objects))
	for k, o := range s.objects {
		objs[k] = o
	}
	for k, o := range objs {
		tfo := s.tfoPool[k]
		if tfo == nil {
			errs = append(errs, fmt.Errorf("object %s cached TFO not found during DataStore.OnAfterDeserialize()", o.Serial().String()))
			continue
		}
		o.OnAfterDeserialize(tfo)
		errs = append(errs, tfo.Errors()...)
	}
	s.InvalidateCache()
	return errs
}

// Write writes the contents to the writer
func (s *DataStore[K]) Write(w io.WriteCloser) []error {
	tfw := NewTagFileWriter(w)
	tfw.WriteComment(s.name)
	tfw.WriteBlankLine()
	for _, o := range s.objects {
		tfw.WriteObject(o)
		tfw.WriteBlankLine()
	}
	tfw.WriteComment("end of file")
	tfw.Close()
	return nil
}

// Marshal marshals all of the object belonging to a given serial pool to a
// segment.
func (s *DataStore[K]) Marshal(seg *marshal.TagFileSegment, pools, pool int) {
	for serial, k := range s.objects {
		if int(serial)%pools == pool {
			k.Marshal(seg)
			seg.PutTag(marshal.TagEndOfList, marshal.TagValueBool, true)
			seg.IncrementRecordCount()
		}
	}
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
		log.Printf("%s:%+v", to.Serial.String(), to)
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
