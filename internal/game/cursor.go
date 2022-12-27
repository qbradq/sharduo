package game

// CursorState represents the current state of the client's cursor from the
// player's point of view. Not that this does not account for targeting cursors.
type CursorState int

// CursorState constants
const (
	CursorStateNormal CursorState = 0
	CursorStatePickUp CursorState = 1
	CursorStateDrop   CursorState = 2
	CursorStateWear   CursorState = 3
)

// Cursor represents a mobile's cursor
type Cursor struct {
	// The current state of the cursor
	State CursorState
	// Item on the cursor
	item Item
	// Previous parent of the object on the cursor before we picked it up
	previousParent Object
}

// Occupied returns true if there is already an item on the cursor
func (c *Cursor) Occupied() bool { return c.item != nil }

// PickUp attempts to pick up the object. Returns true if successful.
func (c *Cursor) PickUp(o Object) bool {
	if o == nil {
		c.item = nil
		c.previousParent = nil
		return true
	}
	if c.item != nil {
		if c.item == o {
			return true
		}
		return false
	}
	item, ok := o.(Item)
	if !ok {
		return false
	}
	c.previousParent = item.Parent()
	c.item = item
	return true
}

// Item returns the item currently in the cursor, or nil if none
func (c *Cursor) Item() Item { return c.item }
