package tpl

import (
	"io"
	"time"

	"text/template/parse"

	"html/template"
	texttemplate "text/template"

	bp "github.com/geego/gean/app/bufferpool"
	"github.com/geego/gean/app/metrics"
)

var (
	_ TemplateExecutor = (*TemplateAdapter)(nil)
)

// TemplateHandler manages the collection of templates.
type TemplateHandler interface {
	TemplateFinder
	AddTemplate(name, tpl string) error
	AddLateTemplate(name, tpl string) error
	LoadTemplates(absPath, prefix string)
	PrintErrors()

	MarkReady()
	RebuildClone()
}

// TemplateFinder finds templates.
type TemplateFinder interface {
	Lookup(name string) *TemplateAdapter
}

// Template is the common interface between text/template and html/template.
type Template interface {
	Execute(wr io.Writer, data interface{}) error
	Name() string
}

// TemplateExecutor adds some extras to Template.
type TemplateExecutor interface {
	Template
	ExecuteToString(data interface{}) (string, error)
	Tree() string
}

// TemplateDebugger prints some debug info to stdoud.
type TemplateDebugger interface {
	Debug()
}

// TemplateAdapter implements the TemplateExecutor interface.
type TemplateAdapter struct {
	Template
	Metrics metrics.Provider
}

// Execute executes the current template. The actual execution is performed
// by the embedded text or html template, but we add an implementation here so
// we can add a timer for some metrics.
func (t *TemplateAdapter) Execute(w io.Writer, data interface{}) error {
	if t.Metrics != nil {
		defer t.Metrics.MeasureSince(t.Name(), time.Now())
	}
	return t.Template.Execute(w, data)
}

// ExecuteToString executes the current template and returns the result as a
// string.
func (t *TemplateAdapter) ExecuteToString(data interface{}) (string, error) {
	b := bp.GetBuffer()
	defer bp.PutBuffer(b)
	if err := t.Execute(b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Tree returns the template Parse tree as a string.
// Note: this isn't safe for parallel execution on the same template
// vs Lookup and Execute.
func (t *TemplateAdapter) Tree() string {
	var tree *parse.Tree
	switch tt := t.Template.(type) {
	case *template.Template:
		tree = tt.Tree
	case *texttemplate.Template:
		tree = tt.Tree
	default:
		panic("Unknown template")
	}

	if tree == nil || tree.Root == nil {
		return ""
	}
	s := tree.Root.String()

	return s
}

// TemplateFuncsGetter allows to get a map of functions.
type TemplateFuncsGetter interface {
	GetFuncs() map[string]interface{}
}

// TemplateTestMocker adds a way to override some template funcs during tests.
// The interface is named so it's not used in regular application code.
type TemplateTestMocker interface {
	SetFuncs(funcMap map[string]interface{})
}
