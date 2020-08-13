package uo

import (
	"testing"
)

func TestHuffmanCompressPacket(t *testing.T) {
	var expected = []byte{
		0x80,
		0xce,
		0xce,
		0x07,
		0xc5,
		0xa0,
	}

	out := make([]byte, 0)
	smp := NewServerPacketSetMap(make([]byte, 0), 0xa8)
	out = HuffmanEncodePacket(smp, out)
	if len(out) != len(expected) {
		t.Fatal("Length mismatch")
	}
	for idx := range out {
		g := out[idx]
		e := expected[idx]
		if g != e {
			t.Fatalf("Bad output at %d got %v wanted %v", idx, g, e)
		}
	}
}
