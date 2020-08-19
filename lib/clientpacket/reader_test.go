package clientpacket

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"
)

var testDataStr string
var testData []byte

func init() {
	var err error

	// Connection header
	testDataStr += "0100007f"
	// Login packet
	testDataStr += "806c617a790000000000000000000000000000000000000000000000000000617364660000000000000000000000000000000000000000000000000000ff"
	// Server select packet
	testDataStr += "a00000"
	// Request God Mode ON (Unsupported packet)
	testDataStr += "0401"
	testData, err = hex.DecodeString(testDataStr)
	if err != nil {
		panic(err)
	}
}

func TestPacketReaderHeader(t *testing.T) {
	r := bytes.NewReader(testData)
	uat := NewReader(r)
	if err := uat.ReadConnectionHeader(); err != nil {
		t.Fatal(err)
	}
	if len(uat.Header) != 4 {
		t.Fatal("Header length != 4")
	}
	for idx, g := range uat.Header {
		e := testData[idx]
		if g != e {
			t.Fatalf("Bad header at position %d, got %X, expected %X", idx, g, e)
		}
	}
}

func TestReader(t *testing.T) {
	r := bytes.NewReader(testData)
	uat := NewReader(r)
	if err := uat.ReadConnectionHeader(); err != nil {
		t.Fatal(err)
	}
	p, err := uat.Read()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.(*AccountLogin); !ok {
		t.Fatal("Failed to get account login packet")
	}
	p, err = uat.Read()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.(*SelectServer); !ok {
		t.Fatal("Failed to get select server packet")
	}
	p, err = uat.Read()
	if err != nil {
		t.Fatal(err)
	}
	if p != nil {
		t.Fatal("Failed to return unsupported packet as nil")
	}
	np, err := uat.Read()
	if np != nil || err != io.EOF {
		t.Fatal("Failed to detect end of client packet stream", np, err)
	}
}

func TestReaderUnknownPacket(t *testing.T) {
	var input = []byte{
		0x01,
		0x02,
		0x03,
		0x04,
		0x0b,
		0x00,
		0x05,
		0x00,
		0x01,
	}

	r := bytes.NewReader(input)
	uat := NewReader(r)
	if err := uat.ReadConnectionHeader(); err != nil {
		t.Fatal(err)
	}
	_, err := uat.Read()
	if err != ErrUnknownPacket {
		t.Fatal("Failed to detect unknown packet")
	}
}
