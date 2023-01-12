package marshal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"

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
// ConcreteType           uint8           A value indicating which Go struct this object will deserialize into
// Serial                 uint32          uo.Serial of the object
// TemplateName           uint32          String reference of the template name this object was created from
// Parent                 uint32          uo.Serial of the object's parent, uo.SerialSystem for the map, or uo.SerialZero for void (leaked)
// Name                   uint32          String reference of the base name string of the object
// Hue                    uint16          uo.Hue of the object
// Location     int16,int16,int8          uo.Location of the object
// EventCount             uint8           Number of event definitions in the object
// Events                 []EventDef      Event definitions
// DynamicTags            []TagValue      The rest of the data for the object as TagValue structures
// TagsEnd
//
// EventDef:
// EventName              uint32          String reference to the event's name
// EventHandler           uint32          String reference to the event's handler name
//
// TagValue:
// Tag                    uint8           Tag code. The width and type of the following data is determined by a lookup table keyed on this value
// Data                   []byte          Raw data
type TagFile struct {
	dict    map[string]uint32 // Map of strings used in the file used to generate the dictionary segment
	backref map[uint32]string // Back reference for dict
	segs    []*TagFileSegment // Segments
	istr    uint32            // Next string ID
}

// NewTagFile creates a new TagFile object from the data given. The data slice
// may be nil, in which case an empty TagFile object is returned ready for
// write operations.
func NewTagFile(d []byte) *TagFile {
	buf := make([]byte, 4)
	t := &TagFile{
		dict:    make(map[string]uint32),
		backref: make(map[uint32]string),
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
		s := &TagFileSegment{
			parent: t,
			tbuf:   make([]byte, 4*2*256+1),
		}
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
		t.backref[sid] = s
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
	s := NewTagFileSegment(which)
	f.segs[which] = s
	return s
}

// StringReference returns a 32-bit reference value for the given string and
// inserts the string into the dictionary if needed.
func (f *TagFile) StringReference(v string) uint32 {
	ref, ok := f.dict[v]
	if ok {
		return ref
	}
	ref = f.istr
	f.istr += 1
	f.dict[v] = ref
	return ref
}

// StringByReference returns the string assigned the reference number, or the
// empty string.
func (f *TagFile) StringByReference(ref uint32) string {
	if s, ok := f.backref[ref]; ok {
		return s
	}
	return ""
}

// TagFileSegment manages a single segment in the file.
type TagFileSegment struct {
	parent *TagFile      // Parent tag file
	id     Segment       // Segment ID
	buf    *bytes.Buffer // Raw data buffer
	tbuf   []byte        // Temporary buffer
}

// NewTagFileSegment returns an initialized TagFile object.
func NewTagFileSegment(id Segment) *TagFileSegment {
	return &TagFileSegment{
		id:   id,
		tbuf: make([]byte, 4*2*256+1),
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

// PutString writes a 32-bit string reference value to the segment, and inserts
// the string into the dictionary if needed.
func (s *TagFileSegment) PutString(v string) {
	ref := s.parent.StringReference(v)
	binary.LittleEndian.PutUint32(s.tbuf[0:4], ref)
	s.buf.Write(s.tbuf[0:4])
}

// PutStringsMap writes a sequential list of strings from the map in
// key, value pairs preceded by the count of pairs (16-bit value).
func (s *TagFileSegment) PutStringsMap(m map[string]string) {
	l := len(m)
	if l > 255 {
		panic("map too big for PutStringsMap")
	}
	s.tbuf[0] = byte(l)
	ofs := 1
	for k, v := range m {
		kref := s.parent.StringReference(k)
		vref := s.parent.StringReference(v)
		binary.LittleEndian.PutUint32(s.tbuf[ofs+0:ofs+4], kref)
		binary.LittleEndian.PutUint32(s.tbuf[ofs+4:ofs+8], vref)
		ofs += 8
	}
	s.buf.Write(s.tbuf[:ofs])
}

// PutObjectReferences writes all of the keys from the given map as 32-bit
// object references to the tag file segment preceded by the count of references
// (8-bit value). This is hacky but avoids copying map keys into slices for
// every single object in the game during save.
func PutObjectReferences[I any](s *TagFileSegment, m map[uo.Serial]I) {
	l := len(m)
	if l > 255 {
		panic("map too big for PutObjectReferences")
	}
	s.tbuf[0] = byte(l)
	ofs := 1
	for k := range m {
		binary.LittleEndian.PutUint32(s.tbuf[ofs+0:ofs+4], uint32(k))
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
	s.tbuf[9] = byte(int8(b.Z))
	s.buf.Write(s.tbuf[0:10])
}

// PutObjectHeader writes an object header to the segment. See documentation for
// TagFile for format description.
func (s *TagFileSegment) PutObjectHeader(
	otype ObjectType,
	serial uo.Serial,
	template string,
	parent uo.Serial,
	name string,
	hue uo.Hue,
	location uo.Location,
	events map[string]string) {
	s.tbuf[0] = byte(otype)
	binary.LittleEndian.PutUint32(s.tbuf[1:5], uint32(serial))
	sref := s.parent.StringReference(template)
	binary.LittleEndian.PutUint32(s.tbuf[5:9], sref)
	binary.LittleEndian.PutUint32(s.tbuf[9:13], uint32(parent))
	sref = s.parent.StringReference(name)
	binary.LittleEndian.PutUint32(s.tbuf[13:17], sref)
	binary.LittleEndian.PutUint16(s.tbuf[17:19], uint16(hue))
	binary.LittleEndian.PutUint16(s.tbuf[19:21], uint16(location.X))
	binary.LittleEndian.PutUint16(s.tbuf[21:23], uint16(location.Y))
	s.tbuf[23] = byte(int8(location.Z))
	s.buf.Write(s.tbuf[0:24])
	s.PutStringsMap(events)
}

// PutTag writes a tagged value to the segment. Please note that to write a
// reference slice please use PutByte(byte(tag)), PutObjectReferences[I].
func (s *TagFileSegment) PutTag(t Tag, value interface{}) {
	if int(t) >= len(tagValueMapping) {
		return
	}
	if t == TagEndOfList {
		s.tbuf[0] = byte(TagEndOfList)
		s.buf.Write(s.tbuf[0:1])
		return
	}
	valueType := tagValueMapping[t]
	switch v := value.(type) {
	case bool:
		if valueType != TagValueBool {
			log.Printf("warning: tag %d is not a bool", t)
			return
		}
		if v {
			s.PutByte(byte(t))
		}
	case uint8:
		if valueType != TagValueByte {
			log.Printf("warning: tag %d is not a byte", t)
			return
		}
		s.PutByte(byte(t))
		s.PutByte(v)
	case int8:
		if valueType != TagValueByte {
			log.Printf("warning: tag %d is not a byte", t)
			return
		}
		s.PutByte(byte(t))
		s.PutByte(byte(v))
	case uint16:
		if valueType != TagValueShort {
			log.Printf("warning: tag %d is not a short", t)
			return
		}
		s.PutByte(byte(t))
		s.PutShort(v)
	case int16:
		if valueType != TagValueShort {
			log.Printf("warning: tag %d is not a short", t)
			return
		}
		s.PutByte(byte(t))
		s.PutShort(uint16(v))
	case uint:
		if valueType != TagValueInt {
			log.Printf("warning: tag %d is not an integer", t)
			return
		}
		s.PutByte(byte(t))
		s.PutInt(uint32(v))
	case int:
		if valueType != TagValueInt {
			log.Printf("warning: tag %d is not an integer", t)
			return
		}
		s.PutByte(byte(t))
		s.PutInt(uint32(v))
	case uint32:
		if valueType != TagValueInt {
			log.Printf("warning: tag %d is not an integer", t)
			return
		}
		s.PutByte(byte(t))
		s.PutInt(uint32(v))
	case int32:
		if valueType != TagValueInt {
			log.Printf("warning: tag %d is not an integer", t)
			return
		}
		s.PutByte(byte(t))
		s.PutInt(uint32(v))
	case uint64:
		if valueType != TagValueLong {
			log.Printf("warning: tag %d is not a long", t)
			return
		}
		s.PutByte(byte(t))
		s.PutLong(v)
	case int64:
		if valueType != TagValueLong {
			log.Printf("warning: tag %d is not a long", t)
			return
		}
		s.PutByte(byte(t))
		s.PutLong(uint64(v))
	case string:
		if valueType != TagValueString {
			log.Printf("warning: tag %d is not a string", t)
			return
		}
		ref := s.parent.StringReference(v)
		s.PutByte(byte(t))
		s.PutInt(ref)
	case uo.Location:
		if valueType != TagValueLocation {
			log.Printf("warning: tag %d is not a location value", t)
			return
		}
		s.PutByte(byte(t))
		s.PutLocation(v)
	case uo.Bounds:
		if valueType != TagValueBounds {
			log.Printf("warning: tag %d is not a bounds value", t)
			return
		}
		s.PutByte(byte(t))
		s.PutBounds(v)
	case []int16:
		if valueType != TagValueShortSlice {
			log.Printf("warning: tag %d is not a short slice", t)
			return
		}
		s.PutByte(byte(t))
		s.PutShortSlice(v)
	default:
		log.Printf("warning: unhandled type for tag %d", t)
	}
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

// String returns the next UTF-8 string in the segment by reading a 32-bit
// string reference and indexing the tag file string dictionary.
func (s *TagFileSegment) String() string {
	s.buf.Read(s.tbuf[:4])
	ref := binary.LittleEndian.Uint32(s.tbuf[:4])
	return s.parent.StringByReference(ref)
}

// StringMap returns the next map[string]string encoded into the segment.
func (s *TagFileSegment) StringMap() map[string]string {
	d := s.buf.Bytes()
	ret := make(map[string]string)
	count := int(d[0])
	ofs := 1
	for i := 0; i < count; i++ {
		kref := binary.LittleEndian.Uint32(d[ofs+0 : ofs+4])
		vref := binary.LittleEndian.Uint32(d[ofs+4 : ofs+8])
		ofs += 8
		k := s.parent.StringByReference(kref)
		v := s.parent.StringByReference(vref)
		if k == "" || v == "" {
			continue
		}
		ret[k] = v
	}
	s.buf.Next(ofs)
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

// Location returns the next uo.Location value encoded into the segment.
func (s *TagFileSegment) Location() uo.Location {
	s.buf.Read(s.tbuf[0:5])
	return uo.Location{
		X: int(binary.LittleEndian.Uint16(s.tbuf[0:2])),
		Y: int(binary.LittleEndian.Uint16(s.tbuf[2:4])),
		Z: int(int8(s.tbuf[5])),
	}
}

// Bounds returns the next uo.Bounds value encoded into the segment.
func (s *TagFileSegment) Bounds() uo.Bounds {
	s.buf.Read(s.tbuf[0:10])
	return uo.Bounds{
		X: int(binary.LittleEndian.Uint16(s.tbuf[0:2])),
		Y: int(binary.LittleEndian.Uint16(s.tbuf[2:4])),
		Z: int(int8(s.tbuf[4])),
		W: int(binary.LittleEndian.Uint16(s.tbuf[5:7])),
		H: int(binary.LittleEndian.Uint16(s.tbuf[7:9])),
		D: int(int8(s.tbuf[9])),
	}
}

// ObjectHeader returns the data of the next object header encoded into the
// segment.
func (s *TagFileSegment) ObjectHeader() (
	otype ObjectType,
	serial uo.Serial,
	template string,
	parent uo.Serial,
	name string,
	hue uo.Hue,
	location uo.Location,
	events map[string]string) {
	d := s.buf.Bytes()
	otype = ObjectType(d[0])
	serial = uo.Serial(binary.LittleEndian.Uint32(d[1:5]))
	sref := binary.LittleEndian.Uint32(d[5:9])
	template = s.parent.StringByReference(sref)
	parent = uo.Serial(binary.LittleEndian.Uint32(d[9:13]))
	sref = binary.LittleEndian.Uint32(d[13:17])
	name = s.parent.StringByReference(sref)
	hue = uo.Hue(binary.LittleEndian.Uint16(d[17:19]))
	location = uo.Location{
		X: int(binary.LittleEndian.Uint16(d[19:21])),
		Y: int(binary.LittleEndian.Uint16(d[21:23])),
		Z: int(int8(d[23])),
	}
	s.buf.Next(24)
	events = s.StringMap()
	return
}

// Tags returns a TagCollection containing the next collection of dynamic tags
// encoded in the segment.
func (s *TagFileSegment) Tags() *TagCollection {
	d := s.buf.Bytes()
	c := &TagCollection{
		tags: make(map[Tag]interface{}),
	}
	ofs := 0
	for {
		tag := Tag(d[ofs])
		ofs++
		if tag == TagEndOfList {
			break
		}
		if int(tag) > len(tagValueMapping) {
			break
		}
		switch tagValueMapping[tag] {
		case TagValueBool:
			c.tags[tag] = true
		case TagValueByte:
			c.tags[tag] = d[ofs]
			ofs++
		case TagValueShort:
			c.tags[tag] = binary.LittleEndian.Uint16(d[ofs+0 : ofs+2])
			ofs += 2
		case TagValueInt:
			c.tags[tag] = binary.LittleEndian.Uint32(d[ofs+0 : ofs+4])
			ofs += 4
		case TagValueLong:
			c.tags[tag] = binary.LittleEndian.Uint64(d[ofs+0 : ofs+8])
			ofs += 8
		case TagValueString:
			sref := binary.LittleEndian.Uint32(d[ofs+0 : ofs+4])
			ofs += 4
			str := s.parent.StringByReference(sref)
			if len(str) > 0 {
				c.tags[tag] = str
			}
		case TagValueReferenceSlice:
			var v []uo.Serial
			count := d[ofs]
			ofs++
			for i := 0; i < int(count); i++ {
				v = append(v, uo.Serial(binary.LittleEndian.Uint32(d[ofs+0:ofs+4])))
				ofs += 4
			}
			c.tags[tag] = v
		case TagValueLocation:
			c.tags[tag] = uo.Location{
				X: int(binary.LittleEndian.Uint16(d[ofs+0 : ofs+2])),
				Y: int(binary.LittleEndian.Uint16(d[ofs+2 : ofs+4])),
				Z: int(int8(d[4])),
			}
			ofs += 5
		case TagValueBounds:
			c.tags[tag] = uo.Bounds{
				X: int(binary.LittleEndian.Uint16(d[ofs+0 : ofs+2])),
				Y: int(binary.LittleEndian.Uint16(d[ofs+2 : ofs+4])),
				Z: int(int8(d[4])),
				W: int(binary.LittleEndian.Uint16(d[ofs+5 : ofs+7])),
				H: int(binary.LittleEndian.Uint16(d[ofs+7 : ofs+9])),
				D: int(int8(d[9])),
			}
			ofs += 10
		case TagValueShortSlice:
			var v []int16
			count := d[ofs]
			ofs++
			for i := 0; i < int(count); i++ {
				v = append(v, int16(binary.LittleEndian.Uint16(d[ofs+0:ofs+2])))
				ofs += 2
			}
			c.tags[tag] = v
		default:
			panic(fmt.Sprintf("unhandled tag value type %d", tagValueMapping[tag]))
		}
	}
	s.buf.Next(ofs)
	return c
}

// TagCollection manages a collection of tag values.
type TagCollection struct {
	tags map[Tag]interface{}
}

// Bool returns true if the tag is contained within the collection, false
// otherwise.
func (c *TagCollection) Bool(t Tag) bool {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueBool {
		panic(fmt.Sprintf("tag %d is not a bool", t))
	}
	_, found := c.tags[t]
	return found
}

// Byte returns the named 8-bit value or the default value if not found.
func (c *TagCollection) Byte(t Tag, def byte) byte {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueByte {
		panic(fmt.Sprintf("tag %d is not a byte", t))
	}
	if ret, found := c.tags[t].(byte); found {
		return ret
	}
	return def
}

// Short returns the named 16-bit value or the default value if not found.
func (c *TagCollection) Short(t Tag, def uint16) uint16 {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueShort {
		panic(fmt.Sprintf("tag %d is not a short", t))
	}
	if ret, found := c.tags[t].(uint16); found {
		return ret
	}
	return def
}

// Int returns the named 32-bit value or the default value if not found.
func (c *TagCollection) Int(t Tag, def uint32) uint32 {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueInt {
		panic(fmt.Sprintf("tag %d is not an int", t))
	}
	if ret, found := c.tags[t].(uint32); found {
		return ret
	}
	return def
}

// Long returns the named 64-bit value or the default value if not found.
func (c *TagCollection) Long(t Tag, def uint64) uint64 {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueLong {
		panic(fmt.Sprintf("tag %d is not a long", t))
	}
	if ret, found := c.tags[t].(uint64); found {
		return ret
	}
	return def
}

// String returns the named string value or the default value if not found.
func (c *TagCollection) String(t Tag, def string) string {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueString {
		panic(fmt.Sprintf("tag %d is not a string", t))
	}
	if ret, found := c.tags[t].(string); found {
		return ret
	}
	return def
}

// Location returns the named location value or the default value if not found.
func (c *TagCollection) Location(t Tag, def uo.Location) uo.Location {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueLocation {
		panic(fmt.Sprintf("tag %d is not a location", t))
	}
	if ret, found := c.tags[t].(uo.Location); found {
		return ret
	}
	return def
}

// Bounds returns the named bounds value or the default value if not found.
func (c *TagCollection) Bounds(t Tag, def uo.Bounds) uo.Bounds {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueBounds {
		panic(fmt.Sprintf("tag %d is not a bounds value", t))
	}
	if ret, found := c.tags[t].(uo.Bounds); found {
		return ret
	}
	return def
}

// ReferenceSlice returns the named slice of references or nil if the tag is not
// found.
func (c *TagCollection) ReferenceSlice(t Tag) []uo.Serial {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueReferenceSlice {
		panic(fmt.Sprintf("tag %d is not a reference slice", t))
	}
	if ret, found := c.tags[t].([]uo.Serial); found {
		return ret
	}
	return nil
}

// ShortSlice returns the named slice of shorts or nil if the tag is not found.
func (c *TagCollection) ShortSlice(t Tag) []int16 {
	if int(t) > len(tagValueMapping) || tagValueMapping[t] != TagValueShortSlice {
		panic(fmt.Sprintf("tag %d is not a short slice", t))
	}
	if ret, found := c.tags[t].([]int16); found {
		return ret
	}
	return nil
}

// TagObject manages all of the information for one object.
type TagObject struct {
	Type     ObjectType
	Serial   uo.Serial
	Template string
	Parent   uo.Serial
	Name     string
	Hue      uo.Hue
	Location uo.Location
	Events   map[string]string
	Tags     *TagCollection
}
