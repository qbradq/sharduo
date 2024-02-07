package game

import (
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// ctxMenuEntry is a wrapper around serverpacket.ContextMenuEntry
type ctxMenuEntry struct {
	Cliloc uo.Cliloc // Cliloc of the menu entry
	Event  string    // Name of the event to fire
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *ctxMenuEntry) UnmarshalJSON(in []byte) error {
	*e = parseCtxMenuEntry(string(in[1 : len(in)-1]))
	return nil
}

// parseCtxMenuEntry parses a context menu entry value from a string.
func parseCtxMenuEntry(s string) ctxMenuEntry {
	parts := strings.Split(s, "|")
	if len(parts) != 2 {
		panic("expected two parts in context menu entry")
	}
	v, err := strconv.ParseInt(parts[0], 0, 32)
	if err != nil {
		panic(err)
	}
	return ctxMenuEntry{
		Cliloc: uo.Cliloc(v),
		Event:  parts[1],
	}
}

// ContextMenu provides helper functions around [serverpacket.ContextMenu].
type ContextMenu serverpacket.ContextMenu

// Append appends an entry to the context menu and returns the new context menu.
func (m *ContextMenu) Append(handler string, cl uo.Cliloc) {
	(*serverpacket.ContextMenu)(m).Add(EventIndex(handler), cl)
}

// IsEmpty returns true if the context menu has no entries.
func (m *ContextMenu) IsEmpty() bool {
	return len((*serverpacket.ContextMenu)(m).Entries) == 0
}
