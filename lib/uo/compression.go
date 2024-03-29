package uo

import (
	"errors"
	"io"
)

// ErrIncompletePacket is used to flag incomplete decompression of a packet.
var ErrIncompletePacket = errors.New("incomplete packet")

// Table and compression algorithm documented here:
// https://sites.google.com/site/ultimaonlineoldpackets/client/compression
var huffmanTable = [...]uint16{
	0x2, 0x000, 0x5, 0x01F, 0x6, 0x022, 0x7, 0x034, 0x7, 0x075, 0x6, 0x028,
	0x6, 0x03B, 0x7, 0x032, 0x8, 0x0E0, 0x8, 0x062, 0x7, 0x056, 0x8, 0x079,
	0x9, 0x19D, 0x8, 0x097, 0x6, 0x02A, 0x7, 0x057, 0x8, 0x071, 0x8, 0x05B,
	0x9, 0x1CC, 0x8, 0x0A7, 0x7, 0x025, 0x7, 0x04F, 0x8, 0x066, 0x8, 0x07D,
	0x9, 0x191, 0x9, 0x1CE, 0x7, 0x03F, 0x9, 0x090, 0x8, 0x059, 0x8, 0x07B,
	0x8, 0x091, 0x8, 0x0C6, 0x6, 0x02D, 0x9, 0x186, 0x8, 0x06F, 0x9, 0x093,
	0xA, 0x1CC, 0x8, 0x05A, 0xA, 0x1AE, 0xA, 0x1C0, 0x9, 0x148, 0x9, 0x14A,
	0x9, 0x082, 0xA, 0x19F, 0x9, 0x171, 0x9, 0x120, 0x9, 0x0E7, 0xA, 0x1F3,
	0x9, 0x14B, 0x9, 0x100, 0x9, 0x190, 0x6, 0x013, 0x9, 0x161, 0x9, 0x125,
	0x9, 0x133, 0x9, 0x195, 0x9, 0x173, 0x9, 0x1CA, 0x9, 0x086, 0x9, 0x1E9,
	0x9, 0x0DB, 0x9, 0x1EC, 0x9, 0x08B, 0x9, 0x085, 0x5, 0x00A, 0x8, 0x096,
	0x8, 0x09C, 0x9, 0x1C3, 0x9, 0x19C, 0x9, 0x08F, 0x9, 0x18F, 0x9, 0x091,
	0x9, 0x087, 0x9, 0x0C6, 0x9, 0x177, 0x9, 0x089, 0x9, 0x0D6, 0x9, 0x08C,
	0x9, 0x1EE, 0x9, 0x1EB, 0x9, 0x084, 0x9, 0x164, 0x9, 0x175, 0x9, 0x1CD,
	0x8, 0x05E, 0x9, 0x088, 0x9, 0x12B, 0x9, 0x172, 0x9, 0x10A, 0x9, 0x08D,
	0x9, 0x13A, 0x9, 0x11C, 0xA, 0x1E1, 0xA, 0x1E0, 0x9, 0x187, 0xA, 0x1DC,
	0xA, 0x1DF, 0x7, 0x074, 0x9, 0x19F, 0x8, 0x08D, 0x8, 0x0E4, 0x7, 0x079,
	0x9, 0x0EA, 0x9, 0x0E1, 0x8, 0x040, 0x7, 0x041, 0x9, 0x10B, 0x9, 0x0B0,
	0x8, 0x06A, 0x8, 0x0C1, 0x7, 0x071, 0x7, 0x078, 0x8, 0x0B1, 0x9, 0x14C,
	0x7, 0x043, 0x8, 0x076, 0x7, 0x066, 0x7, 0x04D, 0x9, 0x08A, 0x6, 0x02F,
	0x8, 0x0C9, 0x9, 0x0CE, 0x9, 0x149, 0x9, 0x160, 0xA, 0x1BA, 0xA, 0x19E,
	0xA, 0x39F, 0x9, 0x0E5, 0x9, 0x194, 0x9, 0x184, 0x9, 0x126, 0x7, 0x030,
	0x8, 0x06C, 0x9, 0x121, 0x9, 0x1E8, 0xA, 0x1C1, 0xA, 0x11D, 0xA, 0x163,
	0xA, 0x385, 0xA, 0x3DB, 0xA, 0x17D, 0xA, 0x106, 0xA, 0x397, 0xA, 0x24E,
	0x7, 0x02E, 0x8, 0x098, 0xA, 0x33C, 0xA, 0x32E, 0xA, 0x1E9, 0x9, 0x0BF,
	0xA, 0x3DF, 0xA, 0x1DD, 0xA, 0x32D, 0xA, 0x2ED, 0xA, 0x30B, 0xA, 0x107,
	0xA, 0x2E8, 0xA, 0x3DE, 0xA, 0x125, 0xA, 0x1E8, 0x9, 0x0E9, 0xA, 0x1CD,
	0xA, 0x1B5, 0x9, 0x165, 0xA, 0x232, 0xA, 0x2E1, 0xB, 0x3AE, 0xB, 0x3C6,
	0xB, 0x3E2, 0xA, 0x205, 0xA, 0x29A, 0xA, 0x248, 0xA, 0x2CD, 0xA, 0x23B,
	0xB, 0x3C5, 0xA, 0x251, 0xA, 0x2E9, 0xA, 0x252, 0x9, 0x1EA, 0xB, 0x3A0,
	0xB, 0x391, 0xA, 0x23C, 0xB, 0x392, 0xB, 0x3D5, 0xA, 0x233, 0xA, 0x2CC,
	0xB, 0x390, 0xA, 0x1BB, 0xB, 0x3A1, 0xB, 0x3C4, 0xA, 0x211, 0xA, 0x203,
	0x9, 0x12A, 0xA, 0x231, 0xB, 0x3E0, 0xA, 0x29B, 0xB, 0x3D7, 0xA, 0x202,
	0xB, 0x3AD, 0xA, 0x213, 0xA, 0x253, 0xA, 0x32C, 0xA, 0x23D, 0xA, 0x23F,
	0xA, 0x32F, 0xA, 0x11C, 0xA, 0x384, 0xA, 0x31C, 0xA, 0x17C, 0xA, 0x30A,
	0xA, 0x2E0, 0xA, 0x276, 0xA, 0x250, 0xB, 0x3E3, 0xA, 0x396, 0xA, 0x18F,
	0xA, 0x204, 0xA, 0x206, 0xA, 0x230, 0xA, 0x265, 0xA, 0x212, 0xA, 0x23E,
	0xB, 0x3AC, 0xB, 0x393, 0xB, 0x3E1, 0xA, 0x1DE, 0xB, 0x3D6, 0xA, 0x31D,
	0xB, 0x3E5, 0xB, 0x3E4, 0xA, 0x207, 0xB, 0x3C7, 0xA, 0x277, 0xB, 0x3D4,
	0x8, 0x0C0, 0xA, 0x162, 0xA, 0x3DA, 0xA, 0x124, 0xA, 0x1B4, 0xA, 0x264,
	0xA, 0x33D, 0xA, 0x1D1, 0xA, 0x1AF, 0xA, 0x39E, 0xA, 0x24F, 0xB, 0x373,
	0xA, 0x249, 0xB, 0x372, 0x9, 0x167, 0xA, 0x210, 0xA, 0x23A, 0xA, 0x1B8,
	0xB, 0x3AF, 0xA, 0x18E, 0xA, 0x2EC, 0x7, 0x062, 0x4, 0x00D,
}

