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
// ID                     uint8           Unique identifier of the segment
// Offset                 uint64          Offset from the beginning of the file where the raw segment data starts
// Length                 uint64          Length of the raw segment data
// Records                uint32          Number of records in the segment
//
// TagObjectSegment:
// Objects                []TagObject     Object data
//
// TagObject:
// ConcreteType           uint8           A value indicating which Go struct this object will deserialize into
// Serial                 uint32          uo.Serial of the object
// TemplateName           string          Template name this object was created from
// Parent                 uint32          uo.Serial of the object's parent, uo.SerialSystem for the map, or uo.SerialZero for void (leaked)
// Name                   string          Base name string of the object
// Hue                    uint16          uo.Hue of the object
// Location     int16,int16,int8          uo.Location of the object
// EventCount             uint8           Number of event definitions in the object
// Events                 []EventDef      Event definitions
// DynamicTags            []TagValue      The rest of the data for the object as TagValue structures
// TagsEnd
//
// EventDef:
// EventName              string          Event name
// EventHandler           string          Event handler
//
// TagValue:
// Tag                    uint8           Tag code
// TagType                uint8           Tag indicating what the format of the following data will be
// Data                   []byte          Raw data
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

// IncrementRecordCount increments the record count of the segment by one.
func (s *TagFileSegment) IncrementRecordCount() { s.records++ }

// RecordCount returns the number of records encoded into the segment.
func (s *TagFileSegment) RecordCount() uint32 { return s.records }

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
	buf := make([]byte, 7)
	s.buf.WriteByte(byte(otype))
	binary.LittleEndian.PutUint32(buf, uint32(serial))
	s.buf.Write(buf)
	s.buf.WriteString(template)
	s.buf.WriteByte(0)
	binary.LittleEndian.PutUint32(buf, uint32(parent))
	s.buf.Write(buf)
	s.buf.WriteString(name)
	s.buf.WriteByte(0)
	binary.LittleEndian.PutUint16(buf[0:2], uint16(hue))
	binary.LittleEndian.PutUint16(buf[2:4], uint16(hue))
	binary.LittleEndian.PutUint16(buf[4:6], uint16(hue))
	buf[6] = byte(int8(location.Z))
	s.buf.Write(buf[0:7])
	s.PutStringsMap(events)
}

// PutTag writes a tagged value to the segment. Please note that to write a
// reference slice please use PutByte(byte(tag)), PutObjectReferences[I].
func (s *TagFileSegment) PutTag(t Tag, value interface{}) {
	if t == TagEndOfList {
		s.buf.WriteByte(byte(TagEndOfList))
		return
	}
	switch v := value.(type) {
	case bool:
		if v {
			s.tbuf[0] = byte(t)
			s.tbuf[1] = byte(TagValueBool)
			s.buf.Write(s.tbuf[0:2])
		}
	case uint8:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueByte)
		s.tbuf[2] = byte(v)
		s.buf.Write(s.tbuf[0:3])
	case int8:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueByte)
		s.tbuf[2] = byte(v)
		s.buf.Write(s.tbuf[0:3])
	case uint16:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueShort)
		binary.LittleEndian.PutUint16(s.tbuf[2:4], v)
		s.buf.Write(s.tbuf[0:4])
	case int16:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueShort)
		binary.LittleEndian.PutUint16(s.tbuf[2:4], uint16(v))
		s.buf.Write(s.tbuf[0:4])
	case uint:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueInt)
		binary.LittleEndian.PutUint32(s.tbuf[2:6], uint32(v))
		s.buf.Write(s.tbuf[0:6])
	case int:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueInt)
		binary.LittleEndian.PutUint32(s.tbuf[2:6], uint32(v))
		s.buf.Write(s.tbuf[0:6])
	case uint32:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueInt)
		binary.LittleEndian.PutUint32(s.tbuf[2:6], uint32(v))
		s.buf.Write(s.tbuf[0:6])
	case int32:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueInt)
		binary.LittleEndian.PutUint32(s.tbuf[2:6], uint32(v))
		s.buf.Write(s.tbuf[0:6])
	case uint64:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueLong)
		binary.LittleEndian.PutUint64(s.tbuf[2:10], uint64(v))
		s.buf.Write(s.tbuf[0:10])
	case int64:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueLong)
		binary.LittleEndian.PutUint64(s.tbuf[2:10], uint64(v))
		s.buf.Write(s.tbuf[0:10])
	case string:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueString)
		s.buf.Write(s.tbuf[0:1])
		s.buf.WriteString(v)
		s.buf.WriteByte(0)
	case uo.Location:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueLocation)
		binary.LittleEndian.PutUint16(s.tbuf[2:4], uint16(v.X))
		binary.LittleEndian.PutUint16(s.tbuf[4:6], uint16(v.Y))
		s.tbuf[6] = uint8(int8(v.Z))
		s.buf.Write(s.tbuf[0:7])
	case uo.Bounds:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueBounds)
		binary.LittleEndian.PutUint16(s.tbuf[2:4], uint16(v.X))
		binary.LittleEndian.PutUint16(s.tbuf[4:6], uint16(v.Y))
		binary.LittleEndian.PutUint16(s.tbuf[6:8], uint16(v.W))
		binary.LittleEndian.PutUint16(s.tbuf[8:10], uint16(v.H))
		s.buf.Write(s.tbuf[0:10])
	case []int16:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueShortSlice)
		if len(v) > 255 {
			log.Printf("error: tag %d slice too long")
			return
		}
		s.tbuf[2] = byte(len(v))
		ofs := 3
		for _, n := range v {
			binary.LittleEndian.PutUint16(s.tbuf[ofs+0:ofs+2], uint16(n))
			ofs += 2
		}
		s.buf.Write(s.tbuf[0:ofs])
	case []uo.Serial:
		s.tbuf[0] = byte(t)
		s.tbuf[1] = byte(TagValueReferenceSlice)
		if len(v) > 255 {
			log.Printf("error: tag %d slice too long")
			return
		}
		s.tbuf[2] = byte(len(v))
		ofs := 3
		for _, n := range v {
			binary.LittleEndian.PutUint32(s.tbuf[ofs+0:ofs+4], uint32(n))
			ofs += 4
		}
		s.buf.Write(s.tbuf[0:ofs])
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
	d, err := s.buf.ReadString(0)
	if err != nil || len(d) == 0 {
		return ""
	}
	return string(d[0 : len(d)-1])
}

