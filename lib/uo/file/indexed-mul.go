package file

import (
	"log"
	"os"
)

// IndexedMul
type IndexedMul struct {
	// Index mul for our data
	index *IndexMul
	// The data of the file
	d []byte
}

// NewIndexedMulFromFile returns a new IndexedMul object initialized with the
// data of the named files, or nil if an error was logged.
func NewIndexedMulFromFile(indexFileName, mulFileName string) *IndexedMul {
	m := &IndexedMul{
		index: NewIndexFrom(indexFileName),
	}
	d, err := os.ReadFile(mulFileName)
	if err != nil {
		log.Println(err)
		return nil
	}
	m.d = d
	return m
}

// NumberOfSegments returns the number of segments in the MUL file
func (m *IndexedMul) NumberOfSegments() int {
	return m.index.NumberOfEntries()
}

// GetSegment returns the raw data of the segment, or nil if the segment index
// was out of range.
func (m *IndexedMul) GetSegment(index int) []byte {
	if index < 0 || index >= m.NumberOfSegments() {
		return nil
	}
	seg := m.index.GetEntry(index)
	if seg == nil {
		return nil
	}
	start := seg.Offset
	end := start + seg.Length
	if end > len(m.d) {
		return nil
	}
	return m.d[start:end]
}
