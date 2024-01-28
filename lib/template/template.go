package template

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

var objctors = make(map[string]func() any)

// RegisterConstructor registers a constructor with the template package.
func RegisterConstructor(name string, fn func() any) {
	if _, duplicate := objctors[name]; duplicate {
		log.Fatalf("fatal: duplicate template constructor registered for name %s", name)
	}
	objctors[name] = fn
}

// GetConstructor returns the named constructor or nil.
func GetConstructor(name string) func() any {
	return objctors[name]
}

// Object is the interface all Template objects must implement.
type Object interface {
	// Serial returns the unique ID of the object.
	Serial() uo.Serial
	// Deserialize takes data from the template object and initializes the
	// object's data structures with it.
	Deserialize(*Template, bool)
	// RecalculateStats is called after Deserialize() and should be used to
	// recalculate any dynamic values of the data structures initialized by
	// Deserialize().
	RecalculateStats()
	// InsertObject adds an object as a child of this object through an empty
	// interface.
	InsertObject(any)
	// SetAmount sets the stack amount for the object, if any.
	SetAmount(int)
}

// Template contains all of the property lines of the template.
type Template struct {
	TypeName     string                 // Name of the object constructor used to create the object.
	TemplateName string                 // Unique name of the template.
	BaseTemplate string                 // Name of the base template. The empty string means a root template.
	IsResolved   bool                   // True if the template's inheritance chain has already been satisfied.
	properties   map[string]*expression // List of all property expressions
}

// NewTemplate creates a new template.T object from the provided TagFileObject. The
// inheritance chain has not been resolved for this object and the text template
// has not been created or compiled.
func NewTemplate(tfo *util.TagFileObject, tm *TemplateManager) (*Template, []error) {
	t := &Template{
		TypeName:   tfo.TypeName(),
		properties: make(map[string]*expression),
	}
	templateName := tfo.GetString("TemplateName", "")
	if templateName == "" {
		panic(fmt.Sprintf("template of type %s missing TemplateName field", tfo.TypeName()))
	}
	errs := tfo.Map(func(name, value string) error {
		if name == "TemplateName" {
			t.TemplateName = value
		} else if name == "BaseTemplate" {
			t.BaseTemplate = value
		} else {
			t.properties[name] = &expression{
				text: value,
			}
		}
		return nil
	})
	return t, errs
}

// compileExpressions prepares every expression for execution and must be done
// once before any values methods are called
func (t *Template) compileExpressions() {
	for _, e := range t.properties {
		if err := e.compile(); err != nil {
			log.Println(err)
		}
	}
}

// GetString returns the named property as a string or the default if not
// found. This function panics if no context is on the stack. See PushContext
// and PopContext.
func (t *Template) GetString(name, def string) string {
	e, ok := t.properties[name]
	if !ok {
		return def
	}
	v, err := e.execute()
	if err != nil {
		log.Println(err)
		return def
	}
	return v
}

// GetNumber returns the named property as a number or the default if not
// found. This function may add errors to the internal error slice.
func (t *Template) GetNumber(name string, def int) int {
	v := t.GetString(name, "")
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 0, 32)
	if err != nil {
		log.Println(err)
		return def
	}
	return int(n)
}

// GetULong returns the named property as a uint64 or the default if not found.
// This function may add errors to the internal error slice. Only use this for
// actual 64-bit values, like uo.Time.
func (t *Template) GetULong(name string, def uint64) uint64 {
	v := t.GetString(name, "")
	if v == "" {
		return def
	}
	n, err := strconv.ParseUint(v, 0, 64)
	if err != nil {
		log.Println(err)
		return def
	}
	return n
}

// GetFloat returns the named property as a float32 or the default if not
// found. This function may add errors to the internal error slice.
func (t *Template) GetFloat(name string, def float32) float32 {
	v := t.GetString(name, "")
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		log.Println(err)
		return def
	}
	return float32(f)
}

// GetHex returns the named property as an unsigned number or the default if not
// found. This function may add errors to the internal error slice.
func (t *Template) GetHex(name string, def uint32) uint32 {
	v := t.GetString(name, "")
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 0, 64)
	if err != nil {
		log.Println(err)
		return def
	}
	return uint32(n)
}

// GetBool returns the named property as a boolean value or the default if not
// found. This function may add errors to the internal error slice.
func (t *Template) GetBool(name string, def bool) bool {
	v := t.GetString(name, "~~NOT~FOUND~~")
	if v == "~~NOT~FOUND~~" {
		return def
	}
	// This is the naked boolean case
	if v == "" {
		return true
	}
	var b bool
	var err error
	if b, err = strconv.ParseBool(v); err != nil {
		log.Println(err)
		return def
	}
	return b
}

// GetObjectReferences returns a slice of uo.Serial values. nil is the default
// value. This function may add errors to the internal error slice.
func (t *Template) GetObjectReferences(name string) []uo.Serial {
	v := t.GetString(name, "")
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	ret := make([]uo.Serial, 0, len(parts))
	for _, str := range parts {
		n, err := strconv.ParseInt(str, 0, 32)
		if err != nil {
			log.Println(err)
		} else {
			ret = append(ret, uo.Serial(n))
		}
	}
	return ret
}

// GetLocation returns a uo.Location value. The default value is returned if the
// named tag is not found.
func (t *Template) GetLocation(name string, def uo.Point) uo.Point {
	str := t.GetString(name, "")
	if str == "" {
		return def
	}
	parts := strings.Split(str, ",")
	if len(parts) != 3 {
		log.Printf("error: GetLocation(%s) did not find three values", name)
		return def
	}
	hasErrors := false
	var l uo.Point
	v, err := strconv.ParseInt(parts[0], 0, 16)
	if err != nil {
		log.Println(err)
		hasErrors = true
	}
	l.X = int16(v)
	v, err = strconv.ParseInt(parts[1], 0, 16)
	if err != nil {
		log.Println(err)
		hasErrors = true
	}
	l.Y = int16(v)
	v, err = strconv.ParseInt(parts[2], 0, 8)
	if err != nil {
		log.Println(err)
		hasErrors = true
	}
	l.Z = int8(v)
	if hasErrors {
		return def
	}
	return l
}

// GetBounds returns a uo.Bounds value. The default value is returned if the
// named tag is not found.
func (t *Template) GetBounds(name string, def uo.Bounds) uo.Bounds {
	str := t.GetString(name, "")
	if str == "" {
		return def
	}
	parts := strings.Split(str, ",")
	if len(parts) != 4 {
		log.Printf("error: GetBounds(%s) did not find four values", name)
		return def
	}
	hasErrors := false
	var b uo.Bounds
	v, err := strconv.ParseInt(parts[0], 0, 16)
	if err != nil {
		log.Println(err)
		hasErrors = true
	}
	b.X = int16(v)
	v, err = strconv.ParseInt(parts[1], 0, 16)
	if err != nil {
		log.Println(err)
		hasErrors = true
	}
	b.Y = int16(v)
	v, err = strconv.ParseInt(parts[2], 0, 16)
	if err != nil {
		log.Println(err)
		hasErrors = true
	}
	b.W = int16(v)
	v, err = strconv.ParseInt(parts[3], 0, 16)
	if err != nil {
		log.Println(err)
		hasErrors = true
	}
	b.H = int16(v)
	if hasErrors {
		return def
	}
	return b
}
