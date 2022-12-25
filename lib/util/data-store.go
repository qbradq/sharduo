package util

import (
	"fmt"
	"io"

	"github.com/qbradq/sharduo/lib/uo"
)

// DataStore is a file-backed key-value store.
type DataStore[K Serializeable] struct {
	SerialPool[K]
	// Pool of deserialization data for rebuilding the objects
	tfoPool map[uo.Serial]*TagFileObject
	// Factory to use to create new objects during the indexing phase
	f *SerializeableFactory
	// Debug name of the data store
	name string
}

// NewDataStore initializes and returns a new DataStore object.
func NewDataStore[K Serializeable](name string, rng uo.RandomSource, f *SerializeableFactory) *DataStore[K] {
	return &DataStore[K]{
		SerialPool: *NewSerialPool[K](name, rng),
		tfoPool:    make(map[uo.Serial]*TagFileObject),
		f:          f,
		name:       name,
	}
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
// has been loaded with a Read call. As a side effect this function calls
// InvalidateCache to free memory that is no longer needed.
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
	for k, o := range s.objects {
		tfo := s.tfoPool[k]
		o.OnAfterDeserialize(tfo)
		errs = append(errs, tfo.Errors()...)
	}
	s.InvalidateCache()
	return errs
}

// Write writes the contents to the writer
func (p *DataStore[K]) Write(w io.WriteCloser) []error {
	tfw := NewTagFileWriter(w)
	tfw.WriteComment(p.name)
	tfw.WriteBlankLine()
	for _, o := range p.objects {
		tfw.WriteObject(o)
		tfw.WriteBlankLine()
	}
	tfw.WriteComment("end of file")
	tfw.Close()
	return nil
}
