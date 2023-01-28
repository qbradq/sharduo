package file

import (
	"encoding/binary"

	"github.com/qbradq/sharduo/lib/uo"
)

// StaticsMul holds the data for "statics0.mul"
type StaticsMul struct {
	mul     *IndexedMul
	statics []uo.Static
}

// NewStaticsMulFromFile loads "statics0.mul" and "staidx0.mul", or nil on error
// tiledata.mul must be loaded first and passed in so we can link static
// definition pointers
func NewStaticsMulFromFile(staidxPath, staticsPath string, tdmul *TileDataMul) *StaticsMul {
	m := &StaticsMul{}
	m.mul = NewIndexedMulFromFile(staidxPath, staticsPath)
	if m.mul == nil {
		return nil
	}
	chunkIdx := 0
	for cx := 0; cx < uo.MapChunksWidth; cx++ {
		for cy := 0; cy < uo.MapChunksHeight; cy++ {
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

				e := uo.NewStatic(
					uo.Location{
						X: int16(cx*uo.ChunkWidth) + int16(cd[staticOfs+2]),
						Y: int16(cy*uo.ChunkHeight) + int16(cd[staticOfs+3]),
						Z: int8(cd[staticOfs+4]),
					},
					&tdmul.staticDefinitions[binary.LittleEndian.Uint16(cd[staticOfs+0:staticOfs+2])])
				staticOfs += 7
				m.statics = append(m.statics, e)
			}
		}
	}
	return m
}

// Statics returns the internal slice of static definitions
func (m *StaticsMul) Statics() []uo.Static {
	return m.statics
}
