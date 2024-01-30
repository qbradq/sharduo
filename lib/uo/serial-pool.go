package uo

// SerialPool manages a range of uo.Serial values ensuring that all issued
// values are unique.
type SerialPool struct {
	minSerial  Serial              // Minimum valid serial in the range
	maxSerial  Serial              // Maximum valid serial in the range
	nilSerial  Serial              // Reserved serial for the nil value
	nextSerial Serial              // The next serial in the sequence
	inUseSet   map[Serial]struct{} // Set of in-use serials
}

// NewSerialPool constructs a new SerialPool for use.
func NewSerialPool(min, max, zero Serial) *SerialPool {
	return &SerialPool{
		minSerial:  min,
		maxSerial:  max,
		nilSerial:  zero,
		nextSerial: min,
		inUseSet:   map[Serial]struct{}{},
	}
}

// Next returns the next available serial in the pool.
func (p *SerialPool) Next() Serial {
	var ret Serial
	for {
		_, inUse := p.inUseSet[p.nextSerial]
		if !inUse {
			p.inUseSet[p.nextSerial] = struct{}{}
			ret = p.nextSerial
		}
		p.nextSerial++
		if p.nextSerial > p.maxSerial {
			p.nextSerial = p.minSerial
		}
		if !inUse {
			break
		}
	}
	return ret
}

// Release flags a serial as no longer in use.
func (p *SerialPool) Release(s Serial) {
	delete(p.inUseSet, s)
}

// Insert inserts the serial into the set.
func (p *SerialPool) Insert(s Serial) {
	p.inUseSet[s] = struct{}{}
	if s > p.nextSerial {
		p.nextSerial = s + 1
	}
}
