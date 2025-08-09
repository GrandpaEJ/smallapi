package smallapi

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Documentation version
const (
	DocsVersion    = "1.0.0"
	OpenAPIVersion = "3.0.0"
)

// SwaggerDoc represents the OpenAPI documentation structure
type SwaggerDoc struct {
	OpenAPI    string                 `json:"openapi"`
	Info       map[string]interface{} `json:"info"`
	Paths      map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components,omitempty"`
}

// SwaggerConfig holds the documentation configuration
type SwaggerConfig struct {
	Title       string            `json:"title"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	BasePath    string            `json:"basePath"`
	Contact     map[string]string `json:"contact,omitempty"`
	Credits     []string          `json:"x-credits,omitempty"`
}

// EnableDocs adds automatic API documentation to the app
func (app *App) EnableDocs(config *SwaggerConfig) {
	if config == nil {
		config = &SwaggerConfig{
			Title:       "SmallAPI Documentation",
			Version:     "1.0.0",
			Description: "API documentation powered by SmallAPI",
			BasePath:    "/",
			Contact: map[string]string{
				"name": "GrandpaEJ",
				"url":  "https://github.com/grandpaej/smallapi",
			},
			Credits: []string{
				"Powered by SmallAPI Framework",
			},
		}
	}

	// Serve static assets
	app.Static("/assets/", "./assets")

	// Serve documentation UI
	app.Get("/docs", func(c *Context) {
		c.HTML(getDocsTemplate())
	})

	// Serve OpenAPI specification
	app.Get("/docs.json", func(c *Context) {
		c.Header("X-Powered-By", "SmallAPI")
		c.Header("Content-Type", "application/json")

		doc := SwaggerDoc{
			OpenAPI: "3.0.0",
			Info: map[string]interface{}{
				"title":       config.Title,
				"version":     config.Version,
				"description": config.Description,
				"contact":     config.Contact,
				"x-credits":   config.Credits,
			},
			Paths: generatePaths(app.router.routes),
		}

		c.JSON(doc)
	})
}

// generatePaths converts routes to OpenAPI paths
func generatePaths(routes []Route) map[string]interface{} {
	paths := make(map[string]interface{})

	for _, route := range routes {
		// Skip documentation routes
		if route.Path == "/docs" || route.Path == "/docs.json" {
			continue
		}

		method := strings.ToLower(route.Method)

		pathItem := map[string]interface{}{
			"summary":    fmt.Sprintf("%s %s", route.Method, route.Path),
			"parameters": extractPathParameters(route.Path),
		}

		if paths[route.Path] == nil {
			paths[route.Path] = make(map[string]interface{})
		}
		paths[route.Path].(map[string]interface{})[method] = pathItem
	}

	return paths
}

// extractPathParameters gets parameters from route path
func extractPathParameters(path string) []map[string]interface{} {
	var params []map[string]interface{}

	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			paramName := segment[1:]
			params = append(params, map[string]interface{}{
				"name":        paramName,
				"in":          "path",
				"required":    true,
				"description": fmt.Sprintf("Path parameter: %s", paramName),
				"schema": map[string]string{
					"type": "string",
				},
			})
		}
	}

	return params
}

// getDocsTemplate returns the HTML template for documentation UI
func getDocsTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>SmallAPI Documentation</title>
    <link rel="stylesheet" type="text/css" href="/assets/css/swagger-ui.css">
    <script src="/assets/js/swagger-ui-bundle.js"></script>
    <style>
        .topbar { display: none }
        .swagger-ui .info { margin: 20px 0 }
        .powered-by {
            text-align: center;
            padding: 10px;
            color: #666;
            border-top: 1px solid #eee;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <div class="powered-by">
        Powered by <a href="https://github.com/grandpaej/smallapi" target="_blank">SmallAPI</a>
    </div>
    <script>
        window.onload = () => {
            SwaggerUIBundle({
                url: '/docs.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
            });
        };
    </script>
</body>
</html>`
}

// APIDoc represents the API documentation structure
type APIDoc struct {
	Title       string           `json:"title"`
	Version     string           `json:"version"`
	Description string           `json:"description"`
	BasePath    string           `json:"basePath"`
	Paths       map[string]*Path `json:"paths"`
}

// Path represents a single API path documentation
type Path struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

// Operation represents an API operation documentation
type Operation struct {
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	Responses   map[string]Response `json:"responses"`
	Tags        []string            `json:"tags,omitempty"`
}

// Parameter represents an API parameter documentation
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // query, path, body
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
}

// Response represents an API response documentation
type Response struct {
	Description string `json:"description"`
}

// Documentation maintains API documentation state
type Documentation struct {
	doc *APIDoc
}

// NewDocumentation creates a new API documentation instance
func NewDocumentation() *Documentation {
	return &Documentation{
		doc: &APIDoc{
			Title:       "SmallAPI Documentation",
			Version:     "1.0.0",
			Description: "API documentation generated by SmallAPI",
			BasePath:    "/",
			Paths:       make(map[string]*Path),
		},
	}
}

// AddRoute adds a route to the documentation
func (d *Documentation) AddRoute(method, path string, handler HandlerFunc) {
	if d.doc.Paths[path] == nil {
		d.doc.Paths[path] = &Path{}
	}

	operation := &Operation{
		Summary:     fmt.Sprintf("%s %s", method, path),
		Description: "Endpoint automatically documented by SmallAPI",
		Responses: map[string]Response{
			"200": {Description: "Successful operation"},
		},
	}

	// Extract path parameters
	if strings.Contains(path, ":") {
		segments := strings.Split(strings.Trim(path, "/"), "/")
		for _, segment := range segments {
			if strings.HasPrefix(segment, ":") {
				paramName := segment[1:]
				operation.Parameters = append(operation.Parameters, Parameter{
					Name:        paramName,
					In:          "path",
					Description: fmt.Sprintf("Path parameter: %s", paramName),
					Required:    true,
					Type:        "string",
				})
			}
		}
	}

	switch strings.ToUpper(method) {
	case "GET":
		d.doc.Paths[path].Get = operation
	case "POST":
		d.doc.Paths[path].Post = operation
	case "PUT":
		d.doc.Paths[path].Put = operation
	case "DELETE":
		d.doc.Paths[path].Delete = operation
	}
}

// GenerateJSON generates JSON documentation
func (d *Documentation) GenerateJSON() ([]byte, error) {
	return json.MarshalIndent(d.doc, "", "  ")
}

// Add these methods to the Router struct in router.go:
