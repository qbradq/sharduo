package game

// Mobile represents one mobile in the game.
type Mobile struct {
	// Display name of the mobile
	Name string
}

// NewMobile creates a new Mobile for use.
func NewMobile() *Mobile {
	return &Mobile{}
}
