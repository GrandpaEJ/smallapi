package smallapi

import (
        "fmt"
        "html/template"
        "io/fs"
        "path/filepath"
        "strings"
)

// TemplateEngine handles template rendering
type TemplateEngine struct {
        templates map[string]*template.Template
        dir       string
        funcs     template.FuncMap
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() *TemplateEngine {
        return &TemplateEngine{
                templates: make(map[string]*template.Template),
                funcs:     make(template.FuncMap),
        }
}

// LoadDir loads all templates from a directory
func (te *TemplateEngine) LoadDir(dir string) error {
        te.dir = dir
        
        // Walk through the directory and load all .html files
        err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
                if err != nil {
                        return err
                }
                
                if d.IsDir() {
                        return nil
                }
                
                if strings.HasSuffix(path, ".html") {
                        name := strings.TrimPrefix(path, dir+"/")
                        // Handle relative path correctly
                        if name == path {
                                // If TrimPrefix didn't change the path, try with just the filename
                                name = filepath.Base(path)
                        }
                        
                        tmpl, err := template.New(name).Funcs(te.funcs).ParseFiles(path)
                        if err != nil {
                                return err
                        }
                        te.templates[name] = tmpl
                }
                
                return nil
        })
        
        return err
}

// AddFunc adds a template function
func (te *TemplateEngine) AddFunc(name string, fn interface{}) {
        te.funcs[name] = fn
}

// Render renders a template with data
func (te *TemplateEngine) Render(name string, data interface{}) (string, error) {
        tmpl, exists := te.templates[name]
        if !exists {
                return "", fmt.Errorf("template %s not found", name)
        }
        
        var buf strings.Builder
        err := tmpl.Execute(&buf, data)
        return buf.String(), err
}

// GetTemplate returns a template by name
func (te *TemplateEngine) GetTemplate(name string) *template.Template {
        return te.templates[name]
}

// DefaultFunctions returns default template functions
func DefaultFunctions() template.FuncMap {
        return template.FuncMap{
                "upper": strings.ToUpper,
                "lower": strings.ToLower,
                "title": strings.Title,
                "join":  strings.Join,
                "split": strings.Split,
                "add": func(a, b int) int {
                        return a + b
                },
                "sub": func(a, b int) int {
                        return a - b
                },
                "mul": func(a, b int) int {
                        return a * b
                },
                "div": func(a, b int) int {
                        if b == 0 {
                                return 0
                        }
                        return a / b
                },
        }
}
