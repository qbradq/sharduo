package game

import "github.com/qbradq/sharduo/lib/uo"

// chunk manages the data for a single chunk of the map.
type chunk struct {
	Items   []*Item                                 // List of all items within the chunk
	Mobiles []*Mobile                               // List of all mobiles within the chunk
	Tiles   [uo.ChunkWidth * uo.ChunkHeight]uo.Tile // Tile matrix
	Statics []uo.Static                             // All statics within the chunk
}
