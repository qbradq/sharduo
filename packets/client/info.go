package client

import (
	"github.com/qbradq/sharduo/common"
	"github.com/qbradq/sharduo/packets/server"
)

type clientPacketDecoder func(r *common.PacketReader, s server.PacketSender)

// PacketInfo describes a client packet and its decoding function
type PacketInfo struct {
	ID      int
	Length  int // -1 = dynamic, 0 = unsupported
	Decoder clientPacketDecoder
}

// PacketInfos is an array of PacketInfo structures describing all
// 256 possible client packet IDs
var PacketInfos = [256]PacketInfo{
	PacketInfo{0x00, 104, nil},
	PacketInfo{0x01, 5, nil},
	PacketInfo{0x02, 7, nil},
	PacketInfo{0x03, -1, nil},
	PacketInfo{0x04, 2, nil},
	PacketInfo{0x05, 5, nil},
	PacketInfo{0x06, 5, nil},
	PacketInfo{0x07, 7, nil},
	PacketInfo{0x08, 14, nil},
	PacketInfo{0x09, 5, nil},
	PacketInfo{0x0a, 11, nil},
	PacketInfo{0x0b, 0, nil},
	PacketInfo{0x0c, -1, nil},
	PacketInfo{0x0d, 0, nil},
	PacketInfo{0x0e, 0, nil},
	PacketInfo{0x0f, 0, nil},
	PacketInfo{0x10, 0, nil},
	PacketInfo{0x11, 0, nil},
	PacketInfo{0x12, -1, nil},
	PacketInfo{0x13, 10, nil},
	PacketInfo{0x14, 6, nil},
	PacketInfo{0x15, 9, nil},
	PacketInfo{0x16, 0, nil},
	PacketInfo{0x17, 0, nil},
	PacketInfo{0x18, 0, nil},
	PacketInfo{0x19, 0, nil},
	PacketInfo{0x1a, 0, nil},
	PacketInfo{0x1b, 0, nil},
	PacketInfo{0x1c, 0, nil},
	PacketInfo{0x1d, 0, nil},
	PacketInfo{0x1e, 4, nil},
	PacketInfo{0x1f, 0, nil},
	PacketInfo{0x20, 0, nil},
	PacketInfo{0x21, 0, nil},
	PacketInfo{0x22, 3, nil},
	PacketInfo{0x23, 0, nil},
	PacketInfo{0x24, 0, nil},
	PacketInfo{0x25, 0, nil},
	PacketInfo{0x26, 0, nil},
	PacketInfo{0x27, 0, nil},
	PacketInfo{0x28, 0, nil},
	PacketInfo{0x29, 0, nil},
	PacketInfo{0x2a, 0, nil},
	PacketInfo{0x2b, 0, nil},
	PacketInfo{0x2c, 2, nil},
	PacketInfo{0x2d, 0, nil},
	PacketInfo{0x2e, 0, nil},
	PacketInfo{0x2f, 0, nil},
	PacketInfo{0x30, 0, nil},
	PacketInfo{0x31, 0, nil},
	PacketInfo{0x32, 0, nil},
	PacketInfo{0x33, 0, nil},
	PacketInfo{0x34, 10, nil},
	PacketInfo{0x35, 0, nil},
	PacketInfo{0x36, 0, nil},
	PacketInfo{0x37, 0, nil},
	PacketInfo{0x38, 7, nil},
	PacketInfo{0x39, 9, nil},
	PacketInfo{0x3a, -1, nil},
	PacketInfo{0x3b, -1, nil},
	PacketInfo{0x3c, 0, nil},
	PacketInfo{0x3d, 0, nil},
	PacketInfo{0x3e, 0, nil},
	PacketInfo{0x3f, 0, nil},
	PacketInfo{0x40, 0, nil},
	PacketInfo{0x41, 0, nil},
	PacketInfo{0x42, 0, nil},
	PacketInfo{0x43, 0, nil},
	PacketInfo{0x44, 0, nil},
	PacketInfo{0x45, 0, nil},
	PacketInfo{0x46, 0, nil},
	PacketInfo{0x47, 0, nil},
	PacketInfo{0x48, 0, nil},
	PacketInfo{0x49, 0, nil},
	PacketInfo{0x4a, 0, nil},
	PacketInfo{0x4b, 0, nil},
	PacketInfo{0x4c, 0, nil},
	PacketInfo{0x4d, 0, nil},
	PacketInfo{0x4e, 0, nil},
	PacketInfo{0x4f, 0, nil},
	PacketInfo{0x50, -1, nil},
	PacketInfo{0x51, -1, nil},
	PacketInfo{0x52, -1, nil},
	PacketInfo{0x53, 0, nil},
	PacketInfo{0x54, 0, nil},
	PacketInfo{0x55, 0, nil},
	PacketInfo{0x56, 11, nil},
	PacketInfo{0x57, 0, nil},
	PacketInfo{0x58, 0, nil},
	PacketInfo{0x59, 0, nil},
	PacketInfo{0x5a, 0, nil},
	PacketInfo{0x5b, 0, nil},
	PacketInfo{0x5c, 0, nil},
	PacketInfo{0x5d, 73, nil},
	PacketInfo{0x5e, 0, nil},
	PacketInfo{0x5f, 0, nil},
	PacketInfo{0x60, 0, nil},
	PacketInfo{0x61, 0, nil},
	PacketInfo{0x62, 0, nil},
	PacketInfo{0x63, 0, nil},
	PacketInfo{0x64, 0, nil},
	PacketInfo{0x65, 0, nil},
	PacketInfo{0x66, -1, nil},
	PacketInfo{0x67, 0, nil},
	PacketInfo{0x68, 0, nil},
	PacketInfo{0x69, 5, nil},
	PacketInfo{0x6a, 0, nil},
	PacketInfo{0x6b, 0, nil},
	PacketInfo{0x6c, 19, nil},
	PacketInfo{0x6d, 0, nil},
	PacketInfo{0x6e, 0, nil},
	PacketInfo{0x6f, -1, nil},
	PacketInfo{0x70, 0, nil},
	PacketInfo{0x71, -1, nil},
	PacketInfo{0x72, 5, nil},
	PacketInfo{0x73, 2, nil},
	PacketInfo{0x74, 0, nil},
	PacketInfo{0x75, 35, nil},
	PacketInfo{0x76, 0, nil},
	PacketInfo{0x77, 0, nil},
	PacketInfo{0x78, 0, nil},
	PacketInfo{0x79, 0, nil},
	PacketInfo{0x7a, 0, nil},
	PacketInfo{0x7b, 0, nil},
	PacketInfo{0x7c, 0, nil},
	PacketInfo{0x7d, 13, nil},
	PacketInfo{0x7e, 0, nil},
	PacketInfo{0x7f, 0, nil},
	PacketInfo{0x80, 62, x80},
	PacketInfo{0x81, 0, nil},
	PacketInfo{0x82, 0, nil},
	PacketInfo{0x83, 39, nil},
	PacketInfo{0x84, 0, nil},
	PacketInfo{0x85, 0, nil},
	PacketInfo{0x86, 0, nil},
	PacketInfo{0x87, 0, nil},
	PacketInfo{0x88, 0, nil},
	PacketInfo{0x89, 0, nil},
	PacketInfo{0x8a, 0, nil},
	PacketInfo{0x8b, 0, nil},
	PacketInfo{0x8c, 0, nil},
	PacketInfo{0x8d, 0, nil},
	PacketInfo{0x8e, 0, nil},
	PacketInfo{0x8f, 0, nil},
	PacketInfo{0x90, 0, nil},
	PacketInfo{0x91, 65, nil},
	PacketInfo{0x92, 0, nil},
	PacketInfo{0x93, 99, nil},
	PacketInfo{0x94, 0, nil},
	PacketInfo{0x95, 9, nil},
	PacketInfo{0x96, 0, nil},
	PacketInfo{0x97, 0, nil},
	PacketInfo{0x98, -1, nil},
	PacketInfo{0x99, 26, nil},
	PacketInfo{0x9a, -1, nil},
	PacketInfo{0x9b, 258, nil},
	PacketInfo{0x9c, 0, nil},
	PacketInfo{0x9d, 0, nil},
	PacketInfo{0x9e, 0, nil},
	PacketInfo{0x9f, -1, nil},
	PacketInfo{0xa0, 3, xA0},
	PacketInfo{0xa1, 0, nil},
	PacketInfo{0xa2, 0, nil},
	PacketInfo{0xa3, 0, nil},
	PacketInfo{0xa4, 149, nil},
	PacketInfo{0xa5, 0, nil},
	PacketInfo{0xa6, 0, nil},
	PacketInfo{0xa7, 4, nil},
	PacketInfo{0xa8, 0, nil},
	PacketInfo{0xa9, 0, nil},
	PacketInfo{0xaa, 0, nil},
	PacketInfo{0xab, 0, nil},
	PacketInfo{0xac, -1, nil},
	PacketInfo{0xad, -1, nil},
	PacketInfo{0xae, 0, nil},
	PacketInfo{0xaf, 0, nil},
	PacketInfo{0xb0, 0, nil},
	PacketInfo{0xb1, -1, nil},
	PacketInfo{0xb2, 0, nil},
	PacketInfo{0xb3, -1, nil},
	PacketInfo{0xb4, 0, nil},
	PacketInfo{0xb5, 64, nil},
	PacketInfo{0xb6, 9, nil},
	PacketInfo{0xb7, 0, nil},
	PacketInfo{0xb8, -1, nil},
	PacketInfo{0xb9, 0, nil},
	PacketInfo{0xba, 0, nil},
	PacketInfo{0xbb, 9, nil},
	PacketInfo{0xbc, 0, nil},
	PacketInfo{0xbd, -1, nil},
	PacketInfo{0xbe, -1, nil},
	PacketInfo{0xbf, -1, nil},
	PacketInfo{0xc0, 0, nil},
	PacketInfo{0xc1, 0, nil},
	PacketInfo{0xc2, -1, nil},
	PacketInfo{0xc3, 0, nil},
	PacketInfo{0xc4, 0, nil},
	PacketInfo{0xc5, 1, nil},
	PacketInfo{0xc6, 0, nil},
	PacketInfo{0xc7, 0, nil},
	PacketInfo{0xc8, 2, nil},
	PacketInfo{0xc9, 6, nil},
	PacketInfo{0xca, 6, nil},
	PacketInfo{0xcb, 0, nil},
	PacketInfo{0xcc, 0, nil},
	PacketInfo{0xcd, 0, nil},
	PacketInfo{0xce, 0, nil},
	PacketInfo{0xcf, 0, nil},
	PacketInfo{0xd0, 0, nil},
	PacketInfo{0xd1, 2, nil},
	PacketInfo{0xd2, 0, nil},
	PacketInfo{0xd3, 0, nil},
	PacketInfo{0xd4, -1, nil},
	PacketInfo{0xd5, 0, nil},
	PacketInfo{0xd6, -1, nil},
	PacketInfo{0xd7, -1, nil},
	PacketInfo{0xd8, 0, nil},
	PacketInfo{0xd9, 199, nil},
	PacketInfo{0xda, 0, nil},
	PacketInfo{0xdb, 0, nil},
	PacketInfo{0xdc, 0, nil},
	PacketInfo{0xdd, 0, nil},
	PacketInfo{0xde, 0, nil},
	PacketInfo{0xdf, 0, nil},
	PacketInfo{0xe0, 0, nil},
	PacketInfo{0xe1, 0, nil},
	PacketInfo{0xe2, 0, nil},
	PacketInfo{0xe3, 0, nil},
	PacketInfo{0xe4, 0, nil},
	PacketInfo{0xe5, 0, nil},
	PacketInfo{0xe6, 0, nil},
	PacketInfo{0xe7, 0, nil},
	PacketInfo{0xe8, 0, nil},
	PacketInfo{0xe9, 0, nil},
	PacketInfo{0xea, 0, nil},
	PacketInfo{0xeb, 0, nil},
	PacketInfo{0xec, 0, nil},
	PacketInfo{0xed, 0, nil},
	PacketInfo{0xee, 0, nil},
	PacketInfo{0xef, 0, nil},
	PacketInfo{0xf0, 0, nil},
	PacketInfo{0xf1, -1, nil},
	PacketInfo{0xf2, 0, nil},
	PacketInfo{0xf3, 0, nil},
	PacketInfo{0xf4, 0, nil},
	PacketInfo{0xf5, 0, nil},
	PacketInfo{0xf6, 0, nil},
	PacketInfo{0xf7, 0, nil},
	PacketInfo{0xf8, 0, nil},
	PacketInfo{0xf9, 0, nil},
	PacketInfo{0xfa, 0, nil},
	PacketInfo{0xfb, 0, nil},
	PacketInfo{0xfc, 0, nil},
	PacketInfo{0xfd, 0, nil},
	PacketInfo{0xfe, 0, nil},
	PacketInfo{0xff, 0, nil},
}
