package uod

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// Global function map for templates
var templateFuncMap = template.FuncMap{
	"New":        templateNew,      // New creates a new object from the named template, adds it to the world datastores, then returns the string representation of the object's serial
	"PartialHue": partialHue,       // Sets the partial hue flag
	"RandomNew":  randomNew,        // RandomNew creates a new object of a template randomly selected from the named list"RandomNew"
	"RandomBool": randomBool,       // RandomBool returns a random boolean value
	"Random":     randomListMember, // Random returns a random string from the named list, or an empty string if the named list was not found
}

// Global context for constants in templates. This is built from
// data/template-variables.ini
var templateContext = map[string]string{}

// ObjectTemplate represents a collection of data used to initialize new objects
type ObjectTemplate struct {
	// Registered type name of the object to create
	typeName string
	// Name of this template, MUST BE UNIQUE
	templateName string
	// Base of the base template this one inherits from
	baseTemplateName string
	// True if the template's inheritance chain has already been satisfied
	isResolved bool
	// Collection of properties, which might be a string or a *template.Template
	properties map[string]interface{}
}

// NewTemplate creates a new Template object from the provided TagFileObject.
// The inheritence chain has not been resolved for this object, but all text
// templates have been pre-compiles and ready to run.
func NewTemplate(tfo *util.TagFileObject, tm *TemplateManager) (*ObjectTemplate, []error) {
	t := &ObjectTemplate{
		typeName:   tfo.TypeName(),
		properties: make(map[string]interface{}),
	}
	templateName := tfo.GetString("TemplateName", "")
	if templateName == "" {
		panic(fmt.Sprintf("template of type %s missing TemplateName field", tfo.TypeName()))
	}
	errs := tfo.Map(func(name, value string) error {
		if name == "TemplateName" {
			t.templateName = value
		} else if name == "BaseTemplate" {
			t.baseTemplateName = value
		}
		if !strings.Contains(value, "{{") {
			t.properties[name] = value
			return nil
		}
		if tm.pctp.Contains(value) {
			return nil
		}
		tt := template.New(value)
		tt = tt.Funcs(templateFuncMap)
		tt, err := tt.Parse(value)
		if err != nil {
			return err
		}
		tm.pctp.Add(value, tt)
		t.properties[name] = tt
		return nil
	})
	return t, errs
}

// generateTagFileObject generates a new TagFileObject by executing the
// templates contained within the strings.
func (t *ObjectTemplate) generateTagFileObject(tm *TemplateManager, ctx map[string]string) (*util.TagFileObject, error) {
	tfo := util.NewTagFileObject()
	for k, p := range t.properties {
		switch v := p.(type) {
		case string:
			tfo.Set(k, v)
		case *template.Template:
			buf := bytes.NewBuffer(nil)
			if err := v.Execute(buf, ctx); err != nil {
				return nil, err
			}
			tfo.Set(k, buf.String())
		default:
			panic("unhandled type in generateTagFileObject")
		}
	}
	return tfo, nil
}

// TemplateManager manages a collection of templates
type TemplateManager struct {
	// Registry of templates in the manager
	templates *util.Registry[string, *ObjectTemplate]
	// Registry of pre-compiled go templates
	pctp *util.Registry[string, *template.Template]
	// Collection of lists used by the random object creation methods
	lists *util.Registry[string, []string]
}

// NewTemplateManager returns a new TemplateManager object
func NewTemplateManager(name string) *TemplateManager {
	return &TemplateManager{
		templates: util.NewRegistry[string, *ObjectTemplate](name),
		pctp:      util.NewRegistry[string, *template.Template]("templates"),
		lists:     util.NewRegistry[string, []string]("lists"),
	}
}

