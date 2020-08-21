package serverpacket

import (
	"bytes"
	"net"
	"testing"
)

func TestCompressedWriter(t *testing.T) {
	input := &ConnectToGameServer{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 2592,
		Key:  []byte{0xdb, 0x1f, 0x2a, 0x70},
	}
	expected := []byte{0x5f, 0x5c, 0xa1, 0xfa, 0xd6, 0xcc, 0xb8, 0xc8, 0x2b, 0x1d}
	outbuf := bytes.NewBuffer(nil)
	uat := NewCompressedWriter()
	if err := uat.Write(input, outbuf); err != nil {
		t.Fatal(err)
	}
	for idx, g := range outbuf.Bytes() {
		e := expected[idx]
		if e != g {
			t.Fatalf("Bad byte at position %d got %#v expected %#v", idx, outbuf.Bytes(), expected)
		}
	}
}
