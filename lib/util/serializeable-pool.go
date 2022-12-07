package util

// SerializeablePool manages a pool of Serializeable objects. NOT THREAD SAFE!
type SerializeablePool struct {
	objects map[uo.Serial]Serializeable
	sm      *SerialManager
	name    string
}

// NewSerializeablePool returns a new SerializeablePool object ready for use.
func NewSerializeablePool(name string) *SerializeablePool {
	return &SerializeablePool{
		objects: make(map[uo.Serial]Serializeable),
		sm:      NewSerialManager(),
		name:    name,
	}
}

// Add adds the serializeable object to the pool, assigning it a unique serial.
func (p *SerializeablePool) Add(o Serializeable, stype uo.SerialType) {
	o.SetSerial(sm.New(stype))
	sm.Add(o.GetSerial())
	objects[o.GetSerial()] = o
}

// Remove removes the serializeable object from the pool, assigning it
// uo.SerialNone.
func (p *SerializeablePool) Remove(o Serializeable) {
	sm.Remove(o.GetSerial())
	delete(objects, o.GetSerial())
	o.SetSerial(uo.SerialNone)
}

// Map executes fn on every object in the pool and returns a slice of all
// non-nil return values.
func (p *SerializeablePool) Map(fn func(Serializeable) error) []error {
	var ret []error
	for _, o := range p.objects {
		if err := fn(o); err != nil {
			ret = append(ret, err)
		}
	}
	return ret
}

// Save writes the SerialiezablePool and its contents to the writer
func (p *SerializeablePool) Save(w io.Writer) []error {
	tfw := NewTagFileWriter(w)
	tfw.WriteCommentLine(p.name)
	tfw.WriteBlankLine()
	for _, o := range p.objects {
		o.Serialize(tfw)
		tfw.WriteBlankLine()
	}
	return tfw.Errors()
}

// Read reads the contents of the reader
