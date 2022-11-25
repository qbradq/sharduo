package clientpacket

import "fmt"

// Factory creates Packet implementations based on a data slice.
type Factory struct {
	// Map of packet constructors
	ctors map[int]func([]byte) Packet
	// Debug name of this factory
	name string
}

func newFactory(name string) *Factory {
	return &Factory{
		ctors: make(map[int]func([]byte) Packet),
		name:  name,
	}
}

func (f *Factory) add(id int, ctor func([]byte) Packet) {
	if _, duplicate := f.ctors[id]; duplicate {
		panic(fmt.Sprintf("in packet factory %s duplicate ctor %d", f.name, id))
	}
	f.ctors[id] = ctor
}

func (f *Factory) ignore(id int) {
	f.add(id, func(in []byte) Packet {
		return &IgnoredPacket{
			Base: Base{ID: id},
		}
	})
}

func (f *Factory) new(id int, in []byte) Packet {
	var ret Packet
	ctor := f.ctors[id]
	if ctor != nil {
		ret = ctor(in)
	} else {
		ret = newUnsupportedPacket(f.name, in)
		ret.setId(id)
	}
	return ret
}
