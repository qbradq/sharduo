package game

import "github.com/qbradq/sharduo/lib/uo"

// Static represents a static object on the map
type Static struct {
	// Graphic of the item
	Graphic uo.Graphic
	// Absolute location
	Location uo.Location
}
