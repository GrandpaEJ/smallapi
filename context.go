package smallapi

import (
        "encoding/json"
        "fmt"
        "html/template"
        "io"
        "net/http"
        "net/url"
        "strconv"
        "strings"
)

// Context provides request/response handling similar to Flask's request and g objects
type Context struct {
        Request    *http.Request
        Response   http.ResponseWriter
        app        *App
        params     map[string]string
        query      url.Values
        form       url.Values
        session    *Session
        data       map[string]interface{} // Similar to Flask's g object
        written    bool
        statusCode int
}

// NewContext creates a new context for a request
func NewContext(w http.ResponseWriter, r *http.Request, app *App, sessionManager *SessionManager) *Context {
        ctx := &Context{
                Request:    r,
                Response:   w,
                app:        app,
                params:     make(map[string]string),
                data:       make(map[string]interface{}),
                statusCode: 200,
        }

        // Parse query parameters
        ctx.query = r.URL.Query()

        // Parse form data if applicable
        if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
                r.ParseForm()
                ctx.form = r.PostForm
        }

        // Initialize session
        ctx.session = sessionManager.GetSession(r, w)

        return ctx
}

// Param returns a URL parameter by name (e.g., /users/:id)
func (c *Context) Param(name string) string {
        return c.params[name]
}

// ParamInt returns a URL parameter as an integer
func (c *Context) ParamInt(name string) (int, error) {
        value := c.params[name]
        if value == "" {
                return 0, fmt.Errorf("parameter %s not found", name)
        }
        return strconv.Atoi(value)
}

// Query returns a query parameter by name
func (c *Context) Query(name string) string {
        return c.query.Get(name)
}

// QueryDefault returns a query parameter with a default value
func (c *Context) QueryDefault(name, defaultValue string) string {
        value := c.query.Get(name)
        if value == "" {
                return defaultValue
        }
        return value
}

// QueryInt returns a query parameter as an integer
func (c *Context) QueryInt(name string) (int, error) {
        value := c.query.Get(name)
        if value == "" {
                return 0, fmt.Errorf("query parameter %s not found", name)
        }
        return strconv.Atoi(value)
}

// Form returns a form field value
func (c *Context) Form(name string) string {
        return c.form.Get(name)
}

// FormDefault returns a form field with a default value
func (c *Context) FormDefault(name, defaultValue string) string {
        value := c.form.Get(name)
        if value == "" {
                return defaultValue
        }
        return value
}

// JSON parses request body as JSON into the provided struct
// If called without arguments, it sends JSON response
func (c *Context) JSON(v interface{}) error {
        if c.Request.Method == "GET" || c.Request.ContentLength == 0 {
                // This is a response
                return c.sendJSON(v)
        }

        // This is parsing request JSON
        decoder := json.NewDecoder(c.Request.Body)
        return decoder.Decode(v)
}

// sendJSON sends a JSON response
func (c *Context) sendJSON(v interface{}) error {
        c.Response.Header().Set("Content-Type", "application/json")
        if c.statusCode != 200 {
                c.Response.WriteHeader(c.statusCode)
        }
        c.written = true
        return json.NewEncoder(c.Response).Encode(v)
}

// String sends a plain text response
func (c *Context) String(text string) {
        c.Response.Header().Set("Content-Type", "text/plain")
        if c.statusCode != 200 {
                c.Response.WriteHeader(c.statusCode)
        }
        c.written = true
        c.Response.Write([]byte(text))
}

// HTML sends an HTML response
func (c *Context) HTML(html string) {
        c.Response.Header().Set("Content-Type", "text/html")
        if c.statusCode != 200 {
                c.Response.WriteHeader(c.statusCode)
        }
        c.written = true
        c.Response.Write([]byte(html))
}

// Render renders a template with data
func (c *Context) Render(templateName string, data interface{}) error {
        c.Response.Header().Set("Content-Type", "text/html")
        if c.statusCode != 200 {
                c.Response.WriteHeader(c.statusCode)
        }
        c.written = true

        // Use the app's template engine
        if c.app.templates != nil {
                output, err := c.app.templates.Render(templateName, data)
                if err != nil {
                        // If template not found, use fallback
                        return fmt.Errorf("template error: %v", err)
                }
                if len(output) == 0 {
                        return fmt.Errorf("template rendered empty content")
                }
                c.Response.Write([]byte(output))
                return nil
        }

        // Fallback template rendering if no template engine is configured
        tmplContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
</head>
<body>
    <h1>{{.title}}</h1>
    <pre>%+v</pre>
</body>
</html>`, data)

        tmpl, err := template.New(templateName).Parse(tmplContent)
        if err != nil {
                return err
        }

        return tmpl.Execute(c.Response, data)
}

// File sends a file response
func (c *Context) File(filePath string) {
        http.ServeFile(c.Response, c.Request, filePath)
        c.written = true
}

// Redirect sends a redirect response
func (c *Context) Redirect(url string) {
        http.Redirect(c.Response, c.Request, url, http.StatusFound)
        c.written = true
}

// Status sets the HTTP status code for the response
func (c *Context) Status(code int) *Context {
        c.statusCode = code
        return c
}

// Header sets a response header
func (c *Context) Header(key, value string) *Context {
        c.Response.Header().Set(key, value)
        return c
}

// Cookie sets a cookie
func (c *Context) Cookie(cookie *http.Cookie) *Context {
        http.SetCookie(c.Response, cookie)
        return c
}

// GetCookie gets a cookie by name
func (c *Context) GetCookie(name string) (*http.Cookie, error) {
        return c.Request.Cookie(name)
}

// Session returns the session for this request
func (c *Context) Session() *Session {
        return c.session
}

// Set stores a value in the context (similar to Flask's g object)
func (c *Context) Set(key string, value interface{}) {
        c.data[key] = value
}

// Get retrieves a value from the context
func (c *Context) Get(key string) interface{} {
        return c.data[key]
}

// GetString retrieves a string value from the context
func (c *Context) GetString(key string) string {
        if value, ok := c.data[key].(string); ok {
                return value
        }
        return ""
}

// GetInt retrieves an int value from the context
func (c *Context) GetInt(key string) int {
        if value, ok := c.data[key].(int); ok {
                return value
        }
        return 0
}

// Body returns the request body as bytes
func (c *Context) Body() ([]byte, error) {
        return io.ReadAll(c.Request.Body)
}

// IsAjax returns true if the request is an AJAX request
func (c *Context) IsAjax() bool {
        return strings.ToLower(c.Request.Header.Get("X-Requested-With")) == "xmlhttprequest"
}

// IP returns the client IP address
func (c *Context) IP() string {
        // Check for forwarded IP first
        if ip := c.Request.Header.Get("X-Forwarded-For"); ip != "" {
                return strings.Split(ip, ",")[0]
        }
        if ip := c.Request.Header.Get("X-Real-IP"); ip != "" {
                return ip
        }
        return c.Request.RemoteAddr
}

// UserAgent returns the User-Agent header
func (c *Context) UserAgent() string {
        return c.Request.Header.Get("User-Agent")
}

// Method returns the HTTP method
func (c *Context) Method() string {
        return c.Request.Method
}

// Path returns the request path
func (c *Context) Path() string {
        return c.Request.URL.Path
}

// URL returns the full request URL
func (c *Context) URL() *url.URL {
        return c.Request.URL
}