// HuffmanEncodePacket encodes the bytes of in as an Ultima Online packet and
// appends it to out until in is closed or error is encountered (like EOF).
func HuffmanEncodePacket(in io.Reader, out io.Writer) error {
	var outBuf, outBufLength uint32
	var b [1]byte

	// Write all data code points
	for {
		if _, err := in.Read(b[:]); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		tableIdx := uint32(b[0]) * 2
		codeLength := huffmanTable[tableIdx]
		codePoint := huffmanTable[tableIdx+1]
		outBuf <<= codeLength
		outBuf |= uint32(codePoint)
		outBufLength += uint32(codeLength)
		for outBufLength >= 8 {
			outBufLength -= 8
			b[0] = byte(outBuf >> outBufLength)
			if _, err := out.Write(b[:]); err != nil {
				return err
			}
		}
	}

	// Write the terminal code
	tableIdx := 0x100 * 2
	codeLength := huffmanTable[tableIdx]
	codePoint := huffmanTable[tableIdx+1]
	outBuf <<= codeLength
	outBuf |= uint32(codePoint)
	outBufLength += uint32(codeLength)
	for outBufLength >= 8 {
		outBufLength -= 8
		b[0] = byte(outBuf >> outBufLength)
		if _, err := out.Write(b[:]); err != nil {
			return err
		}
	}

	// Zero-pad
	if outBufLength != 0 {
		b[0] = byte(outBuf << (8 - outBufLength))
		if _, err := out.Write(b[:]); err != nil {
			return err
		}
	}
	return nil
}

