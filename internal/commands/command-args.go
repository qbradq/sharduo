package commands

import "strconv"

// CommandArgs is a thin wrapper around a slice of command line arguments
type CommandArgs []string

// String returns argument n as a string value, or the default
func (c CommandArgs) String(n int) string {
	return c[n]
}

// Int returns argument n as an integer value, or the default
func (c CommandArgs) Int(n int) int {
	if n < 0 || n >= len(c) {
		return 0
	}
	ret, err := strconv.ParseInt(c[n], 0, 32)
	if err != nil {
		return 0
	}
	return int(ret)
}
