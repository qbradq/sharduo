package marshal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

// TagFile manages an object database in-memory.
//
// Encoding formats:
//
// A tag file consist of a series of segments. Each segment is treated as a
// binary blob within the file structure. It's exact encoding is determined by
// the application at run-time.
//
// File Format:
// MagicString            uint64          A fixed value used to identify this as a ShardUO TagFile: 0x6BB50D00B87E33A4
// SegmentCount           uint8           Number of segments in the file
// Headers                []SegmentHeader Segment headers
// Blob                   []byte          Raw segment data
//
// Segment Header:
// ID                     uint8           Unique identifier of the segment
// Offset                 uint64          Offset from the beginning of the file where the raw segment data starts
// Length                 uint64          Length of the raw segment data
// Records                uint32          Number of records in the segment
type TagFile struct {
	segs []*TagFileSegment // Segments
}

// NewTagFile creates a new TagFile object from the data given. The data slice
// may be nil, in which case an empty TagFile object is returned ready for
// write operations.
func NewTagFile(d []byte) *TagFile {
	t := &TagFile{}
	if d == nil {
		return t
	}
	// Header data
	magic := binary.LittleEndian.Uint64(d[0:8])
	if magic != 0x6BB50D00B87E33A4 {
		panic("tag file does not have correct magic number")
	}
	nSegments := int(d[8])
	// Load segments
	ofs := 9
	t.segs = make([]*TagFileSegment, nSegments)
	for i := range t.segs {
		s := NewTagFileSegment(0, t)
		t.segs[i] = s
		s.id = Segment(d[ofs])
		segmentOffset := binary.LittleEndian.Uint64(d[ofs+1 : ofs+9])
		segmentLength := binary.LittleEndian.Uint64(d[ofs+9 : ofs+17])
		s.records = binary.LittleEndian.Uint32(d[ofs+17 : ofs+21])
		s.buf = bytes.NewBuffer(d[segmentOffset : segmentOffset+segmentLength])
		ofs += 21
	}
	return t
}

// Close releases internal memory and MUST be called!
func (f *TagFile) Close() {
	for idx, s := range f.segs {
		s.parent = nil
		f.segs[idx] = nil
	}
}

// Output writes the file to the writer in TagFile format.
func (f *TagFile) Output(w io.Writer) {
	buf := make([]byte, 21)
	// Write file header
	binary.LittleEndian.PutUint64(buf[0:8], 0x6BB50D00B87E33A4) // Magic string
	w.Write(buf[0:8])
	buf[0] = byte(len(f.segs))
	w.Write(buf[0:1])
	// Write segment headers
	var ofs uint64 = 9 + 21*(uint64(len(f.segs)))
	// Output segments
	for _, seg := range f.segs {
		buf[0] = byte(seg.id)                                           // Segment ID
		binary.LittleEndian.PutUint64(buf[1:9], ofs)                    // File offset
		binary.LittleEndian.PutUint64(buf[9:17], uint64(seg.buf.Len())) // Length
		binary.LittleEndian.PutUint32(buf[17:21], seg.records)          // Record count
		w.Write(buf[0:21])
		ofs += uint64(seg.buf.Len())
	}
	// Segment data
	for _, seg := range f.segs {
		w.Write(seg.buf.Bytes())
	}
}

// Segment returns the named segment. A new, empty segment is created if needed.
func (f *TagFile) Segment(which Segment) *TagFileSegment {
	for _, s := range f.segs {
		if s.id == which {
			return s
		}
	}
	s := NewTagFileSegment(which, f)
	f.segs = append(f.segs, s)
	return s
}

// TagFileSegment manages a single segment in the file.
type TagFileSegment struct {
	parent  *TagFile      // Parent tag file
	id      Segment       // Segment ID
	buf     *bytes.Buffer // Raw data buffer
	tbuf    []byte        // Temporary buffer
	records uint32        // Number of records in this segment
}

