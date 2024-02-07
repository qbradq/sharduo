package game

import (
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Object holds all of the common dynamic data for all items and mobiles.
type Object struct {
	// Static variables
	BaseTemplate       string               // Name of the template this template inherited from
	Name               string               // Descriptive name without articles
	ArticleA           bool                 // If true the article A is used to refer to this item
	ArticleAn          bool                 // If true the article An is used to refer to this item
	Events             map[string]string    // Raw event names
	PostCreationEvents []*postCreationEvent // List of events to execute after creation
	ContextMenu        []ctxMenuEntry
	// Persistent variables
	TemplateName string       // Name of the template used to create the item
	Serial       uo.Serial    // Unique serial of the object
	Location     uo.Point     // Location of the object
	Facing       uo.Direction // Current facing of the object
	Hue          uo.Hue
	// Transient values
	Removed    bool                    // If true the object is slated for removal from the game
	NoRent     bool                    // If true the object will not be persisted to backup saves
	Spawner    Spawner                 // The spawner managing this object if any
	Visibility uo.Visibility           // Current visibility state of the object
	opl        *serverpacket.OPLPacket // Cached OPLPacket
	oplInfo    *serverpacket.OPLInfo   // Cached OPLInfo
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
