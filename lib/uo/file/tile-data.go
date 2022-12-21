package file

import (
	"encoding/binary"
	"os"

	"github.com/qbradq/sharduo/lib/uo"
)

// TileDataMul loads "tiledata.mul"
type TileDataMul struct {
	tileDefinitions   []*uo.TileDefinition
	staticDefinitions []*uo.StaticDefinition
}

// NewTileDataMul creates a new TileDataMul and loads it from the path
func NewTileDataMul(fname string) *TileDataMul {
	// Load the file
	d, err := os.ReadFile(fname)
	if err != nil {
		return nil
	}
	ret := &TileDataMul{
		tileDefinitions: make([]*uo.TileDefinition, 0x4000),
	}
	dofs := 0
	// Load tile definitions
	tilei := 0
	for tileDataChunk := 0; tileDataChunk < 512; tileDataChunk++ {
		dofs += 4 // Skip header
		for tileDataI := 0; tileDataI < 32; tileDataI++ {
			ret.tileDefinitions[tilei] = &uo.TileDefinition{
				Graphic:   uo.Graphic(tilei),
				TileFlags: uo.TileFlags(binary.LittleEndian.Uint64(d[dofs+0 : dofs+8])),
				Texture:   uo.Texture(binary.LittleEndian.Uint16(d[dofs+8 : dofs+10])),
				Name:      string(d[dofs+10 : dofs+30]),
			}
			tilei++
			dofs += 30
		}
	}
	// Load static definitions
	statici := 0
	for staticDataChunk := 0; dofs < len(d); staticDataChunk++ {
		dofs += 4
		for staticDataI := 0; staticDataI < 32; staticDataI++ {
			ret.staticDefinitions = append(ret.staticDefinitions, &uo.StaticDefinition{
				Graphic:   uo.Graphic(statici),
				TileFlags: uo.TileFlags(binary.LittleEndian.Uint64(d[dofs+0 : dofs+8])),
				Weight:    int(uint8(d[dofs+8])),
				Layer:     uo.Layer(d[dofs+9]),
				Count:     int(binary.LittleEndian.Uint32(d[dofs+10 : dofs+14])),
				Animation: uo.Animation(binary.LittleEndian.Uint16(d[dofs+14 : dofs+16])),
				Hue:       uo.Hue(binary.LittleEndian.Uint16(d[dofs+16 : dofs+18])),
				Light:     uo.Light(binary.LittleEndian.Uint16(d[dofs+18 : dofs+20])),
				Height:    int(d[dofs+20]),
				Name:      string(d[dofs+21 : dofs+41]),
			})
			statici++
			dofs += 41
		}
	}
	return ret
}

// GetTileDefinition returns the given tile definition, or the NoDraw definition
func (m *TileDataMul) GetTileDefinition(idx int) *uo.TileDefinition {
	if idx < 0 || idx >= 0x4000 {
		return m.tileDefinitions[2]
	}
	return m.tileDefinitions[idx]
}

// GetStaticDefinition returns the given static definition, or the NoDraw
// definition
func (m *TileDataMul) GetStaticDefinition(idx int) *uo.StaticDefinition {
	if idx < 0 || idx >= len(m.staticDefinitions) {
		return m.staticDefinitions[1]
	}
	return m.staticDefinitions[idx]
}