// NewTagFileSegment returns an initialized TagFile object.
func NewTagFileSegment(id Segment, parent *TagFile) *TagFileSegment {
	return &TagFileSegment{
		parent: parent,
		id:     id,
		buf:    &bytes.Buffer{},
		tbuf:   make([]byte, 4*256+3),
	}
}

// IsEmpty returns true if the segment contains no data.
func (s *TagFileSegment) IsEmpty() bool { return s.buf.Len() == 0 }

// IncrementRecordCount increments the record count of the segment by one.
func (s *TagFileSegment) IncrementRecordCount() { s.records++ }

// RecordCount returns the number of records encoded into the segment.
func (s *TagFileSegment) RecordCount() uint32 { return s.records }

// PutBool writes a single 8-bit value to the segment based on the value of v.
func (s *TagFileSegment) PutBool(v bool) {
	if v {
		s.buf.WriteByte(1)
	} else {
		s.buf.WriteByte(0)
	}
}

// PutByte writes a single 8-bit value to the segment.
func (s *TagFileSegment) PutByte(v byte) {
	s.buf.WriteByte(v)
}

// PutShort writes a single 16-bit value to the segment.
func (s *TagFileSegment) PutShort(v uint16) {
	binary.LittleEndian.PutUint16(s.tbuf[0:2], v)
	s.buf.Write(s.tbuf[0:2])
}

// PutInt writes a single 32-bit value to the segment.
func (s *TagFileSegment) PutInt(v uint32) {
	binary.LittleEndian.PutUint32(s.tbuf[0:4], v)
	s.buf.Write(s.tbuf[0:4])
}

// PutLong writes a single 64-bit value to the segment.
func (s *TagFileSegment) PutLong(v uint64) {
	binary.LittleEndian.PutUint64(s.tbuf[0:8], v)
	s.buf.Write(s.tbuf[0:8])
}

// PutFloat writes a single 64-bit float value to the segment.
func (s *TagFileSegment) PutFloat(v float64) {
	binary.LittleEndian.PutUint64(s.tbuf[0:8], math.Float64bits(v))
	s.buf.Write(s.tbuf[0:8])
}

// PutString writes a 32-bit string reference value to the segment, and inserts
// the string into the dictionary if needed.
func (s *TagFileSegment) PutString(v string) {
	s.buf.WriteString(v)
	s.buf.WriteByte(0)
}

// PutStringsMap writes a sequential list of strings from the map in
// key, value pairs preceded by the count of pairs (16-bit value).
func (s *TagFileSegment) PutStringsMap(m map[string]string) {
	l := len(m)
	if l > 255 {
		panic("map too big for PutStringsMap")
	}
	s.buf.WriteByte(byte(l))
	for k, v := range m {
		s.buf.WriteString(k)
		s.buf.WriteByte(0)
		s.buf.WriteString(v)
		s.buf.WriteByte(0)
	}
}

// PutObjectReferences writes all of the serials from the given slice as 32-bit
// object references to the tag file segment preceded by the count of references
// (8-bit value).
func (s *TagFileSegment) PutObjectReferences(serials []uo.Serial) {
	l := len(serials)
	if l > 255 {
		panic("slice too big for PutObjectReferences")
	}
	s.tbuf[0] = byte(l)
	ofs := 1
	for _, serial := range serials {
		binary.LittleEndian.PutUint32(s.tbuf[ofs+0:ofs+4], uint32(serial))
		ofs += 4
	}
	s.buf.Write(s.tbuf[:ofs])
}

// PutShortSlice writes all of the values in the slice to the segment preceded
// by the number of shorts (8-bit value).
func (s *TagFileSegment) PutShortSlice(v []int16) {
	l := len(v)
	if l > 255 {
		panic("slice too big for PutShortSlice")
	}
	s.tbuf[0] = byte(l)
	ofs := 1
	for _, v := range v {
		binary.LittleEndian.PutUint16(s.tbuf[ofs+0:ofs+2], uint16(v))
		ofs += 2
	}
	s.buf.Write(s.tbuf[:ofs])
}

