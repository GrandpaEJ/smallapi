# SmallAPI API Reference

Complete reference for all SmallAPI functions, methods, and types.

## Table of Contents

- [Application](#application)
- [Context](#context)
- [Routing](#routing)
- [Middleware](#middleware)
- [Templates](#templates)
- [Sessions](#sessions)
- [Authentication](#authentication)
- [Validation](#validation)
- [WebSockets](#websockets)

## Application

### `smallapi.New() *App`

Creates a new SmallAPI application instance.

```go
app := smallapi.New()
```

**Returns:** A new App instance ready for configuration.

### `App.Use(middleware MiddlewareFunc) *App`

Adds middleware to the application. Middleware runs in the order they are added.

```go
app.Use(smallapi.Logger())
app.Use(smallapi.CORS())
```

**Parameters:**
- `middleware`: A function that implements the `MiddlewareFunc` signature

**Returns:** The App instance for method chaining.

### HTTP Method Handlers

#### `App.Get(path string, handler HandlerFunc) *App`
#### `App.Post(path string, handler HandlerFunc) *App`
#### `App.Put(path string, handler HandlerFunc) *App`
#### `App.Delete(path string, handler HandlerFunc) *App`
#### `App.Patch(path string, handler HandlerFunc) *App`
#### `App.Options(path string, handler HandlerFunc) *App`

Register handlers for specific HTTP methods.

```go
app.Get("/users", getUsersHandler)
app.Post("/users", createUserHandler)
app.Put("/users/:id", updateUserHandler)
app.Delete("/users/:id", deleteUserHandler)
```

**Parameters:**
- `path`: URL path pattern (supports parameters like `:id`)
- `handler`: Function to handle the request

**Returns:** The App instance for method chaining.

### `App.Route(methods []string, path string, handler HandlerFunc) *App`

Register a handler for multiple HTTP methods.

```go
app.Route([]string{"GET", "POST"}, "/users", userHandler)
```

### `App.Group(prefix string) *RouteGroup`

Create a route group with a common prefix.

```go
api := app.Group("/api/v1")
api.Get("/users", getUsersHandler)  // Routes to /api/v1/users
```

### `App.Static(urlPath, dirPath string) *App`

Serve static files from a directory.

```go
app.Static("/static", "./public")    // Serve ./public/* at /static/*
app.Static("/", "./assets")          // Serve ./assets/* at /*
```

### `App.Templates(dir string) *App`

Set the template directory for HTML rendering.

```go
app.Templates("./views")
```

### `App.Run(addr string) error`

Start the HTTP server.

```go
app.Run(":8080")          // Listen on port 8080
app.Run("0.0.0.0:3000")   // Listen on all interfaces, port 3000
```

### `App.RunDev(addr string) error`

Start the server in development mode with hot reload.

```go
app.RunDev(":8080")
```

## Context

The Context object provides access to request and response functionality.

### Request Information

#### `Context.Param(name string) string`

Get a URL parameter value.

```go
// Route: /users/:id
id := c.Param("id")
```

#### `Context.ParamInt(name string) (int, error)`

Get a URL parameter as an integer.

```go
id, err := c.ParamInt("id")
if err != nil {
    c.Status(400).JSON(map[string]string{"error": "Invalid ID"})
    return
}
```

#### `Context.Query(name string) string`

Get a query parameter value.

```go
// URL: /search?q=golang
query := c.Query("q")  // Returns "golang"
```

#### `Context.QueryDefault(name, defaultValue string) string`

Get a query parameter with a default value.

```go
page := c.QueryDefault("page", "1")
```

#### `Context.QueryInt(name string) (int, error)`

Get a query parameter as an integer.

```go
limit, err := c.QueryInt("limit")
```

#### `Context.Form(name string) string`

Get a form field value.

```go
email := c.Form("email")
```

#### `Context.FormDefault(name, defaultValue string) string`

Get a form field with a default value.

```go
message := c.FormDefault("message", "No message")
```

#### `Context.JSON(v interface{}) error`

Parse request body as JSON or send JSON response.

```go
// Parse request JSON
var user User
if err := c.JSON(&user); err != nil {
    c.Status(400).JSON(map[string]string{"error": "Invalid JSON"})
    return
}

// Send JSON response
c.JSON(map[string]string{"status": "success"})
```

#### `Context.Body() ([]byte, error)`

Get the raw request body.

```go
body, err := c.Body()
```

### Response Methods

#### `Context.String(text string)`

Send a plain text response.

```go
c.String("Hello, World!")
```

#### `Context.HTML(html string)`

Send an HTML response.

```go
c.HTML("<h1>Welcome</h1>")
```

#### `Context.Render(templateName string, data interface{}) error`

Render an HTML template with data.

```go
c.Render("index.html", map[string]interface{}{
    "title": "Home",
    "user":  currentUser,
})
```

#### `Context.File(filePath string)`

Send a file as response.

```go
c.File("./uploads/document.pdf")
```

#### `Context.Redirect(url string)`

Send a redirect response.

```go
c.Redirect("/login")
```

#### `Context.Status(code int) *Context`

Set the HTTP status code for the response.

```go
c.Status(201).JSON(newUser)
c.Status(404).JSON(map[string]string{"error": "Not found"})
```

#### `Context.Header(key, value string) *Context`

Set a response header.

```go
c.Header("Content-Type", "application/json")
c.Header("X-Custom-Header", "value")
```

#### `Context.Cookie(cookie *http.Cookie) *Context`

Set a cookie.

```go
cookie := &http.Cookie{
    Name:     "session",
    Value:    "abc123",
    Path:     "/",
    HttpOnly: true,
}
c.Cookie(cookie)
```

#### `Context.GetCookie(name string) (*http.Cookie, error)`

Get a cookie from the request.

```go
cookie, err := c.GetCookie("session")
```

### Context Data

#### `Context.Set(key string, value interface{})`
#### `Context.Get(key string) interface{}`
#### `Context.GetString(key string) string`
#### `Context.GetInt(key string) int`

Store and retrieve data within the request context.

```go
c.Set("user_id", 123)
userID := c.GetInt("user_id")
```

### Session Management

#### `Context.Session() *Session`

Get the session for the current request.

```go
session := c.Session()
session.Set("user_id", userID)
userID := session.Get("user_id")
```

### Request Validation

#### `Context.Validate(v interface{}) error`

Validate a struct using validation tags.

```go
type User struct {
    Name  string `validate:"required,min=2"`
    Email string `validate:"required,email"`
}

var user User
c.JSON(&user)
if err := c.Validate(&user); err != nil {
    c.Status(400).JSON(map[string]string{"error": err.Error()})
    return
}
```

### Utility Methods

#### `Context.Method() string`

Get the HTTP method.

```go
method := c.Method()  // "GET", "POST", etc.
```

#### `Context.Path() string`

Get the request path.

```go
path := c.Path()  // "/users/123"
```

#### `Context.IP() string`

Get the client IP address.

```go
ip := c.IP()
```

#### `Context.UserAgent() string`

Get the User-Agent header.

```go
ua := c.UserAgent()
```

#### `Context.IsAjax() bool`

Check if the request is an AJAX request.

```go
if c.IsAjax() {
    c.JSON(data)
} else {
    c.Render("page.html", data)
}
```

## Middleware

SmallAPI provides built-in middleware for common functionality.

### `Logger() MiddlewareFunc`

Logs HTTP requests with method, path, status, and duration.

```go
app.Use(smallapi.Logger())
```

### `CORS() MiddlewareFunc`

Handles Cross-Origin Resource Sharing with default settings.

```go
app.Use(smallapi.CORS())
```

### `CORSWithConfig(config CORSConfig) MiddlewareFunc`

CORS middleware with custom configuration.

```go
app.Use(smallapi.CORSWithConfig(smallapi.CORSConfig{
    AllowOrigins: []string{"https://myapp.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Content-Type", "Authorization"},
}))
```

### `Recovery() MiddlewareFunc`

Recovers from panics and returns a 500 error.

```go
app.Use(smallapi.Recovery())
```

### `RateLimit(requestsPerMinute int) MiddlewareFunc`

Limits requests per IP address.

```go
app.Use(smallapi.RateLimit(60))  // 60 requests per minute
```

### `Secure() MiddlewareFunc`

Adds security headers to responses.

```go
app.Use(smallapi.Secure())
```

### `BasicAuth(username, password string) MiddlewareFunc`

HTTP Basic Authentication.

```go
app.Use(smallapi.BasicAuth("admin", "secret"))
```

### `RequireAuth() MiddlewareFunc`

Requires user authentication via session.

```go
app.Use(smallapi.RequireAuth())
```

### `SessionMiddleware() MiddlewareFunc`

Provides session support (sessions are enabled by default).

```go
app.Use(smallapi.SessionMiddleware())
```

## Validation

SmallAPI supports struct validation using tags.

### Validation Tags

- `required`: Field must have a value
- `min=N`: Minimum length/value
- `max=N`: Maximum length/value
- `email`: Valid email format
- `url`: Valid URL format
- `numeric`: Contains only numbers
- `alpha`: Contains only letters
- `alphanum`: Contains only letters and numbers
- `regex=pattern`: Matches custom regex pattern

### Example

```go
type User struct {
    Username string `json:"username" validate:"required,min=3,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"min=18,max=120"`
    Website  string `json:"website" validate:"url"`
    Password string `json:"password" validate:"required,min=8"`
}

func createUser(c *smallapi.Context) {
    var user User
    if err := c.JSON(&user); err != nil {
        c.Status(400).JSON(map[string]string{"error": "Invalid JSON"})
        return
    }
    
    if err := c.Validate(&user); err != nil {
        c.Status(400).JSON(map[string]string{"error": err.Error()})
        return
    }
    
    // User data is valid
}
```

## Sessions

Sessions provide server-side storage for user data across requests.

### Session Methods

#### `Session.Set(key string, value interface{})`

Store a value in the session.

```go
c.Session().Set("user_id", 123)
c.Session().Set("username", "john")
```

#### `Session.Get(key string) interface{}`

Retrieve a value from the session.

```go
userID := c.Session().Get("user_id")
```

#### `Session.GetString(key string) string`
#### `Session.GetInt(key string) int`

Type-safe getters.

```go
username := c.Session().GetString("username")
userID := c.Session().GetInt("user_id")
```

#### `Session.Delete(key string)`

Remove a key from the session.

```go
c.Session().Delete("temp_data")
```

#### `Session.Clear()`

Remove all data from the session.

```go
c.Session().Clear()
```

#### `Session.Has(key string) bool`

Check if a key exists in the session.

```go
if c.Session().Has("user_id") {
    // User is logged in
}
```

## Authentication

SmallAPI provides built-in authentication components.

### AuthManager

#### `NewAuthManager() *AuthManager`

Create a new authentication manager.

```go
authManager := smallapi.NewAuthManager()
```

#### `AuthManager.Register(username, email, password string) (*User, error)`

Register a new user.

```go
user, err := authManager.Register("john", "john@example.com", "password123")
```

#### `AuthManager.Login(username, password string) (string, *User, error)`

Authenticate a user and return a session token.

```go
token, user, err := authManager.Login("john", "password123")
if err == nil {
    c.Session().Set("auth_token", token)
}
```

#### `AuthManager.GetUser(token string) *User`

Get user by session token.

```go
token := c.Session().Get("auth_token")
user := authManager.GetUser(token.(string))
```

### Authentication Middleware

#### `Auth(authManager *AuthManager) MiddlewareFunc`

Adds user information to context if authenticated.

```go
app.Use(smallapi.Auth(authManager))

// In handlers:
user := c.Get("user")  // *User or nil
```

#### `RequireUser(authManager *AuthManager) MiddlewareFunc`

Requires authentication, returns 401 if not authenticated.

```go
protected := app.Group("/api")
protected.Use(smallapi.RequireUser(authManager))
```

## WebSockets

SmallAPI supports WebSocket upgrades for real-time communication.

### `Context.Upgrade(handler WebSocketHandler) error`

Upgrade an HTTP connection to WebSocket.

```go
app.Get("/ws", func(c *smallapi.Context) {
    c.Upgrade(func(ws *smallapi.WebSocket) {
        for {
            message, err := ws.ReadMessage()
            if err != nil {
                break
            }
            ws.WriteMessage([]byte("Echo: " + string(message)))
        }
    })
})
```

### WebSocket Methods

#### `WebSocket.ReadMessage() ([]byte, error)`

Read a message from the WebSocket connection.

```go
message, err := ws.ReadMessage()
```

#### `WebSocket.WriteMessage(data []byte) error`

Write a message to the WebSocket connection.

```go
err := ws.WriteMessage([]byte("Hello, WebSocket!"))
```

#### `WebSocket.WriteText(text string) error`

Write a text message.

```go
err := ws.WriteText("Hello, World!")
```

#### `WebSocket.WriteJSON(v interface{}) error`
#### `WebSocket.ReadJSON(v interface{}) error`

Read and write JSON messages.

```go
// Send JSON
err := ws.WriteJSON(map[string]string{"type": "notification"})

// Read JSON
var message map[string]interface{}
err := ws.ReadJSON(&message)
```

#### `WebSocket.Close() error`

Close the WebSocket connection.

```go
defer ws.Close()
```

## Route Groups

Route groups allow organizing routes with common prefixes and middleware.

### `RouteGroup.Get/Post/Put/Delete(path, handler)`

Add routes to the group.

```go
api := app.Group("/api/v1")
api.Use(someMiddleware())

api.Get("/users", getUsersHandler)      // Routes to /api/v1/users
api.Post("/users", createUserHandler)   // Routes to /api/v1/users
```

### `RouteGroup.Use(middleware)`

Add middleware to all routes in the group.

```go
adminRoutes := app.Group("/admin")
adminRoutes.Use(requireAdminAuth())

adminRoutes.Get("/dashboard", adminDashboard)
adminRoutes.Get("/users", adminUsers)
```

## Error Handling

SmallAPI provides automatic error recovery and easy error handling patterns.

### Automatic Recovery

The `Recovery()` middleware automatically catches panics:

```go
app.Use(smallapi.Recovery())
```

### Manual Error Handling

```go
app.Get("/users/:id", func(c *smallapi.Context) {
    id := c.Param("id")
    
    user, err := getUserFromDB(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.Status(404).JSON(map[string]string{"error": "User not found"})
            return
        }
        
        log.Printf("Database error: %v", err)
        c.Status(500).JSON(map[string]string{"error": "Internal server error"})
        return
    }
    
    c.JSON(user)
})
```

## Types

### HandlerFunc

Function signature for route handlers.

```go
type HandlerFunc func(*Context)
```

### MiddlewareFunc

Function signature for middleware.

```go
type MiddlewareFunc func(*Context) bool
```

Return `true` to continue to the next middleware/handler, `false` to stop processing.

### WebSocketHandler

Function signature for WebSocket handlers.

```go
type WebSocketHandler func(*WebSocket)
```

This completes the SmallAPI API reference. For more examples and tutorials, see the [Getting Started](getting-started.md) guide and [Examples](examples.md).
