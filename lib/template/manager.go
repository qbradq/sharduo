package template

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	txtmp "text/template"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

const (
	MaxRecursion int = 16
)

// Global template manager instance
var tm *TemplateManager

// TemplateManager manages a collection of templates
type TemplateManager struct {
	ctxStack    []map[string]string                     // Stack of contexts used for creating new objects
	nextCtx     int                                     // Index of the next context to be popped
	defaultCtx  map[string]string                       // Default context used during save file loading
	templates   *util.Registry[string, *Template]       // Registry of templates in the manager
	pctp        *util.Registry[string, *txtmp.Template] // Registry of pre-compiled go templates
	lists       *util.Registry[string, []string]        // Collection of lists used by the random object creation methods
	rng         uo.RandomSource                         // Source of randomness for randomly generating things
	storeObject func(Object)                            // Adds the object to game datastores
}

// Initialize initializes the template package and must be called prior to all
// other package calls. The function parameter fn should add the game object to
// internal data stores when a new object is created by the template engine.
func Initialize(templatePath, listPath, variablesFilePath string, rng uo.RandomSource, fn func(Object)) []error {
	// Initialize the manager
	tm = &TemplateManager{
		ctxStack:    make([]map[string]string, MaxRecursion),
		defaultCtx:  make(map[string]string),
		templates:   util.NewRegistry[string, *Template]("templates"),
		pctp:        util.NewRegistry[string, *txtmp.Template]("template-pre-cache"),
		lists:       util.NewRegistry[string, []string]("template-lists"),
		rng:         rng,
		storeObject: fn,
	}
	// Load all templates
	errs := data.Walk(templatePath, tm.loadTemplateFile)
	if len(errs) > 0 {
		return errs
	}
	// Resolve template inheritance chains
	errs = tm.resolveInheritance()
	if len(errs) > 0 {
		return errs
	}
	// Load all lists
	errs = data.Walk(listPath, tm.loadListFile)
	if len(errs) > 0 {
		return errs
	}
	// Load all pre-defined variables for the entire context stack
	for i := range tm.ctxStack {
		tm.ctxStack[i] = make(map[string]string)
	}
	vd, err := data.FS.Open(variablesFilePath)
	if err != nil {
		return []error{err}
	}
	tfr := &util.TagFileReader{}
	tfr.StartReading(vd)
	tfo := tfr.ReadObject()
	if tfo == nil {
		return []error{errors.New("failed to read template variables list")}
	}
	tfo.Map(func(name, value string) error {
		tm.defaultCtx[name] = value
		for i := range tm.ctxStack {
			tm.ctxStack[i][name] = value
		}
		return nil
	})

	return nil
}

// FindTemplate returns a pointer to the named template or nil if not found.
func FindTemplate(which string) *Template {
	t, found := tm.templates.Get(which)
	if !found {
		log.Printf("warning: template %s not found\n", which)
		return nil
	}
	return t
}

// pushContext pushes a context onto the stack. This function panics on stack
// overflow.
func (m *TemplateManager) pushContext() {
	if m.nextCtx >= len(m.ctxStack) {
		panic("template manager context stack overflow")
	}
	m.nextCtx++
}

// popContext decrements the stack pointer by one. This function panics if the
// stack pointer is decremented below zero.
func (m *TemplateManager) popContext() {
	m.nextCtx--
	if m.nextCtx < 0 {
		panic("template manager context stack underflow")
	}
}

// CurrentContext returns the current template execution context. This function
// panics if the context stack is empty.
func (m *TemplateManager) CurrentContext() map[string]string {
	if m.nextCtx == 0 {
		return m.defaultCtx
	}
	return m.ctxStack[m.nextCtx-1]
}

// updateDynamicValues randomizes the values of all context variables with
// random or changing values.
func (m *TemplateManager) updateDynamicValues(ctx map[string]string) {
	// Inject dynamic values into the template context
	if tm.rng.RandomBool() {
		ctx["IsFemale"] = "true"
	} else {
		ctx["IsFemale"] = ""
	}
}

