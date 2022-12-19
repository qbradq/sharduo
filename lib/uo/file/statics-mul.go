package file

import (
	"encoding/binary"
	"log"

	"github.com/qbradq/sharduo/lib/uo"
)

// StaticsMulEntry describes a single static placement
type StaticsMulEntry struct {
	// Graphic ID of the static
	Graphic uo.Graphic
	// Absolute map position of the static
	Location uo.Location
}

// StaticsMul holds the data for "statics0.mul"
type StaticsMul struct {
	mul     *IndexedMul
	statics []StaticsMulEntry
}

// NewStaticsMulFromFile loads "statics0.mul" and "staidx0.mul", or nil on error
func NewStaticsMulFromFile(staidxPath, staticsPath string) *StaticsMul {
	m := &StaticsMul{}
	m.mul = NewIndexedMulFromFile(staidxPath, staticsPath)
	if m.mul == nil {
		return nil
	}
	chunkIdx := 0
	for cx := 0; cx < uo.MapChunksWidth; cx++ {
		for cy := 0; cy < uo.MapChunksHeight; cy++ {
			if cx == 1324/uo.ChunkWidth && cy == 1624/uo.ChunkHeight {
				log.Println("debug hook")
			}
			cd := m.mul.GetSegment(chunkIdx)
			chunkIdx++
			if cd == nil {
				continue
			}
			staticOfs := 0
			for {
				if staticOfs >= len(cd) {
					break
				}
				e := StaticsMulEntry{
					Graphic: uo.Graphic(binary.LittleEndian.Uint16(cd[staticOfs+0 : staticOfs+4])),
					Location: uo.Location{
						X: (int(cx) * int(uo.ChunkWidth)) + int(cd[staticOfs+4]),
						Y: (int(cy) * int(uo.ChunkHeight)) + int(cd[staticOfs+5]),
						Z: int(int8(cd[staticOfs+6])),
					},
				}
				staticOfs += 7
				m.statics = append(m.statics, e)
			}
		}
	}
	return m
}

// Statics returns the internal slice of static definitions
func (m *StaticsMul) Statics() []StaticsMulEntry {
	return m.statics
}
