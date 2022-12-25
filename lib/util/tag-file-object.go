package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/lib/uo"
)

// TagFileObject is the intermediate representation of an object in tag file
// format.
type TagFileObject struct {
	t    string
	p    map[string]string
	errs []error
}

// NewTagFileObject creates a new TagFileObject for the given tag file reader.
func NewTagFileObject() *TagFileObject {
	return &TagFileObject{
		p: make(map[string]string),
	}
}

// TypeName returns the type name of the object described
func (o *TagFileObject) TypeName() string {
	return o.t
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

// HasErrors returns true if the object has encountered any errors.
func (o *TagFileObject) HasErrors() bool {
	return len(o.errs) > 0
}

// Errors returns a slice of all of the errors encountered by this object.
func (o *TagFileObject) Errors() []error {
	if len(o.errs) > 0 {
		return o.errs
	}
	return nil
}

// InjectError injects the error into this object's error slice. This is used
// by higher-level data loading functions to report out errors without panicing.
func (o *TagFileObject) InjectError(err error) {
	o.errs = append(o.errs, err)
}

// Map executes fn for every key/value pair in the object.
func (o *TagFileObject) Map(fn func(name, value string) error) []error {
	var errs []error
	for k, v := range o.p {
		if err := fn(k, v); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Set adds or overwrites the raw string associated with the name
func (o *TagFileObject) Set(name, value string) {
	o.p[name] = value
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
// found. This function may add errors to the internal error slice.
func (o *TagFileObject) GetNumber(name string, def int) int {
	if v, found := o.p[name]; found {
		n, err := strconv.ParseInt(v, 0, 32)
		if err != nil {
			o.errs = append(o.errs, fmt.Errorf("error in GetNumber %s=%s:%s", name, v, err))
			return def
		}
		return int(n)
	}
	return def
}

// GetHex returns the named property as an unsigned number or the default if not
// found. This function may add errors to the internal error slice.
func (o *TagFileObject) GetHex(name string, def uint32) uint32 {
	if v, found := o.p[name]; found {
		n, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			o.errs = append(o.errs, fmt.Errorf("error in GetHex %s=%s:%s", name, v, err))
			return def
		}
		return uint32(n)
	}
	return def
}

// GetBool returns the named property as a boolean value or the default if not
// found. This function may add errors to the internal error slice.
func (o *TagFileObject) GetBool(name string, def bool) bool {
	if v, found := o.p[name]; found {
		// This is the naked boolean case
		if v == "" {
			return true
		}
		var b bool
		var err error
		if b, err = strconv.ParseBool(v); err != nil {
			o.errs = append(o.errs, fmt.Errorf("error in GetBool %s=%s:%s", name, v, err))
			return def
		}
		return b
	}
	return def
}

// GetObjectReferences returns a slice of uo.Serial values. nil is the default
// value. This function may add errors to the internal error slice.
func (o *TagFileObject) GetObjectReferences(name string) []uo.Serial {
	if v, found := o.p[name]; found {
		parts := strings.Split(v, ",")
		ret := make([]uo.Serial, 0, len(parts))
		for idx, str := range parts {
			n, err := strconv.ParseInt(str, 0, 32)
			if err != nil {
				o.errs = append(o.errs, fmt.Errorf("error in GetObjectReferences at index %d %s=%s:%s", idx, name, v, err))
			} else {
				ret = append(ret, uo.Serial(n))
			}
		}
		return ret
	}
	return nil
}

// GetLocation returns a uo.Location value. The default value is returned if the
// named tag is not found.
func (o *TagFileObject) GetLocation(name string, def uo.Location) uo.Location {
	if v, found := o.p[name]; found {
		parts := strings.Split(v, ",")
		if len(parts) != 3 {
			o.errs = append(o.errs, fmt.Errorf("GetLocation(%s) did not find three values", name))
			return def
		}
		hasErrors := false
		var l uo.Location
		v, err := strconv.ParseInt(parts[0], 0, 32)
		if err != nil {
			o.errs = append(o.errs, err)
			hasErrors = true
		}
		l.X = int(v)
		v, err = strconv.ParseInt(parts[1], 0, 32)
		if err != nil {
			o.errs = append(o.errs, err)
			hasErrors = true
		}
		l.Y = int(v)
		v, err = strconv.ParseInt(parts[2], 0, 32)
		if err != nil {
			o.errs = append(o.errs, err)
			hasErrors = true
		}
		l.Z = int(v)
		if hasErrors {
			return def
		}
		return l
	}
	return def
}

// GetBounds returns a uo.Bounds value. The default value is returned if the
// named tag is not found.
func (o *TagFileObject) GetBounds(name string, def uo.Bounds) uo.Bounds {
	if v, found := o.p[name]; found {
		parts := strings.Split(v, ",")
		if len(parts) != 4 {
			o.errs = append(o.errs, fmt.Errorf("GetBounds(%s) did not find four values", name))
			return def
		}
		hasErrors := false
		var b uo.Bounds
		v, err := strconv.ParseInt(parts[0], 0, 32)
		if err != nil {
			o.errs = append(o.errs, err)
			hasErrors = true
		}
		b.X = int(v)
		v, err = strconv.ParseInt(parts[1], 0, 32)
		if err != nil {
			o.errs = append(o.errs, err)
			hasErrors = true
		}
		b.Y = int(v)
		v, err = strconv.ParseInt(parts[2], 0, 32)
		if err != nil {
			o.errs = append(o.errs, err)
			hasErrors = true
		}
		b.W = int(v)
		v, err = strconv.ParseInt(parts[3], 0, 32)
		if err != nil {
			o.errs = append(o.errs, err)
			hasErrors = true
		}
		b.H = int(v)
		if hasErrors {
			return def
		}
		return b
	}
	return def
}
