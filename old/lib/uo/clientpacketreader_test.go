package uo

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

func TestPacketReader(t *testing.T) {
	var testReader = bytes.NewReader(testData)
	uat := NewClientPacketReader(testReader)
	_, err := uat.ReadConnectionHeader()
	if err != nil {
		t.Fatal(err)
	}
	lp, err := uat.ReadClientPacket()
	if err != nil {
		t.Fatal(err)
	}
	_, ok := lp.(ClientPacketAccountLogin)
	if !ok {
		t.Fatal("Failed to decode the login packet")
	}
	ssp, err := uat.ReadClientPacket()
	if err != nil {
		t.Fatal(err)
	}
	_, ok = ssp.(ClientPacketSelectServer)
	if !ok {
		t.Fatal("Failed to decode the select server packet")
	}
	gmp, err := uat.ReadClientPacket()
	if err != nil {
		t.Fatal(err)
	}
	_, ok = gmp.(ClientPacketNotSupported)
	if !ok {
		t.Fatal("Failed to detect unsupported client packet")
	}
	np, err := uat.ReadClientPacket()
	if np != nil || err != io.EOF {
		t.Fatal("Failed to detect end of client packet stream")
	}
}
