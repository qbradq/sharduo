package util

import (
	"errors"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/lib/uo"
)

// TagFileObject is the intermediate representation of an object in tag file
// format.
type TagFileObject struct {
	t string
	p map[string]string
}

// NewTagFileObject creates a new TagFileObject for the given tag file reader.
func NewTagFileObject() *TagFileObject {
	return &TagFileObject{
		p: make(map[string]string),
	}
}

// stripLine strips a line of whitespace and returns it.
func stripLine(line string) string {
	return strings.TrimSpace(line)
}

// isCommentLine returns true if the line given looks like a comment.
func isCommentLine(line string) bool {
	return len(line) == 0 || line[0] == ';'
}

// isTypeLine returns true if the line given looks like a object type.
func isTypeLine(line string) bool {
	return len(line) > 0 && line[0] == '[' && line[len(line)-1] == ']'
}

// extractObjectType returns the type from an object line.
func extractObjectType(line string) string {
	if len(line) < 3 {
		return ""
	}
	return line[1 : len(line)-1]
}

// TypeName returns the type name of the object described
func (o *TagFileObject) TypeName() string {
	return o.t
}

// HandleTypeLine handles the type line for the next object.
func (o *TagFileObject) HandleTypeLine(line string) error {
	if !isTypeLine(line) {
		return errors.New("syntax error at type line")
	}
	o.t = extractObjectType(line)
	return nil
}

// HandlePropertyLine attempts to handle a single property line in the file.
func (o *TagFileObject) HandlePropertyLine(line string) error {
	parts := strings.SplitN(line, "=", 2)
	var key, value string
	if len(parts) == 1 {
		key = strings.TrimSpace(parts[0])
		value = ""
	} else if len(parts) == 2 {
		key = strings.TrimSpace(parts[0])
		value = strings.TrimSpace(parts[1])
	} else if len(parts) != 2 {
		return errors.New("syntax error")
	}
	o.p[key] = value
	return nil
}

// Delete removes the named property from the object
func (o *TagFileObject) Delete(name string) {
	delete(o.p, name)
}

// GetString returns the named property as a string or the default if not
// found.
func (o *TagFileObject) GetString(name, def string) string {
	if v, found := o.p[name]; found {
		return v
	}
	return def
}

// GetNumber returns the named property as a number or the default if not
// found.
func (o *TagFileObject) GetNumber(name string, def int) int {
	if v, found := o.p[name]; found {
		n, err := strconv.ParseInt(v, 0, 32)
		if err != nil {
			panic(err)
		}
		return int(n)
	}
	return def
}

// GetBool returns the named property as a boolean value or the default if not
// found.
func (o *TagFileObject) GetBool(name string, def bool) bool {
	if v, found := o.p[name]; found {
		// This is the naked boolean case
		if v == "" {
			return true
		}
		var b bool
		var err error
		if b, err = strconv.ParseBool(v); err != nil {
			return def
		}
		return b
	}
	return def
}

// GetObjectReferences returns a slice of uo.Serial values. nil is the default
// value.
func (o *TagFileObject) GetObjectReferences(name string) []uo.Serial {
	if v, found := o.p[name]; found {
		parts := strings.Split(v, ",")
		ret := make([]uo.Serial, 0, len(parts))
		for _, str := range parts {
			n, err := strconv.ParseInt(str, 0, 32)
			if err != nil {
				panic(err)
			}
			ret = append(ret, uo.Serial(n))
		}
		return ret
	}
	return nil
}
