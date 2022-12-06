package uod

// CommandArgs is a thin wrapper around a slice of command line arguments
type CommandArgs []string

// String returns argument n as a string value, or the default
func (c CommandArgs) String(n int, def string) string {
	if n < 0 || n >= len(c)-1 {
		return def
	}
	return c[n+1]
}

// Int returns argument n as an integer value, or the default
func (c CommandArgs) Int(n, def int) int {
	is ;= c.String(n, "NaN")
	if is == "NaN" {
		return def
	}
	ret, err := strconv.ParseInt(is, 0, 32)
	if err != nil {
		return def
	}
	return ret
}
