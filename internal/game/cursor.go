package game

import "github.com/qbradq/sharduo/lib/uo"

// CursorState represents the current state of the client's cursor from the
// player's point of view. Not that this does not account for targeting cursors.
type CursorState int

// CursorState constants
const (
	CursorStateNormal CursorState = 0
	CursorStatePickUp CursorState = 1
	CursorStateDrop   CursorState = 2
	CursorStateEquip  CursorState = 3
	CursorStateReturn CursorState = 4
)

// Cursor represents a mobile's cursor
type Cursor struct {
	// The current state of the cursor
	State CursorState
	// Item on the cursor
	item Item
	// Previous parent of the object on the cursor before we picked it up
	previousParent Object
	// Previous location of the object
	previousLocation uo.Location
}

// Occupied returns true if there is already an item on the cursor
func (c *Cursor) Occupied() bool { return c.item != nil }

// PickUp attempts to pick up the object. Returns true if successful.
func (c *Cursor) PickUp(o Object) bool {
	if o == nil {
		c.previousLocation = uo.Location{}
		c.State = CursorStateNormal
		c.item = nil
		c.previousParent = nil
		return true
	}
	c.State = CursorStatePickUp
	if c.item != nil {
		return c.item.Serial() == o.Serial()
	}
	item, ok := o.(Item)
	if !ok {
		return false
	}
	c.previousParent = item.Parent()
	c.item = item
	c.previousLocation = item.Location()
	return true
}

// Return send the item on the cursor back to it's previous parent.
func (c *Cursor) Return() {
	oldLocation := c.previousLocation
	oldParent := c.previousParent
	item := c.item
	c.previousLocation = uo.Location{}
	c.previousParent = nil
	c.item = nil
	item.SetLocation(oldLocation)
	// Old parent was the map, just send it back
	if oldParent == nil {
		if !world.Map().AddObject(item) {
			world.Map().ForceAddObject(item)
		}
		return
	}
	// Old parent was a stack of items that we might be able to combine with
	if oldItemParent, ok := oldParent.(Item); ok && oldItemParent.CanCombineWith(item) {
		if oldItemParent.Combine(item) {
			// Combine successful
			return
		}
		// But the item back into the stack's parent
		if oldParent.Parent() == nil {
			if !world.Map().AddObject(item) {
				world.Map().ForceAddObject(item)
			}
		} else {
			if !oldParent.Parent().AddObject(item) {
				oldParent.Parent().ForceAddObject(item)
			}
		}
		return
	}
	if !oldParent.Parent().AddObject(item) {
		oldParent.Parent().ForceAddObject(item)
	}
}

// Drop attempts to drop the item on the cursor onto the target and returns true
// if successful.
func (c *Cursor) Drop(target Object) bool {
	c.State = CursorStateDrop
	if !world.Map().SetNewParent(c.item, target) {
		c.State = CursorStateReturn
		return false
	}
	c.State = CursorStateNormal
	return true
}

// Wear attempts to wear the item on the cursor onto the target and returns true
// if successful.
func (c *Cursor) Wear(target Object) bool {
	c.State = CursorStateEquip
	if !world.Map().SetNewParent(c.item, target) {
		c.State = CursorStateReturn
		return false
	}
	c.State = CursorStateNormal
	return true
}

// Item returns the item currently in the cursor, or nil if none
func (c *Cursor) Item() Item { return c.item }
