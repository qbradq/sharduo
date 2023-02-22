package game

import (
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

type ContextMenu serverpacket.ContextMenu

// Append appends an entry to the context menu and returns the new context menu.
func (m *ContextMenu) Append(handler string, cl uo.Cliloc) {
	(*serverpacket.ContextMenu)(m).Add(eventIndexGetter(handler), cl)
}

// IsEmpty returns true if the context menu has no entries.
func (m *ContextMenu) IsEmpty() bool {
	return len((*serverpacket.ContextMenu)(m).Entries) == 0
}

// ContextMenuEntry abstracts an entry for a ContextMenu.
type ContextMenuEntry struct {
	// Event name to execute
	Event string
	// Cliloc of the label, must be in the range 3_000_000 - 3_060_000 inclusive
	Cliloc uo.Cliloc
}
