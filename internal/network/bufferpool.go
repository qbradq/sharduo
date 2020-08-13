package network

import "sync"

// Buffer is a wrapper for a byte slice.
type Buffer struct {
	B []byte
}

// Reset resets the byte slice but not the underlying array.
func (b *Buffer) Reset() {
	b.B = b.B[:0]
}

var buffersAvailable = make([]*Buffer, 0)
var bufferMux sync.Mutex

// GetBuffer returns a new buffer from the pool.
func GetBuffer() *Buffer {
	bufferMux.Lock()
	defer bufferMux.Unlock()
	if len(buffersAvailable) > 0 {
		b := buffersAvailable[len(buffersAvailable)-1]
		buffersAvailable = buffersAvailable[:len(buffersAvailable)-1]
		return b
	}
	return &Buffer{
		B: make([]byte, 0, 64*1024),
	}
}

// PutBuffer returns buffer b to the pool.
func PutBuffer(b *Buffer) {
	bufferMux.Lock()
	defer bufferMux.Unlock()
	buffersAvailable = append(buffersAvailable, b)
}