// PutSerialSlice writes all of the values in the slice to the segment preceded
// by the number of shorts (8-bit value).
func (s *TagFileSegment) PutSerialSlice(v []uo.Serial) {
	l := len(v)
	if l > 255 {
		panic("slice too big for PutShortSlice")
	}
	s.tbuf[0] = byte(l)
	ofs := 1
	for _, v := range v {
		binary.LittleEndian.PutUint32(s.tbuf[ofs+0:ofs+4], uint32(v))
		ofs += 4
	}
	s.buf.Write(s.tbuf[:ofs])
}

// PutLocation writes a location value to the segment as a tuple of
// int16,int16,int8.
func (s *TagFileSegment) PutLocation(l uo.Location) {
	binary.LittleEndian.PutUint16(s.tbuf[0:2], uint16(l.X))
	binary.LittleEndian.PutUint16(s.tbuf[2:4], uint16(l.Y))
	s.tbuf[4] = byte(int8(l.Z))
	s.buf.Write(s.tbuf[0:5])
}

// PutBounds writes a bounds value to the segment as a tuple of
// int16,int16,int16,int16
func (s *TagFileSegment) PutBounds(b uo.Bounds) {
	binary.LittleEndian.PutUint16(s.tbuf[0:2], uint16(b.X))
	binary.LittleEndian.PutUint16(s.tbuf[2:4], uint16(b.Y))
	s.tbuf[4] = byte(int8(b.Z))
	binary.LittleEndian.PutUint16(s.tbuf[5:7], uint16(b.W))
	binary.LittleEndian.PutUint16(s.tbuf[7:9], uint16(b.H))
	binary.LittleEndian.PutUint16(s.tbuf[9:11], uint16(b.D))
	s.buf.Write(s.tbuf[0:11])
}

// PutObject writes an object to the segment.
func (s *TagFileSegment) PutObject(o Marshaler) {
	s.PutByte(byte(o.ObjectType()))
	s.PutInt(uint32(o.Serial()))
	s.PutString(o.TemplateName())
	o.Marshal(s)
}

// Bool returns the next 8-bit number as a boolean, any value other than 0 is
// returned as true
func (s *TagFileSegment) Bool() bool {
	s.buf.Read(s.tbuf[:1])
	return s.tbuf[0] != 0
}

// Byte returns the next 8-bit number in the segment
func (s *TagFileSegment) Byte() byte {
	s.buf.Read(s.tbuf[:1])
	return s.tbuf[0]
}

// Short returns the next 16-bit number in the segment
func (s *TagFileSegment) Short() uint16 {
	s.buf.Read(s.tbuf[:2])
	return binary.LittleEndian.Uint16(s.tbuf[:2])
}

// Int returns the next 32-bit number in the segment
func (s *TagFileSegment) Int() uint32 {
	s.buf.Read(s.tbuf[:4])
	return binary.LittleEndian.Uint32(s.tbuf[:4])
}

// Long returns the next 64-bit number in the segment
func (s *TagFileSegment) Long() uint64 {
	s.buf.Read(s.tbuf[:8])
	return binary.LittleEndian.Uint64(s.tbuf[:8])
}

// Float returns the next 64-bit float in the segment
func (s *TagFileSegment) Float() float64 {
	s.buf.Read(s.tbuf[:8])
	return math.Float64frombits(binary.LittleEndian.Uint64(s.tbuf[:8]))
}

// String returns the next UTF-8 string in the segment by reading a 32-bit
// string reference and indexing the tag file string dictionary.
func (s *TagFileSegment) String() string {
	d, err := s.buf.ReadString(0)
	if len(d) > 0 {
		d = d[:len(d)-1]
	}
	if err != nil || len(d) == 0 {
		return ""
	}
	return string(d)
}

