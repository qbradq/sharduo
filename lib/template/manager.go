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

// Global context for constants in templates. This is built from
// data/template-variables.ini
var templateContext = map[string]string{}

// Global template manager instance
var tm *TemplateManager

// TemplateManager manages a collection of templates
type TemplateManager struct {
	// Registry of templates in the manager
	templates *util.Registry[string, *T]
	// Registry of pre-compiled go templates
	pctp *util.Registry[string, *txtmp.Template]
	// Collection of lists used by the random object creation methods
	lists *util.Registry[string, []string]
	// Source of randomness for randomly generating things
	rng uo.RandomSource
	// Adds the object to game datastores
	storeObject func(Object)
}

// Initialize initializes the template package and must be called prior to all
// other package calls. The function parameter fn should add the game object to
// internal data stores when a new object is created by the template engine.
func Initialize(templatePath, listPath, variablesFilePath string, rng uo.RandomSource, fn func(Object)) []error {
	// Initialize the manager
	tm = &TemplateManager{
		templates:   util.NewRegistry[string, *T]("templates"),
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
	// Load all variables
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
		templateContext[name] = value
		return nil
	})

	return nil
}

// FindTemplate returns a pointer to the named template or nil if not found.
func FindTemplate(which string) *T {
	t, found := tm.templates.Get(which)
	if !found {
		log.Printf("template %s not found\n", which)
		return nil
	}
	return t
}

// UpdateContextValues randomizes the values of all context variables with
// random or changing values.
func UpdateContextValues() {
	// Inject dynamic values into the template context
	if tm.rng.RandomBool() {
		templateContext["IsFemale"] = "true"
	} else {
		templateContext["IsFemale"] = ""
	}
}

// GenerateObject returns a pointer to a newly constructed object as in Create()
// but the object is not added to the data stores.
func GenerateObject(which string) Object {
	t := FindTemplate(which)
	if t == nil {
		return nil
	}
	UpdateContextValues()
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
	s.Deserialize(t, true)
	// Recalculate stats
	s.RecalculateStats()
	return s
}

// Create returns a pointer to a newly constructed object that has been assigned
// a unique serial of the correct type and added to the data stores. It does not
// yet have a parent nor has it been placed on the map. One of these things must
// be done otherwise the data store will leak this object.
func Create(which string) Object {
	o := GenerateObject(which)
	tm.storeObject(o)
	return o
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
		t, terrs := New(tfo, m)
		if len(terrs) > 0 {
			errs = append(errs, terrs...)
		} else {
			m.templates.Add(t.TemplateName, t)
		}
	}

	return errs
}

func (m *TemplateManager) resolveTemplate(t *T) error {
	// Refuse to do duplicate work
	if t.IsResolved {
		return nil
	}

	// This is a primary template, nothing to resolve
	if t.BaseTemplate == "" {
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
	for k, v := range base.properties {
		// If we are not overriding the value from the base, use the base value
		if _, found := t.properties[k]; !found {
			t.properties[k] = v
		}
	}
	t.IsResolved = true

	return nil
}

// resolveInheritance resolves the inheritance chains of all templates the
// manager has loaded.
func (m *TemplateManager) resolveInheritance() []error {
	errs := m.templates.Map(func(k string, t *T) error {
		if err := m.resolveTemplate(t); err != nil {
			return err
		}
		return nil
	})
	return errs
}

// // GetTemplateObject returns the util.TagFileObject generated from the named
// // template, or nil if the template was not found.
// func (m *TemplateManager) GetTemplateObject(templateName string) (*T, *util.TagFileObject) {
// 	// Find the template
// 	t, found := m.templates.Get(templateName)
// 	if !found {
// 		log.Printf("template %s not found\n", templateName)
// 		return nil, nil
// 	}
// 	// Inject dynamic values into the template context
// 	if m.rng.RandomBool() {
// 		templateContext["IsFemale"] = "true"
// 	} else {
// 		templateContext["IsFemale"] = ""
// 	}
// 	// Generate the deserialization object
// 	tfo, err := t.generateTagFileObject(m, templateContext)
// 	if err != nil {
// 		log.Printf("error while processing template %s", t.TemplateName)
// 		return t, nil
// 	}
// 	return t, tfo
// }

// // newObject creates a new object with the given template name, or nil if the
// // template was not found, there was an error executing the template, or there
// // was an error deserializing the object.
// func (m *TemplateManager) newObject(templateName string) game.Object {
// 	// Fetch the template
// 	t, tfo := m.GetTemplateObject(templateName)
// 	if t == nil || tfo == nil {
// 		return nil
// 	}
// 	// Create the object
// 	ctor := game.Constructor(t.TypeName)
// 	if ctor == nil {
// 		log.Printf("error: constructor not found for type %s", t.TypeName)
// 		return nil
// 	}
// 	s := ctor()
// 	if s == nil {
// 		log.Printf("error: constructor for type %s returned nil", t.TypeName)
// 		return nil
// 	}
// 	// Deserialize the object.
// 	s.Deserialize(tfo)
// 	errs := tfo.Errors()
// 	for _, err := range errs {
// 		log.Println(err)
// 	}
// 	if len(errs) > 0 {
// 		// Do not return a bad object
// 		return nil
// 	}
// 	// Recalculate stats
// 	s.RecalculateStats()
// 	return s
// }
