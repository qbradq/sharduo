package game

import "github.com/qbradq/sharduo/lib/uo"

// Error represents an error with UO communications details
type Error struct {
	Cliloc        uo.Cliloc // Cliloc number of the error message, if any
	ClilocStrings []string  // Strings to inject into the cliloc message
	String        string    // Error string
}

// SendTo sends the error message to the given net state
func (e *Error) SendTo(n NetState, from Object) {
	if e.Cliloc != uo.Cliloc(0) {
		n.Cliloc(from, e.Cliloc, e.ClilocStrings...)
	} else {
		n.Speech(from, e.String)
	}
}
