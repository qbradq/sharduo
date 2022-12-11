package uod

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"strings"
	"text/template"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// Global function map for templates
var templateFuncMap = template.FuncMap{
	// RandomBool returns a random boolean value
	"RandomBool": randomBool,
	// RandomSkinHue returns a random human skin hue
	"RandomSkinHue": randomSkinHue,
}

// Global context for constants in templates
var templateContext = map[string]string{
	"BodyHuman": "400",
}

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
			return nil
		}
		if name == "BaseTemplate" {
			t.baseTemplateName = value
			return nil
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
			tm.syncBuffer.Reset()
			if err := v.Execute(tm.syncBuffer, ctx); err != nil {
				return nil, err
			}
			tfo.Set(k, tm.syncBuffer.String())
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
	// Text buffer for template execution, to reduce memory allocations
	syncBuffer *bytes.Buffer
}

// NewTemplateManager returns a new TemplateManager object
func NewTemplateManager(name string) *TemplateManager {
	return &TemplateManager{
		templates:  util.NewRegistry[string, *ObjectTemplate](name),
		pctp:       util.NewRegistry[string, *template.Template]("templates"),
		syncBuffer: bytes.NewBuffer(nil),
	}
}

// LoadAll recursively loads all template tiles under the given directory. Note
// that this uses the embedded file system from the data package. On successful
// completion, all templates will have inheritance resolved, be parsed and be
// ready to execute.
func (m *TemplateManager) LoadAll(dirPath string) []error {
	var errs []error
	files, err := data.FS.ReadDir(dirPath)
	if err != nil {
		return []error{err}
	}
	for _, file := range files {
		if file.IsDir() {
			errs = append(errs, m.LoadAll(path.Join(dirPath, file.Name()))...)
		} else {
			errs = append(errs, m.loadFile(path.Join(dirPath, file.Name()))...)
		}
	}
	// If we have errors at this point then at least one template failed to
	// to load. This might break inheritance chains, so go ahead and bail.
	if len(errs) > 0 {
		return errs
	}
	return m.resolveInheritance()
}

// loadFile loads all the template definitions in the provided file
func (m *TemplateManager) loadFile(filePath string) []error {
	var errs []error
	d, err := data.FS.ReadFile(filePath)
	if err != nil {
		return []error{err}
	}
	tfr := util.NewTagFileReader(bytes.NewReader(d))

	// Load all template objects in the file
	for {
		tfo, err := tfr.ReadObject()
		if errors.Is(err, io.EOF) {
			errs = append(errs, tfr.Errors()...)
			break
		}
		if tfo == nil {
			continue
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

// NewObject creates a new object with the given template name, or nil if the
// template was not found, there was an error executing the template, or there
// was an error deserializing the object.
func (m *TemplateManager) NewObject(templateName string) game.Object {
	t, found := m.templates.Get(templateName)
	if !found {
		log.Printf("template %s not found\n", templateName)
		return nil
	}
	tfo, err := t.generateTagFileObject(m, templateContext)
	if err != nil {
		log.Printf("error while processing template %s", t.templateName)
		return nil
	}
	s := game.ObjectFactory.New(t.typeName, nil)
	if s == nil {
		log.Printf("error while creating object for type %s", t.typeName)
		return nil
	}
	// If we've gotten here we at least have an uninitialized object of the
	// proper type. We can return it in case of error.
	s.Deserialize(tfo)
	for _, err := range tfo.Errors() {
		log.Println(err)
	}
	return s.(game.Object)
}

// ResolveTemplates

func randomBool() bool {
	return world.Random().RandomBool()
}

func randomSkinHue() uo.Hue {
	return uo.RandomSkinHue(world.Random())
}