// LoadAll recursively loads all template tiles under the given directory. Note
// that this uses the embedded file system from the data package. On successful
// completion, all templates will have inheritance resolved, be parsed and be
// ready to execute.
func (m *TemplateManager) LoadAll(templatePath, listPath string) []error {
	// Load all templates
	errs := data.Walk(templatePath, m.loadTemplateFile)
	if len(errs) > 0 {
		return errs
	}
	// Resolve template inheritance chains
	errs = m.resolveInheritance()
	if len(errs) > 0 {
		return errs
	}
	// Load all lists
	errs = data.Walk(listPath, m.loadListFile)
	if len(errs) > 0 {
		return errs
	}
	// Load all variables
	vd, err := data.FS.Open(configuration.TemplateVariablesFile)
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
			m.templates.Add(t.templateName, t)
		}
	}

	return errs
}

func (m *TemplateManager) resolveTemplate(t *ObjectTemplate) error {
	// Refuse to do duplicate work
	if t.isResolved {
		return nil
	}

	// This is a primary template, nothing to resolve
	if t.baseTemplateName == "" {
		t.isResolved = true
		return nil
	}

	// Try to resolve our base template first
	base, ok := m.templates.Get(t.baseTemplateName)
	if !ok || base == nil {
		return fmt.Errorf("base template %s not found while resolving template %s", t.baseTemplateName, t.templateName)
	}
	if err := m.resolveTemplate(base); err != nil {
		return err
	}

	// Integrate the base template's properties into our own
	for k, v := range base.properties {
		// If we are not overriding the value from the base, use the base value
		if _, found := t.properties[k]; !found {
			t.properties[k] = v
		}
	}
	t.isResolved = true

	return nil
}

// resolveInheritance resolves the inheritance chains of all templates the
// manager has loaded.
func (m *TemplateManager) resolveInheritance() []error {
	errs := m.templates.Map(func(k string, t *ObjectTemplate) error {
		if err := m.resolveTemplate(t); err != nil {
			return err
		}
		return nil
	})
	return errs
}

// newObject creates a new object with the given template name, or nil if the
// template was not found, there was an error executing the template, or there
// was an error deserializing the object.
func (m *TemplateManager) newObject(templateName string) game.Object {
	// Find the template
	t, found := m.templates.Get(templateName)
	if !found {
		log.Printf("template %s not found\n", templateName)
		return nil
	}
	// Inject dynamic values into the template context
	if world.Random().RandomBool() {
		templateContext["IsFemale"] = "true"
	} else {
		templateContext["IsFemale"] = ""
	}
	// Generate the deserialization object
	tfo, err := t.generateTagFileObject(m, templateContext)
	if err != nil {
		log.Printf("error while processing template %s", t.templateName)
		return nil
	}
	// Create the object
	ctor := game.Constructor(t.typeName)
	if ctor == nil {
		log.Printf("error: constructor not found for type %s", t.typeName)
		return nil
	}
	s := ctor()
	if s == nil {
		log.Printf("error: constructor for type %s returned nil", t.typeName)
		return nil
	}
	// Deserialize the object.
	s.Deserialize(tfo)
	errs := tfo.Errors()
	for _, err := range errs {
		log.Println(err)
	}
	if len(errs) > 0 {
		// Do not return a bad object
		return nil
	}
	// Recalculate stats
	s.RecalculateStats()
	return s
}

func randomBool() bool {
	return world.Random().RandomBool()
}

func templateNew(name string) string {
	o := world.New(name)
	if o == nil {
		return "0"
	}
	return o.Serial().String()
}

func randomListMember(name string) string {
	l, ok := templateManager.lists.Get(name)
	if !ok || len(l) == 0 {
		log.Printf("list %s not found\n", name)
		return ""
	}
	return l[world.Random().Random(0, len(l)-1)]
}

func randomNew(name string) string {
	tn := randomListMember(name)
	if tn == "" {
		return "0"
	}
	return templateNew(tn)
}

func partialHue(hue string) string {
	v, err := strconv.ParseInt(hue, 0, 32)
	if err != nil {
		return hue
	}
	h := uo.Hue(v).SetPartialHue()
	return fmt.Sprintf("%d", h)
}
