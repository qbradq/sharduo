package template

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/qbradq/sharduo/lib/util"
)

// T contains all of the property lines of the template.
type T struct {
	// Name of the object constructor used to create the object.
	TypeName string
	// Unique name of the template.
	TemplateName string
	// Name of the base template. The empty string means a root template.
	BaseTemplate string
	// True if the template's inheritance chain has already been satisfied.
	IsResolved bool
	// List of all properties
	properties map[string]interface{}
}

// New creates a new template.T object from the provided TagFileObject. The
// inheritance chain has not been resolved for this object, but all text
// templates have been pre-compiles and ready to run.
func New(tfo *util.TagFileObject, tm *TemplateManager) (*T, []error) {
	t := &T{
		TypeName:   tfo.TypeName(),
		properties: make(map[string]interface{}),
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

// // generateTagFileObject generates a new TagFileObject by executing the
// // templates contained within the strings.
// func (t *T) generateTagFileObject(tm *TemplateManager, ctx map[string]string) (*util.TagFileObject, error) {
// 	tfo := util.NewTagFileObject()
// 	for k, p := range t.properties {
// 		switch v := p.(type) {
// 		case string:
// 			tfo.Set(k, v)
// 		case *template.Template:
// 			buf := bytes.NewBuffer(nil)
// 			if err := v.Execute(buf, ctx); err != nil {
// 				return nil, err
// 			}
// 			tfo.Set(k, buf.String())
// 		default:
// 			panic("unhandled type in generateTagFileObject")
// 		}
// 	}
// 	return tfo, nil
// }
