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

// Template represents a collection of data used to initialize new objects
type Template struct {
	// Registered type name of the object to create
	typeName string
	// Name of this template, MUST BE UNIQUE
	templateName string
	// Base of the base template this one inherits from
	baseTemplateName string
	// Collection of properties, which might be a string or a *template.Template
	properties map[string]interface{}
}

// NewTemplate creates a new Template object from the provided TagFileObject.
// The inheritence chain has not been resolved for this object, but all text
// templates have been pre-compiles and ready to run.
func NewTemplate(tfo *util.TagFileObject, tm *TemplateManager) (*Template, []error) {
	t := &Template{
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
func (t *Template) generateTagFileObject(tm *TemplateManager, ctx map[string]string) (*util.TagFileObject, error) {
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
	templates *util.Registry[string, *Template]
	// Registry of pre-compiled go templates
	pctp *util.Registry[string, *template.Template]
	// Text buffer for template execution, to reduce memory allocations
	syncBuffer *bytes.Buffer
}

// NewTemplateManager returns a new TemplateManager object
func NewTemplateManager(name string) *TemplateManager {
	return &TemplateManager{
		templates:  util.NewRegistry[string, *Template](name),
		pctp:       util.NewRegistry[string, *template.Template]("templates"),
		syncBuffer: bytes.NewBuffer(nil),
	}
}

// LoadAll recursively loads all template tiles under the given directory. Note
// that this uses the embedded file system from the data package.
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
	return errs
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
			return append(errs, tfr.Errors()...)
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

func randomBool() bool {
	return world.Random().RandomBool()
}

func randomSkinHue() uo.Hue {
	return uo.RandomSkinHue(world.Random())
}
