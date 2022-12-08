package util

import (
	"errors"
	"io"

	"github.com/qbradq/sharduo/lib/uo"
)

// DataStore is a file-backed key-value store.
type DataStore[K Serializeable] struct {
	SerializeablePool
	f *SerializeableFactory
}

// NewDataStore initializes and returns a new DataStore object.
func NewDataStore[K Serializeable](name string, rng uo.RandomSource, f *SerializeableFactory) *DataStore[K] {
	return &DataStore[K]{
		SerializeablePool: *NewSerializeablePool(name, rng),
		f:                 f,
	}
}

// Read reads the data store from the tag file.
func (s *DataStore[K]) Read(r io.Reader) []error {
	f := NewTagFileReader(r, s.f)
	for {
		o, err := f.ReadObject()
		if errors.Is(err, io.EOF) {
			return f.Errors()
		} else if o != nil {
			s.Insert(o)
		}
	}
}

// Write writes the contents to the writer
func (p *DataStore[K]) Write(w io.Writer) []error {
	tfw := NewTagFileWriter(w)
	tfw.WriteCommentLine(p.name)
	tfw.WriteBlankLine()
	for _, o := range p.objects {
		o.Serialize(tfw)
		tfw.WriteBlankLine()
	}
	return tfw.Errors()
}

// Get returns the named object or nil.
func (s *DataStore[K]) Get(which uo.Serial) K {
	return s.SerializeablePool.Get(which).(K)
}