// StringMap returns the next map[string]string encoded into the segment.
func (s *TagFileSegment) StringMap() map[string]string {
	ret := make(map[string]string)
	count, _ := s.buf.ReadByte()
	ofs := 1
	for i := 0; i < int(count); i++ {
		k, _ := s.buf.ReadString(0)
		v, _ := s.buf.ReadString(0)
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

// TagObject returns the data of the next object encoded into the segment.
func (s *TagFileSegment) TagObject() *TagObject {
	otype := ObjectType(s.Byte())
	serial := uo.Serial(s.Int())
	template := s.String()
	parent := uo.Serial(s.Int())
	name := s.String()
	hue := uo.Hue(s.Short())
	location := uo.Location{
		X: int(s.Short()),
		Y: int(s.Short()),
		Z: int(int8(s.Byte())),
	}
	events := s.StringMap()
	tags := s.Tags()
	return &TagObject{
		Type:     otype,
		Serial:   serial,
		Template: template,
		Parent:   parent,
		Name:     name,
		Hue:      hue,
		Location: location,
		Events:   events,
		Tags:     tags,
	}
}

// Tags returns a TagCollection containing the next collection of dynamic tags
// encoded in the segment.
func (s *TagFileSegment) Tags() *TagCollection {
	c := &TagCollection{
		tags: make([]interface{}, TagLastValidValue+1),
	}
	for {
		tag := Tag(s.Byte())
		if tag == TagEndOfList {
			break
		}
		if tag > TagLastValidValue {
			// Something is wrong
			log.Printf("warning: save file contains unknown tag ID %d", tag)
			break
		}
		tagType := TagValue(s.Byte())
		switch tagType {
		case TagValueBool:
			c.tags[tag] = true
		case TagValueByte:
			c.tags[tag] = s.Byte()
		case TagValueShort:
			c.tags[tag] = s.Short()
		case TagValueInt:
			c.tags[tag] = s.Int()
		case TagValueLong:
			c.tags[tag] = s.Long()
		case TagValueString:
			str := s.String()
			if len(str) > 0 {
				c.tags[tag] = str
			}
		case TagValueReferenceSlice:
			var v []uo.Serial
			count := int(s.Byte())
			for i := 0; i < count; i++ {
				v = append(v, uo.Serial(s.Int()))
			}
			c.tags[tag] = v
		case TagValueLocation:
			c.tags[tag] = s.Location()
		case TagValueBounds:
			c.tags[tag] = s.Bounds()
		case TagValueShortSlice:
			var v []int16
			count := int(s.Byte())
			for i := 0; i < count; i++ {
				v = append(v, int16(s.Short()))
			}
			c.tags[tag] = v
		default:
			panic(fmt.Sprintf("unhandled tag value type %d", tagType))
		}
	}
	return c
}

// TagCollection manages a collection of tag values.
type TagCollection struct {
	tags []interface{}
}

// Bool returns true if the tag is contained within the collection, false
// otherwise.
func (c *TagCollection) Bool(t Tag) bool {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].(bool)
}

// Byte returns the named 8-bit value
func (c *TagCollection) Byte(t Tag) byte {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].(byte)
}

// Short returns the named 16-bit value
func (c *TagCollection) Short(t Tag) uint16 {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].(uint16)
}

// Int returns the named 32-bit value
func (c *TagCollection) Int(t Tag) uint32 {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].(uint32)
}

// Long returns the named 64-bit value or the default value if not found.
func (c *TagCollection) Long(t Tag) uint64 {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].(uint64)
}

// String returns the named string value
func (c *TagCollection) String(t Tag) string {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].(string)
}

// Location returns the named location value
func (c *TagCollection) Location(t Tag) uo.Location {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].(uo.Location)
}

// Bounds returns the named bounds value
func (c *TagCollection) Bounds(t Tag) uo.Bounds {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].(uo.Bounds)
}

// ReferenceSlice returns the named slice of references
func (c *TagCollection) ReferenceSlice(t Tag) []uo.Serial {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].([]uo.Serial)
}

// ShortSlice returns the named slice of shorts
func (c *TagCollection) ShortSlice(t Tag) []int16 {
	if t > TagLastValidValue {
		panic(fmt.Sprintf("warning: unknown tag type %d", t))
	}
	return c.tags[t].([]int16)
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
