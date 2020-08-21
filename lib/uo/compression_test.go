package uo

import (
	"bytes"
	"testing"
)

func TestHuffmanCompressPacket(t *testing.T) {
	var input = []byte{
		0xbf,
		0x00,
		0x06,
		0x00,
		0x08,
		0xa8,
	}
	var expected = []byte{
		0x80,
		0xce,
		0xce,
		0x07,
		0xc5,
		0xa0,
	}

	in := bytes.NewBuffer(input)
	outbuf := bytes.NewBuffer(nil)
	if err := HuffmanEncodePacket(in, outbuf); err != nil {
		t.Fatal(err)
	}
	out := outbuf.Bytes()
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

func TestHuffmanDecompressPacket(t *testing.T) {
	var input = []byte{
		0x80,
		0xce,
		0xce,
		0x07,
		0xc5,
		0xa0,
	}
	var expected = []byte{
		0xbf,
		0x00,
		0x06,
		0x00,
		0x08,
		0xa8,
	}

	in := bytes.NewBuffer(input)
	outbuf := bytes.NewBuffer(nil)
	if err := HuffmanDecodePacket(in, outbuf); err != nil {
		t.Fatal(err)
	}
	out := outbuf.Bytes()
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

func TestHuffmanDecompressFragmentedPacket(t *testing.T) {
	var input = []byte{
		0x80,
		0xce,
		0xce,
		0x07,
	}

	in := bytes.NewBuffer(input)
	outbuf := bytes.NewBuffer(nil)
	if err := HuffmanDecodePacket(in, outbuf); err != ErrIncompletePacket {
		t.Fatal("Failed to detect premature end of bitstream")
	}
}
