package uod

import "strconv"

// CommandArgs is a thin wrapper around a slice of command line arguments
type CommandArgs []string

// String returns argument n as a string value, or the default
func (c CommandArgs) String(n int) string {
	if n < 0 || n >= len(c)-1 {
		return ""
	}
	return c[n+1]
}

// Int returns argument n as an integer value, or the default
func (c CommandArgs) Int(n int) int {
	is := c.String(n)
	if is == "" {
		return 0
	}
	ret, err := strconv.ParseInt(is, 0, 32)
	if err != nil {
		return 0
	}
	return int(ret)
}
