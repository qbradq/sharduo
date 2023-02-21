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
