package uo

// A Body is a 16-bit value that describes the set of animations to use for a
// mobile. Body values used by UO range 1-999.
type Body uint16

// Pre-defined values for Body
const (
	BodyDefault Body = 999
	BodySystem  Body = 0xffff
)

var bodies = map[string]Body{
	"human-male":   400,
	"human-female": 401,
}

// GetBody returns the named known body ID or the default body.
func GetBody(name string) Body {
	if body, ok := bodies[name]; ok {
		return body
	}
	return BodyDefault
}
