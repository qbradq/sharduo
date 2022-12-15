package file

import (
	"github.com/qbradq/sharduo/lib/dataconv"
)

// SegmentIndex represents one index in the file
type SegmentIndex struct {
	// Offset into the MUL file of the segment
	Offset int
	// Length of the segment
	Length int
	// Extra data for this segment
	Extra int
}

// Index represents a MUL file index. These files typically have a similar name
// as the MUL file they apply to with idx somewhere in the file name.
type IndexMul struct {
	indexes []SegmentIndex
}

// NewIndexFrom returns a new Index object from the named file
func NewIndexFrom(idxp string) *IndexMul {
	m := NewStaticMulFromFile(idxp, 12, 0)
	if m == nil {
		return nil
	}
	// Allocate storage
	idx := &IndexMul{
		indexes: make([]SegmentIndex, m.NumberOfSegments()),
	}
	// Read in values
	for i := range idx.indexes {
		d := m.GetSegment(i)
		idx.indexes[i] = SegmentIndex{
			Offset: int(dataconv.GetUint32(d[0:4])),
			Length: int(dataconv.GetUint32(d[4:8])),
			Extra:  int(dataconv.GetUint32(d[8:12])),
		}
	}
	return idx
}
