package marshal

import (
	"bytes"
	"encoding/binary"
	"io"
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
// ID     uint8  Unique identifier of the segment
// Offset uint64 Offset from the beginning of the file where the raw segment data starts
// Length uint64 Length of the raw segment data
//
// StringDictionarySegment:
// A tag file will always have a string dictionary segment with segment ID
// SegmentStringDictionary. It is used to look up all string references in all
// TagObject segments in the file. Format as follows:
// StringCount            uint32          Number of strings in the segment
// Strings                []string        Array of null-terminated UTF-8 strings
//
// TagObjectSegment:
// ObjectCount            uint32          Number of objects in the segment
// Objects                []TagObject     Object data
//
// TagObject:
// Serial                 uint32          uo.Serial of the object
// ConcreteType           uint8           A value indicating which Go struct this object will deserialize into
// TemplateName           uint32          String reference of the template name this object was created from
// Parent                 uint32          uo.Serial of the object's parent, uo.SerialSystem for the map, or uo.SerialZero for void (leaked)
// Name                   uint32          String reference of the base name string of the object
// Hue                    uint16          uo.Hue of the object
// Location     int16,int16,int8          uo.Location of the object
// TagCount               uint8           Number of dynamic tags in the object
// DynamicTags            []TagValue      The rest of the data for the object as TagValue structures
//
// TagValue:
// Tag                    uint8           Tag code. The width and type of the following data is determined by a lookup table keyed on this value
// Data                   []byte          Raw data
type TagFile struct {
	dict map[string]uint32 // Map of strings used in the file used to generate the dictionary segment
	segs []*TagFileSegment // Segments
}

// NewTagFile creates a new TagFile object from the data given. The data slice
// may be nil, in which case an empty TagFile object is returned ready for
// write operations.
func NewTagFile(d []byte) *TagFile {
	buf := make([]byte, 4)
	t := &TagFile{
		dict: make(map[string]uint32),
	}
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
		s := &TagFileSegment{}
		t.segs[i] = s
		s.id = Segment(d[ofs])
		segmentOffset := binary.LittleEndian.Uint64(d[ofs+1 : ofs+9])
		segmentLength := binary.LittleEndian.Uint64(d[ofs+9 : ofs+17])
		s.buf = bytes.NewBuffer(d[segmentOffset : segmentOffset+segmentLength])
		ofs += 17
	}
	// Pre-load the string dictionary
	var sdseg *TagFileSegment
	for _, s := range t.segs {
		if s.id == SegmentStringDictionary {
			sdseg = s
		}
	}
	if sdseg == nil {
		panic("string dictionary segment not found")
	}
	sdd := bytes.NewBuffer(sdseg.buf.Bytes())
	sdd.Read(buf[0:4])
	dictLength := binary.LittleEndian.Uint32(buf[0:4])
	for i := uint32(0); i < dictLength; i++ {
		sdd.Read(buf[0:4])
		sid := binary.LittleEndian.Uint32(buf[0:4])
		stringData, _ := sdd.ReadBytes(0)
		s := string(stringData[:len(stringData)-1])
		t.dict[s] = sid
	}
	return t
}

// WriteTo writes the file to the writer in TagFile format.
func (f *TagFile) WriteTo(w io.Writer) {
	buf := make([]byte, 17)
	// First we have to generate the raw data for the dictionary segment so we
	// know how long it will be.
	dictBuf := &bytes.Buffer{}
	binary.LittleEndian.PutUint32(buf[0:4], uint32(len(f.dict))) // String count
	dictBuf.Write(buf[0:4])
	for s, id := range f.dict {
		binary.LittleEndian.PutUint32(buf[0:4], id) // String ID - NOT SEQUENTIAL ORDER!
		dictBuf.Write(buf[0:4])
		dictBuf.WriteString(s) // Null-terminated string
		dictBuf.WriteByte(0)
	}
	// Write file header
	binary.LittleEndian.PutUint64(buf[0:8], 0x6BB50D00B87E33A4) // Magic string
	w.Write(buf[0:8])
	buf[0] = byte(len(f.segs) + 1) // +1 is for the dictionary segment
	w.Write(buf[0:1])
	// Write segment headers
	var ofs uint64 = 9 + 17*(uint64(len(f.segs))+1)
	// Dictionary segment header
	buf[0] = byte(SegmentStringDictionary)                          // Segment ID
	binary.LittleEndian.PutUint64(buf[1:9], ofs)                    // File offset
	binary.LittleEndian.PutUint64(buf[9:17], uint64(dictBuf.Len())) // Length
	w.Write(buf[0:17])
	ofs += uint64(dictBuf.Len())
	// Other segments
	for _, seg := range f.segs {
		buf[0] = byte(seg.id)                                           // Segment ID
		binary.LittleEndian.PutUint64(buf[1:9], ofs)                    // File offset
		binary.LittleEndian.PutUint64(buf[9:17], uint64(seg.buf.Len())) // Length
		w.Write(buf[0:17])
		ofs += uint64(seg.buf.Len())
	}
	// Segment data
	w.Write(dictBuf.Bytes())
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
	s := &TagFileSegment{
		id:  which,
		buf: &bytes.Buffer{},
	}
	return s
}

// TagFileSegment manages a single segment in the file.
type TagFileSegment struct {
	id  Segment       // Segment ID
	buf *bytes.Buffer // Raw data buffer
}
