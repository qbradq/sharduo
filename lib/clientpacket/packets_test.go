package clientpacket

import (
	"testing"
)

func TestPackets(t *testing.T) {
	var tests = []struct {
		id   byte
		data []byte
	}{
		{0x80, []byte{0x80, 0x6c, 0x61, 0x7a, 0x79, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x61, 0x73, 0x64, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff}},
		{0x91, []byte{0x91, 0xdb, 0x1f, 0x2a, 0x70, 0x6c, 0x61, 0x7a, 0x79, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x61, 0x73, 0x64, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{0xA0, []byte{0xa0, 0x0, 0x0}},
	}

	for _, test := range tests {
		p := New(test.data)
		if _, ok := p.(*UnsupportedPacket); ok {
			t.Fatalf("Unsupported packet %X", test.id)
		}
		if test.id != p.GetID() {
			t.Fatalf("Packet %X ID mismatch", test.id)
		}
	}
}
