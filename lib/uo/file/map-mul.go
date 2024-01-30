package file

import (
	"encoding/binary"

	"github.com/qbradq/sharduo/lib/uo"
)

// MapMulChunk represents one chunk of the map from map0.mul
type MapMulChunk struct {
	Tiles []uo.Tile
}

// MapMul represents the Map0.mul file
type MapMul struct {
	Chunks []MapMulChunk
}

// NewMapMulFromFile loads map0.mul from the given path. Not that this ONLY
// works for map0.mul/map1.mul without diffs. The tile data mul must be loaded
// prior to this and passed in so we can perform tile linkage.
func NewMapMulFromFile(fname string, tdmul *TileDataMul) *MapMul {
	// Initialize map storage
	m := &MapMul{
		Chunks: make([]MapMulChunk, uo.MapChunksWidth*uo.MapChunksHeight),
	}
	for i := range m.Chunks {
		m.Chunks[i] = MapMulChunk{
			Tiles: make([]uo.Tile, uo.ChunkWidth*uo.ChunkHeight),
		}
	}
	// Load the mul and do sanity checks
	sm := NewStaticMulFromFile(fname, 196, 0)
	if sm.NumberOfSegments() != uo.MapChunksWidth*uo.MapChunksHeight {
		return nil
	}
	// Load all chunks
	iseg := 0
	for cx := 0; cx < uo.MapChunksWidth; cx++ {
		for cy := 0; cy < uo.MapChunksHeight; cy++ {
			seg := sm.GetSegment(iseg)
			iseg++
			chunk := m.Chunks[cy*uo.MapChunksWidth+cx]
			// Load all tiles in the chunk
			sofs := 4 // Each map chunk has a 4-byte header of unknown use
			for ty := 0; ty < uo.ChunkHeight; ty++ {
				for tx := 0; tx < uo.ChunkWidth; tx++ {
					tileIdx := binary.LittleEndian.Uint16(seg[sofs : sofs+2])
					z := int(int8(seg[sofs+2]))
					sofs += 3
					chunk.Tiles[ty*uo.ChunkWidth+tx] = uo.NewTile(z, tdmul.GetTileDefinition(int(tileIdx)))
				}
			}
		}
	}
	return m
}

// GetChunk returns a pointer to the given MapMulChunk, or nil if the
// chunk coordinates are out of bounds.
func (m *MapMul) GetChunk(x, y int) MapMulChunk {
	var zero MapMulChunk
	if x < 0 || x >= uo.MapChunksWidth || y < 0 || y >= uo.MapChunksHeight {
		return zero
	}
	return m.Chunks[y*uo.MapChunksWidth+x]
}

// GetTile returns the tile at the given coordinates, or the zero value if the
// coordinates are out of bounds.
func (m *MapMul) GetTile(x, y int) uo.Tile {
	c := m.GetChunk(x/uo.ChunkWidth, y/uo.ChunkHeight)
	return c.Tiles[(y%uo.ChunkHeight)*uo.ChunkWidth+(x%uo.ChunkWidth)]
}
