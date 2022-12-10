package uod

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"path"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

// Template represents a collection of data used to initialize new objects
type Template struct {
	// Name of this template, MUST BE UNIQUE
	templateName string
	// Base of the base template this one inherits from
	baseTemplateName string
	// The TagFileObject that describes this template, with template-specific
	// fields stripped
	tfo *util.TagFileObject
}

// NewTemplate creates a new Template object from the provided TagFileObject.
// The inheritence chain has not been resolved for this object.
func NewTemplate(tfo *util.TagFileObject) *Template {
	t := &Template{
		templateName:     tfo.GetString("TemplateName", "__ERROR__"),
		baseTemplateName: tfo.GetString("BaseTemplate", ""),
		tfo:              tfo,
	}
	if t.templateName == "__ERROR__" {
		panic(fmt.Sprintf("template of type %s missing TemplateName field", tfo.TypeName()))
	}
	tfo.Delete("TemplateName")
	tfo.Delete("BaseTemplate")
	return t
}

// generateTagFileObject generates a new TagFileObject by executing the
// preprocessor over all of the strings of the template.
func (t *Template) generateTagFileObject() *util.TagFileObject {
	return t.tfo
}

// TemplateManager manages a collection of templates
type TemplateManager struct {
	templates *util.Registry[string, *Template]
}

// NewTemplateManager returns a new TemplateManager object
func NewTemplateManager(name string) *TemplateManager {
	return &TemplateManager{
		templates: util.NewRegistry[string, *Template](name),
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
	log.Println(filePath)
	d, err := data.FS.ReadFile(filePath)
	if err != nil {
		return []error{err}
	}
	tfr := util.NewTagFileReader(bytes.NewReader(d))
	for {
		tfo, err := tfr.ReadObject()
		if errors.Is(err, io.EOF) {
			return tfr.Errors()
		}
		if tfo == nil {
			continue
		}
		t := NewTemplate(tfo)
		m.templates.Add(t.templateName, t)
	}
}

// NewObject creates a new object with the given template name, or nil if the
// template was not found.
func (m *TemplateManager) NewObject(templateName string) game.Object {
	t, found := m.templates.Get(templateName)
	if !found {
		return nil
	}
	s := game.ObjectFactory.New(t.tfo.TypeName(), t.generateTagFileObject())
	s.Deserialize(t.tfo)
	return s.(game.Object)
}
