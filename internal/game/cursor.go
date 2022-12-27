package game

// Cursor represents a mobile's cursor
type Cursor struct {
	// Object on the cursor
	object Object
	// Previous parent of the object on the cursor before we picked it up
	previousParent Object
}

// Occupied returns true if there is already an object on the cursor
func (c *Cursor) Occupied() bool { return c.object != nil }
