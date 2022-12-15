package file

import (
	"log"
	"os"
)

// StaticMul represents a single generic MUL file that has no index file
type StaticMul struct {
	// Length of a single segment
	pitch int
	// Offset from the start of the file to start reading
	offset int
	// Data of the file
	d []byte
}

// NewStaticMulFromFile returns a new Mul object initialized with the data of
// the named file, or nil if an error was logged.
func NewStaticMulFromFile(fname string, pitch, offset int) *StaticMul {
	d, err := os.ReadFile(fname)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &StaticMul{
		pitch:  pitch,
		offset: offset,
		d:      d,
	}
}

// NumberOfSegments returns the number of segments in the MUL file
func (m *StaticMul) NumberOfSegments() int {
	return (len(m.d) - m.offset) / m.pitch
}

// GetSegment returns the raw data of the segment, or nil if the segment index
// was out of range.
func (m *StaticMul) GetSegment(index int) []byte {
	if index < 0 || index >= m.NumberOfSegments() {
		return nil
	}
	start := index*m.pitch + m.offset
	end := start + m.pitch
	if end > len(m.d) {
		return nil
	}
	return m.d[start:end]
}
