package uo

// A Body is a 16-bit value that describes the set of animations to use for a
// mobile. Body values used by UO range 1-999.
type Body uint16

// Pre-defined values for Body
const (
	BodyNone    Body = 0
	BodyDefault Body = 999
	BodySystem  Body = 0x7fff
)

var bodies = map[string]Body{
	"h_male":   400,
	"h_female": 401,
}

// GetBody returns the named known body ID or the default body.
func GetBody(name string) Body {
	if body, ok := bodies[name]; ok {
		return body
	}
	return BodyDefault
}

// GetHumanBody returns the proper human body code
func GetHumanBody(female bool) Body {
	if female {
		return bodies["h_female"]
	}
	return bodies["h_male"]
}