// StringMap returns the next map[string]string encoded into the segment.
func (s *TagFileSegment) StringMap() map[string]string {
	ret := make(map[string]string)
	count, _ := s.buf.ReadByte()
	for i := 0; i < int(count); i++ {
		k := s.String()
		v := s.String()
		if k == "" || v == "" {
			continue
		}
		ret[k] = v
	}
	return ret
}

// ObjectReferences returns the next []uo.Serial encoded into the segment.
func (s *TagFileSegment) ObjectReferences() []uo.Serial {
	d := s.buf.Bytes()
	var ret []uo.Serial
	count := int(d[0])
	ofs := 1
	for i := 0; i < count; i++ {
		ref := binary.LittleEndian.Uint32(d[ofs+0 : ofs+4])
		ofs += 4
		ret = append(ret, uo.Serial(ref))
	}
	s.buf.Next(ofs)
	return ret
}

// ShortSlice returns the next []int16 encoded into the segment.
func (s *TagFileSegment) ShortSlice() []int16 {
	d := s.buf.Bytes()
	var ret []int16
	count := int(d[0])
	ofs := 1
	for i := 0; i < count; i++ {
		v := int16(binary.LittleEndian.Uint16(d[ofs+0 : ofs+2]))
		ofs += 2
		ret = append(ret, v)
	}
	s.buf.Next(ofs)
	return ret
}

// SerialSlice returns the next []uo.Serial encoded into the segment.
func (s *TagFileSegment) SerialSlice() []uo.Serial {
	d := s.buf.Bytes()
	var ret []uo.Serial
	count := int(d[0])
	ofs := 1
	for i := 0; i < count; i++ {
		v := uo.Serial(binary.LittleEndian.Uint32(d[ofs+0 : ofs+4]))
		ofs += 4
		ret = append(ret, v)
	}
	s.buf.Next(ofs)
	return ret
}

// Location returns the next uo.Location value encoded into the segment.
func (s *TagFileSegment) Location() uo.Location {
	s.buf.Read(s.tbuf[0:5])
	return uo.Location{
		X: int16(binary.LittleEndian.Uint16(s.tbuf[0:2])),
		Y: int16(binary.LittleEndian.Uint16(s.tbuf[2:4])),
		Z: int8(s.tbuf[4]),
	}
}

// Bounds returns the next uo.Bounds value encoded into the segment.
func (s *TagFileSegment) Bounds() uo.Bounds {
	s.buf.Read(s.tbuf[0:11])
	return uo.Bounds{
		X: int16(binary.LittleEndian.Uint16(s.tbuf[0:2])),
		Y: int16(binary.LittleEndian.Uint16(s.tbuf[2:4])),
		Z: int8(s.tbuf[4]),
		W: int16(binary.LittleEndian.Uint16(s.tbuf[5:7])),
		H: int16(binary.LittleEndian.Uint16(s.tbuf[7:9])),
		D: int16(binary.LittleEndian.Uint16(s.tbuf[9:11])),
	}
}

// Object returns the next object encoded into the segment which must support
// the Unmarshaler interface. The object will be fully unmarshaled upon return.
func (s *TagFileSegment) Object() Unmarshaler {
	// Object construction
	ot := ObjectType(s.Byte())
	ctor := Constructor(ot)
	if ctor == nil {
		panic(fmt.Sprintf("no constructor found for object type code 0x%02X", ot))
	}
	o := ctor()
	um, ok := o.(Unmarshaler)
	if !ok {
		panic(fmt.Sprintf("object for type code 0x%02X does not implement Unmarshaler", ot))
	}
	// Self-serial and dataset insertion
	ss := uo.Serial(s.Int())
	um.SetSerial(ss)
	insertFunction(um)
	// Template deserialization handling
	tn := s.String()
	t := template.FindTemplate(tn)
	um.Deserialize(t, false)
	// Data unmarshaling
	um.Unmarshal(s)
	return um
}
