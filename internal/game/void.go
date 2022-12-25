package game

// VoidObject is an object that blindly accepts object adds and removes and does
// not track them. If an object is left parented to a Void object it will leak.
type VoidObject struct {
	BaseObject
}

// Global instance of the void object
var TheVoid *VoidObject = &VoidObject{
	BaseObject: BaseObject{
		name: "the void",
	},
}

// RemoveObject implements the Object interface
func (o *VoidObject) RemoveObject(c Object) bool {
	return true
}

// AddObject implements the Object interface
func (o *VoidObject) AddObject(c Object) bool {
	return true
}
