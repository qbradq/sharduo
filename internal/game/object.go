package game

import "github.com/qbradq/sharduo/lib/uo"

// Object holds all of the common dynamic data for all items and mobiles.
type Object struct {
	// Static variables
	Name      string // Descriptive name without articles
	ArticleA  bool   // If true the article A is used to refer to this item
	ArticleAn bool   // If true the article An is used to refer to this item
	// Persistent variables
	Serial   uo.Serial    // Unique serial of the object
	Location uo.Point     // Location of the object
	Facing   uo.Direction // Current facing of the object
	Hue      uo.Hue
	// Transient values
	Removed    bool          // If true the object is slated for removal from the game
	Visibility uo.Visibility // Current visibility state of the object
}

// DisplayName returns the normalized displayable name of the object.
func (o *Object) DisplayName() string {
	if o.ArticleA {
		return "a " + o.Name
	}
	if o.ArticleAn {
		return "an " + o.Name
	}
	return o.Name
}