// GenerateObject returns a pointer to a newly constructed object as in Create()
// but the object is not added to the data stores.
func GenerateObject(which string) Object {
	// Bind the template
	t := FindTemplate(which)
	if t == nil {
		return nil
	}
	// Create the object
	ctor := GetConstructor(t.TypeName)
	if ctor == nil {
		log.Printf("error: constructor not found for type %s", t.TypeName)
		return nil
	}
	s := ctor().(Object)
	if s == nil {
		log.Printf("error: constructor for type %s returned nil", t.TypeName)
		return nil
	}
	// Deserialize the object.
	tm.pushContext()
	tm.updateDynamicValues(tm.CurrentContext())
	s.Deserialize(t, true)
	tm.popContext()
	// Recalculate stats
	s.RecalculateStats()
	return s
}

// Create returns a pointer to a newly constructed object that has been assigned
// a unique serial of the correct type and added to the data stores. It does not
// yet have a parent nor has it been placed on the map. One of these things must
// be done otherwise the data store will leak this object.
func Create[K Object](which string) K {
	var zero K
	o := GenerateObject(which)
	if o == nil {
		return zero
	}
	ro, ok := o.(K)
	if !ok {
		return zero
	}
	tm.storeObject(ro)
	return ro
}

// loadListFile loads a single list file and merges that list with the current.
func (m *TemplateManager) loadListFile(filePath string, d []byte) []error {
	var errs []error

	// Load all segments in the file
	lfr := &util.ListFileReader{}
	lfr.StartReading(bytes.NewReader(d))
	for l := lfr.ReadNextSegment(); l != nil; l = lfr.ReadNextSegment() {
		// Avoid panic on duplicate Add, because I'd like to see all the errors
		if m.lists.Contains(l.Name) {
			errs = append(errs, fmt.Errorf("duplicate list name %s", l.Name))
			continue
		}
		m.lists.Add(l.Name, l.Contents)
	}

	return errs
}

// loadTemplateFile loads all the template definitions in the provided file
func (m *TemplateManager) loadTemplateFile(filePath string, d []byte) []error {
	var errs []error

	// Load all template objects in the file
	tfr := &util.TagFileReader{}
	tfr.StartReading(bytes.NewReader(d))
	for {
		tfo := tfr.ReadObject()
		if tfo == nil {
			errs = append(errs, tfr.Errors()...)
			break
		}
		t, terrs := NewTemplate(tfo, m)
		if len(terrs) > 0 {
			errs = append(errs, terrs...)
		} else {
			m.templates.Add(t.TemplateName, t)
		}
	}

	return errs
}

func (m *TemplateManager) resolveTemplate(t *Template) error {
	// Refuse to do duplicate work
	if t.IsResolved {
		return nil
	}

	// This is a primary template, nothing to resolve
	if t.BaseTemplate == "" {
		t.compileExpressions()
		t.IsResolved = true
		return nil
	}

	// Try to resolve our base template first
	base, ok := m.templates.Get(t.BaseTemplate)
	if !ok || base == nil {
		return fmt.Errorf("base template %s not found while resolving template %s", t.BaseTemplate, t.TemplateName)
	}
	if !base.IsResolved {
		if err := m.resolveTemplate(base); err != nil {
			return err
		}
	}

	// Integrate the base template's properties into our own
	for k, e := range base.properties {
		if _, found := t.properties[k]; !found {
			// If we are not overriding the expression from the base we just use
			// the base expression.
			t.properties[k] = e
		} else {
			// We are overriding the base expression in some way, handle that
			ne := t.properties[k]
			if len(ne.text) > 1 && ne.text[0] == '+' {
				// This is a list addition to the base expression
				ne.text = e.text + "," + ne.text[1:]
			} // Else this is a normal override
		}
	}
	t.compileExpressions()
	t.IsResolved = true

	return nil
}

// resolveInheritance resolves the inheritance chains of all templates the
// manager has loaded.
func (m *TemplateManager) resolveInheritance() []error {
	errs := m.templates.Map(func(k string, t *Template) error {
		if err := m.resolveTemplate(t); err != nil {
			return err
		}
		return nil
	})
	return errs
}