// Table and decompression algortihm documented here:
// https://raw.githubusercontent.com/andreakarasho/ClassicUO/master/src/Network/Huffman.cs
var huffmanDecodeTable = [][]int{
	{2, 1}, {4, 3}, {0, 5}, {7, 6}, {9, 8}, {11, 10}, {13, 12}, {14, -256},
	{16, 15}, {18, 17}, {20, 19}, {22, 21}, {23, -1}, {25, 24}, {27, 26}, {29, 28},
	{31, 30}, {33, 32}, {35, 34}, {37, 36}, {39, 38}, {-64, 40}, {42, 41}, {44, 43},
	{45, -6}, {47, 46}, {49, 48}, {51, 50}, {52, -119}, {53, -32}, {-14, 54}, {-5, 55},
	{57, 56}, {59, 58}, {-2, 60}, {62, 61}, {64, 63}, {66, 65}, {68, 67}, {70, 69},
	{72, 71}, {73, -51}, {75, 74}, {77, 76}, {-111, -101}, {-97, -4}, {79, 78}, {80, -110},
	{-116, 81}, {83, 82}, {-255, 84}, {86, 85}, {88, 87}, {90, 89}, {-10, -15}, {92, 91},
	{93, -21}, {94, -117}, {96, 95}, {98, 97}, {100, 99}, {101, -114}, {102, -105}, {103, -26},
	{105, 104}, {107, 106}, {109, 108}, {111, 110}, {-3, 112}, {-7, 113}, {-131, 114}, {-144, 115},
	{117, 116}, {118, -20}, {120, 119}, {122, 121}, {124, 123}, {126, 125}, {128, 127}, {-100, 129},
	{-8, 130}, {132, 131}, {134, 133}, {135, -120}, {-31, 136}, {138, 137}, {-234, -109}, {140, 139},
	{142, 141}, {144, 143}, {145, -112}, {146, -19}, {148, 147}, {-66, 149}, {-145, 150}, {-65, -13},
	{152, 151}, {154, 153}, {155, -30}, {157, 156}, {158, -99}, {160, 159}, {162, 161}, {163, -23},
	{164, -29}, {165, -11}, {-115, 166}, {168, 167}, {170, 169}, {171, -16}, {172, -34}, {-132, 173},
	{-108, 174}, {-22, 175}, {-9, 176}, {-84, 177}, {-37, -17}, {178, -28}, {180, 179}, {182, 181},
	{184, 183}, {186, 185}, {-104, 187}, {-78, 188}, {-61, 189}, {-178, -79}, {-134, -59}, {-25, 190},
	{-18, -83}, {-57, 191}, {192, -67}, {193, -98}, {-68, -12}, {195, 194}, {-128, -55}, {-50, -24},
	{196, -70}, {-33, -94}, {-129, 197}, {198, -74}, {199, -82}, {-87, -56}, {200, -44}, {201, -248},
	{-81, -163}, {-123, -52}, {-113, 202}, {-41, -48}, {-40, -122}, {-90, 203}, {204, -54}, {-192, -86},
	{206, 205}, {-130, 207}, {208, -53}, {-45, -133}, {210, 209}, {-91, 211}, {213, 212}, {-88, -106},
	{215, 214}, {217, 216}, {-49, 218}, {220, 219}, {222, 221}, {224, 223}, {226, 225}, {-102, 227},
	{228, -160}, {229, -46}, {230, -127}, {231, -103}, {233, 232}, {234, -60}, {-76, 235}, {-121, 236},
	{-73, 237}, {238, -149}, {-107, 239}, {240, -35}, {-27, -71}, {241, -69}, {-77, -89}, {-118, -62},
	{-85, -75}, {-58, -72}, {-80, -63}, {-42, 242}, {-157, -150}, {-236, -139}, {-243, -126}, {-214, -142},
	{-206, -138}, {-146, -240}, {-147, -204}, {-201, -152}, {-207, -227}, {-209, -154}, {-254, -153}, {-156, -176},
	{-210, -165}, {-185, -172}, {-170, -195}, {-211, -232}, {-239, -219}, {-177, -200}, {-212, -175}, {-143, -244},
	{-171, -246}, {-221, -203}, {-181, -202}, {-250, -173}, {-164, -184}, {-218, -193}, {-220, -199}, {-249, -190},
	{-217, -230}, {-216, -169}, {-197, -191}, {243, -47}, {245, 244}, {247, 246}, {-159, -148}, {249, 248},
	{-93, -92}, {-225, -96}, {-95, -151}, {251, 250}, {252, -241}, {-36, -161}, {254, 253}, {-39, -135},
	{-124, -187}, {-251, 255}, {-238, -162}, {-38, -242}, {-125, -43}, {-253, -215}, {-208, -140}, {-235, -137},
	{-237, -158}, {-205, -136}, {-141, -155}, {-229, -228}, {-168, -213}, {-194, -224}, {-226, -196}, {-233, -183},
	{-167, -231}, {-189, -174}, {-166, -252}, {-222, -198}, {-179, -188}, {-182, -223}, {-186, -180}, {-247, -245},
}

// HuffmanDecodePacket decodes the bytes of in as Huffman-encoded Ultima Online
// packet and appends it to out. ErrIncompletePacket is returned on a
// fragmented packet.
func HuffmanDecodePacket(in io.ByteReader, out io.ByteWriter) error {
	node := 0
	bitNum := 8
	var inByte byte
	var err error

	for {
		if bitNum == 8 {
			inByte, err = in.ReadByte()
			if err != nil {
				// Here an EOF would be unexpected
				return ErrIncompletePacket
			}
		}
		leaf := (inByte >> (bitNum - 1)) & 0x01
		leafValue := huffmanDecodeTable[node][leaf]

		// Halt codeword
		if leafValue == -256 {
			return nil
		}

		// Data codeword
		if leafValue < 1 {
			if err := out.WriteByte(byte(-leafValue)); err != nil {
				return err
			}
			leafValue = 0
		}

		// Bit step
		bitNum--
		node = leafValue

		// Byte step
		if bitNum < 1 {
			bitNum = 8
		}
	}
}
