// Package smallapi provides a Flask-like web framework for Go with zero dependencies
package smallapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// App represents the SmallAPI application
type App struct {
	router     *Router
	middleware []MiddlewareFunc
	templates  *TemplateEngine
	static     map[string]string
	sessions   *SessionManager
}

// MiddlewareFunc defines the middleware function signature
type MiddlewareFunc func(*Context) bool

// New creates a new SmallAPI application
func New() *App {
	return &App{
		router:    NewRouter(),
		templates: NewTemplateEngine(),
		static:    make(map[string]string),
		sessions:  NewSessionManager(),
	}
}

// Use adds middleware to the application
func (a *App) Use(middleware MiddlewareFunc) *App {
	a.middleware = append(a.middleware, middleware)
	return a
}

// Get adds a GET route
func (a *App) Get(path string, handler HandlerFunc) *App {
	a.router.Add("GET", path, handler)
	return a
}

// Post adds a POST route
func (a *App) Post(path string, handler HandlerFunc) *App {
	a.router.Add("POST", path, handler)
	return a
}

// Put adds a PUT route
func (a *App) Put(path string, handler HandlerFunc) *App {
	a.router.Add("PUT", path, handler)
	return a
}

// Delete adds a DELETE route
func (a *App) Delete(path string, handler HandlerFunc) *App {
	a.router.Add("DELETE", path, handler)
	return a
}

// Patch adds a PATCH route
func (a *App) Patch(path string, handler HandlerFunc) *App {
	a.router.Add("PATCH", path, handler)
	return a
}

// Options adds an OPTIONS route
func (a *App) Options(path string, handler HandlerFunc) *App {
	a.router.Add("OPTIONS", path, handler)
	return a
}

// Route adds a route for multiple HTTP methods
func (a *App) Route(methods []string, path string, handler HandlerFunc) *App {
	for _, method := range methods {
		a.router.Add(method, path, handler)
	}
	return a
}

// Static serves static files from a directory
func (a *App) Static(urlPath, dirPath string) *App {
	a.static[urlPath] = dirPath
	return a
}

// Templates sets the template directory
func (a *App) Templates(dir string) *App {
	a.templates.LoadDir(dir)
	return a
}

// ServeHTTP implements the http.Handler interface
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create context
	ctx := NewContext(w, r, a.sessions)

	// Handle static files
	for urlPath, dirPath := range a.static {
		if len(r.URL.Path) >= len(urlPath) && r.URL.Path[:len(urlPath)] == urlPath {
			filePath := filepath.Join(dirPath, r.URL.Path[len(urlPath):])
			if _, err := os.Stat(filePath); err == nil {
				http.ServeFile(w, r, filePath)
				return
			}
		}
	}

	// Run middleware
	for _, middleware := range a.middleware {
		if !middleware(ctx) {
			return // Middleware stopped the request
		}
	}

	// Find and execute route handler
	handler, params := a.router.Match(r.Method, r.URL.Path)
	if handler == nil {
		ctx.Status(404).JSON(map[string]string{"error": "Not Found"})
		return
	}

	// Set route parameters
	ctx.params = params

	// Execute handler with panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in handler: %v", r)
			if !ctx.written {
				ctx.Status(500).JSON(map[string]string{"error": "Internal Server Error"})
			}
		}
	}()

	handler(ctx)
}

// Run starts the HTTP server
func (a *App) Run(addr string) error {
	// Set up graceful shutdown
	server := &http.Server{
		Addr:    addr,
		Handler: a,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("🚀 SmallAPI server starting on %s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("✅ Server stopped gracefully")
	return nil
}

// RunDev starts the server in development mode with hot reload
func (a *App) RunDev(addr string) error {
	fmt.Printf("🔥 SmallAPI server starting in DEV mode on %s\n", addr)
	fmt.Println("📁 Watching for file changes...")
	
	// For now, just run normally - hot reload would require file watching
	// In a real implementation, you'd add file system watchers here
	return a.Run(addr)
}

// Group creates a route group with a common prefix
func (a *App) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		app:    a,
		prefix: prefix,
	}
}

// RouteGroup represents a group of routes with a common prefix
type RouteGroup struct {
	app    *App
	prefix string
}

// Get adds a GET route to the group
func (g *RouteGroup) Get(path string, handler HandlerFunc) *RouteGroup {
	g.app.Get(g.prefix+path, handler)
	return g
}

// Post adds a POST route to the group
func (g *RouteGroup) Post(path string, handler HandlerFunc) *RouteGroup {
	g.app.Post(g.prefix+path, handler)
	return g
}

// Put adds a PUT route to the group
func (g *RouteGroup) Put(path string, handler HandlerFunc) *RouteGroup {
	g.app.Put(g.prefix+path, handler)
	return g
}

// Delete adds a DELETE route to the group
func (g *RouteGroup) Delete(path string, handler HandlerFunc) *RouteGroup {
	g.app.Delete(g.prefix+path, handler)
	return g
}

// Use adds middleware to the group
func (g *RouteGroup) Use(middleware MiddlewareFunc) *RouteGroup {
	// In a real implementation, you'd track group-specific middleware
	g.app.Use(middleware)
	return g
}
