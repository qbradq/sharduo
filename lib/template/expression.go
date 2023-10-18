package template

import (
	"bytes"
	"strings"
	"text/template"
)

// expression holds a text expression that may be interpreted as a Go template
type expression struct {
	text string             // Raw text of the expression
	pct  *template.Template // If non-nil this is the pre-compiled Go template
}

// compile preps the expression to execute
func (e *expression) compile() error {
	if strings.Contains(e.text, "{{") {
		var err error
		e.pct = template.New(e.text)
		e.pct = e.pct.Funcs(templateFuncMap)
		e.pct, err = e.pct.Parse(e.text)
		return err
	}
	return nil
}

// execute executes the expression returning the string value
func (e *expression) execute() (string, error) {
	if e.pct == nil {
		return e.text, nil
	}
	buf := bytes.NewBuffer(nil)
	if err := e.pct.Execute(buf, tm.CurrentContext()); err != nil {
		return "", err
	}
	return buf.String(), nil
}
