package uo

import (
	"bytes"
	"testing"
)

func TestServerPacketWriter(t *testing.T) {
	var expected = []byte{
		0xbf, // Extended packet
		0x00, // Length
		0x06,
		0x00, // Subcommand 8, change map
		0x08,
		0xa8, // Map ID
	}

	smp := NewServerPacketSetMap(make([]byte, 0), 0xa8)
	buf := bytes.NewBuffer(nil)
	uat := NewServerPacketWriter(buf)
	if err := uat.WritePacket(smp); err != nil {
		t.Fatal(err)
	}
	out := buf.Bytes()
	if len(out) != len(expected) {
		t.Fatal("Length mismatch")
	}
	for idx, b := range out {
		if b != expected[idx] {
			t.Fatalf("Bad write at %d, got %v expected %v", idx, b, expected[idx])
		}
	}
}

func TestServerPacketWriterCompression(t *testing.T) {
	var expected = []byte{
		0x80,
		0xce,
		0xce,
		0x07,
		0xc5,
		0xa0,
	}

	smp := NewServerPacketSetMap(make([]byte, 0), 0xa8)
	buf := bytes.NewBuffer(nil)
	uat := NewServerPacketWriter(buf)
	uat.SetCompression(true)
	if err := uat.WritePacket(smp); err != nil {
		t.Fatal(err)
	}
	out := buf.Bytes()
	if len(out) != len(expected) {
		t.Fatal("Length mismatch")
	}
	for idx, b := range out {
		if b != expected[idx] {
			t.Fatalf("Bad write at %d, got %v expected %v", idx, b, expected[idx])
		}
	}
}
