package smallapi

import (
	"strings"
)

// HandlerFunc defines the handler function signature
type HandlerFunc func(*Context)

// Route represents a single route
type Route struct {
	Method  string
	Path    string
	Handler HandlerFunc
	Pattern *RoutePattern
}

// RoutePattern represents a compiled route pattern
type RoutePattern struct {
	Segments []Segment
}

// Segment represents a part of a route pattern
type Segment struct {
	IsParam bool
	Name    string
	Value   string
}

// Router handles routing for the application
type Router struct {
	routes []Route
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes: make([]Route, 0),
	}
}

// Add adds a new route to the router
func (r *Router) Add(method, path string, handler HandlerFunc) {
	pattern := r.compilePattern(path)
	route := Route{
		Method:  method,
		Path:    path,
		Handler: handler,
		Pattern: pattern,
	}
	r.routes = append(r.routes, route)
}

// compilePattern compiles a route pattern like "/users/:id/posts/:postId"
func (r *Router) compilePattern(path string) *RoutePattern {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	pattern := &RoutePattern{
		Segments: make([]Segment, len(segments)),
	}

	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			// Parameter segment
			pattern.Segments[i] = Segment{
				IsParam: true,
				Name:    segment[1:], // Remove the ":"
			}
		} else {
			// Static segment
			pattern.Segments[i] = Segment{
				IsParam: false,
				Value:   segment,
			}
		}
	}

	return pattern
}

// Match finds a matching route for the given method and path
func (r *Router) Match(method, path string) (HandlerFunc, map[string]string) {
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	
	for _, route := range r.routes {
		if route.Method != method {
			continue
		}

		params, matches := r.matchPattern(route.Pattern, pathSegments)
		if matches {
			return route.Handler, params
		}
	}

	return nil, nil
}

// matchPattern checks if a path matches a route pattern and extracts parameters
func (r *Router) matchPattern(pattern *RoutePattern, pathSegments []string) (map[string]string, bool) {
	if len(pattern.Segments) != len(pathSegments) {
		return nil, false
	}

	params := make(map[string]string)

	for i, segment := range pattern.Segments {
		if segment.IsParam {
			// Parameter segment - capture the value
			params[segment.Name] = pathSegments[i]
		} else {
			// Static segment - must match exactly
			if segment.Value != pathSegments[i] {
				return nil, false
			}
		}
	}

	return params, true
}

// Routes returns all registered routes (useful for debugging)
func (r *Router) Routes() []Route {
	return r.routes
}

// PrintRoutes prints all registered routes (useful for debugging)
func (r *Router) PrintRoutes() {
	for _, route := range r.routes {
		println(route.Method, route.Path)
	}
}
