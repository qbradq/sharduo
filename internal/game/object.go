package game

import (
	"io"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// Object holds all of the common dynamic data for all items and mobiles.
type Object struct {
	// Static variables
	Name      string            // Descriptive name without articles
	ArticleA  bool              // If true the article A is used to refer to this item
	ArticleAn bool              // If true the article An is used to refer to this item
	Events    map[string]string // Raw event names
	// Persistent variables
	Serial   uo.Serial    // Unique serial of the object
	Location uo.Point     // Location of the object
	Facing   uo.Direction // Current facing of the object
	Hue      uo.Hue
	// Transient values
	Removed    bool          // If true the object is slated for removal from the game
	NoRent     bool          // If true the object will not be persisted to backup saves
	Spawner    Spawner       // The spawner managing this object if any
	Visibility uo.Visibility // Current visibility state of the object
}

// Write writes the object's persistent variables to w.
func (o *Object) Write(w io.Writer) {
	util.PutUInt32(w, 0)                // Version
	util.PutUInt32(w, uint32(o.Serial)) // Serial
	util.PutPoint(w, o.Location)        // Location
	util.PutByte(w, byte(o.Facing))     // Facing
	util.PutUInt16(w, uint16(o.Hue))    // Hue
}

// Read reads the object's persistent variables from r.
func (o *Object) Read(r io.Reader) {
	_ = util.GetUInt32(r)                    // Version
	o.Serial = uo.Serial(util.GetUInt32(r))  // Serial
	o.Location = util.GetPoint(r)            // Location
	o.Facing = uo.Direction(util.GetByte(r)) // Facing
	o.Hue = uo.Hue(util.GetUInt16(r))        // Hue
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

// ExecuteEvent executes the named event handler if any is configured. Returns
// true if the handler was found and also returned true.
func (o *Object) ExecuteEvent(which string, s, v any) bool {
	hn, ok := o.Events[which]
	if !ok {
		return false
	}
	return ExecuteEventHandler(hn, o, s, v)
}
