# SmallAPI 🚀

A Go web framework with Python/Flask-like simplicity - production-ready, zero dependencies, with hot reload and intuitive API design.

[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
Made With AI Agent and GrandpaEJ

## Why SmallAPI?

SmallAPI brings the simplicity and developer experience of Python's Flask to Go development. If you love Flask's minimalism but need Go's performance, SmallAPI is for you.

```go
// Python Flask
@app.route('/')
def hello():
    return {'hello': 'world'}

// SmallAPI (Go)
app.Get("/", func(c *smallapi.Context) {
    c.JSON(map[string]string{"hello": "world"})
})
```

## ✨ Features

### 🔥 **Flask-like Simplicity**
- Intuitive routing: `app.Get("/users/:id", handler)`
- Context-based request handling
- Familiar patterns for Python developers

### ⚡ **Zero Dependencies**
- Built entirely on Go's standard library
- No external dependencies to manage
- Maximum compatibility and security

### 🛠️ **Developer Experience**
- Hot reload in development mode
- Comprehensive error messages
- Auto-loading templates and static files

### 🔒 **Production Ready**
- Built-in security headers and middleware
- Rate limiting and request throttling
- Graceful shutdown and panic recovery
- Session management and authentication

### 📊 **Rich Feature Set**
- RESTful routing with parameter capture
- JSON/Form/Multipart data parsing
- Template rendering with data binding
- WebSocket support for real-time apps
- Request validation with struct tags
- Middleware system for composable functionality

## ⚠️ Warning 
This PKG still in beta mode . many feature not working smoothly <br>
<b>EX: `hot-reloading`</b> 

## 🚀 Quick Start

### Installation

```bash
go mod init your-project
go get github.com/grandpaej/smallapi
```

### Hello World

```go
package main

import "github.com/grandpaej/smallapi"

func main() {
    app := smallapi.New()
    
    app.Get("/", func(c *smallapi.Context) {
        c.JSON(map[string]string{
            "message":   "Hello, SmallAPI!",
            "framework": "SmallAPI",
            "language":  "Go",
        })
    })
    
    app.Run(":8080")
}
```

### Development with Hot Reload

```go
// Use RunDev() for automatic restart on file changes
app.RunDev(":8080")
```

Visit `http://localhost:8080` and see your API in action!

## 🔧 Core Concepts

### Routing & Parameters

```go
app.Get("/users/:id", func(c *smallapi.Context) {
    userID := c.Param("id")
    c.JSON(map[string]string{"user_id": userID})
})

app.Get("/search", func(c *smallapi.Context) {
    query := c.Query("q")
    page := c.QueryDefault("page", "1")
    // Handle search logic
})
```

### Middleware

```go
// Built-in middleware
app.Use(smallapi.Logger())       // Request logging
app.Use(smallapi.CORS())         // Cross-origin requests
app.Use(smallapi.Recovery())     // Panic recovery
app.Use(smallapi.RateLimit(60))  // Rate limiting
app.Use(smallapi.Secure())       // Security headers

// Custom middleware
app.Use(func(c *smallapi.Context) bool {
    start := time.Now()
    defer func() {
        log.Printf("Request took %v", time.Since(start))
    }()
    return true // Continue to next middleware
})
```

### Request Handling

```go
// JSON requests
app.Post("/users", func(c *smallapi.Context) {
    var user User
    if err := c.JSON(&user); err != nil {
        c.Status(400).JSON(map[string]string{"error": "Invalid JSON"})
        return
    }
    
    // Validate data
    if err := c.Validate(&user); err != nil {
        c.Status(400).JSON(map[string]string{"error": err.Error()})
        return
    }
    
    // Save user...
    c.Status(201).JSON(user)
})

// Form handling
app.Post("/contact", func(c *smallapi.Context) {
    name := c.Form("name")
    email := c.Form("email")
    message := c.FormDefault("message", "No message")
    
    // Process form...
    c.JSON(map[string]string{"status": "submitted"})
})
```

### Templates

```go
app.Templates("./views")

app.Get("/", func(c *smallapi.Context) {
    data := map[string]interface{}{
        "title": "Welcome",
        "user":  currentUser,
        "items": []string{"Item 1", "Item 2"},
    }
    c.Render("index.html", data)
})
```

### Sessions & Authentication

```go
// Enable sessions
app.Use(smallapi.SessionMiddleware())

app.Post("/login", func(c *smallapi.Context) {
    // Validate credentials...
    c.Session().Set("user_id", userID)
    c.JSON(map[string]string{"status": "logged in"})
})

// Protected routes
app.Use(smallapi.RequireAuth())
app.Get("/dashboard", dashboardHandler)
```

### Static Files

```go
app.Static("/static", "./public")   // Serve ./public/* at /static/*
app.Static("/css", "./assets/css")  // Multiple static directories
```

### Route Groups

```go
// API v1
v1 := app.Group("/api/v1")
v1.Use(apiAuthMiddleware())
v1.Get("/users", getUsersHandler)
v1.Post("/users", createUserHandler)

// Admin routes
admin := app.Group("/admin")
admin.Use(requireAdminAuth())
admin.Get("/dashboard", adminDashboard)
```

### WebSockets

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

## 📚 Examples

### Complete REST API

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

func main() {
    app := smallapi.New()
    
    // Middleware
    app.Use(smallapi.Logger())
    app.Use(smallapi.CORS())
    app.Use(smallapi.Recovery())
    
    // Routes
    app.Get("/users", getUsers)
    app.Post("/users", createUser)
    app.Get("/users/:id", getUser)
    app.Put("/users/:id", updateUser)
    app.Delete("/users/:id", deleteUser)
    
    app.Run(":8080")
}

func getUsers(c *smallapi.Context) {
    // Get users from database...
    users := []User{
        {ID: 1, Name: "John", Email: "john@example.com"},
        {ID: 2, Name: "Jane", Email: "jane@example.com"},
    }
    c.JSON(users)
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
    
    // Save to database...
    user.ID = generateID()
    c.Status(201).JSON(user)
}
```

### Web Application with Templates

```go
func main() {
    app := smallapi.New()
    
    app.Templates("./views")
    app.Static("/static", "./public")
    
    app.Get("/", func(c *smallapi.Context) {
        data := map[string]interface{}{
            "title": "SmallAPI Demo",
            "users": getUsers(),
        }
        c.Render("index.html", data)
    })
    
    app.Get("/user/:id", func(c *smallapi.Context) {
        id := c.Param("id")
        user := getUserByID(id)
        
        if user == nil {
            c.Status(404).Render("404.html", nil)
            return
        }
        
        c.Render("user.html", map[string]interface{}{
            "user": user,
        })
    })
    
    app.Run(":8080")
}
```

## 🎯 Why Choose SmallAPI?

### Coming from Flask?

SmallAPI provides the same development experience you love in Flask, but with Go's performance and type safety:

| Flask (Python) | SmallAPI (Go) |
|---|---|
| `@app.route('/')` | `app.Get("/", ...)` |
| `request.json` | `c.JSON(&data)` |
| `request.form['name']` | `c.Form("name")` |
| `session['user_id']` | `c.Session().Set("user_id", ...)` |
| `render_template()` | `c.Render()` |

### Coming from other Go frameworks?

SmallAPI prioritizes simplicity without sacrificing power:

| Feature | SmallAPI | Others |
|---|---|---|
| **Learning Curve** | Minimal - Flask-like syntax | Steep - framework-specific patterns |
| **Dependencies** | Zero | Many external packages |
| **Performance** | Native Go speed | Varies |
| **Development** | Hot reload built-in | Additional setup required |
| **Documentation** | Complete & practical | Often incomplete |

## 📖 Documentation

- **[Getting Started](docs/getting-started.md)** - Complete tutorial
- **[API Reference](docs/api-reference.md)** - Detailed function documentation
- **[Examples](docs/examples.md)** - Real-world use cases
- **[Migration from Flask](docs/migration-from-flask.md)** - For Python developers

## 📁 Project Structure

```
your-app/
├── main.go              # Application entry point
├── handlers/            # Request handlers
│   ├── users.go
│   └── auth.go
├── middleware/          # Custom middleware
│   └── auth.go
├── models/              # Data models
│   └── user.go
├── views/               # HTML templates
│   ├── layout.html
│   ├── index.html
│   └── user.html
└── public/              # Static files
    ├── css/
    ├── js/
    └── images/
```

## 🚀 Production Deployment

SmallAPI apps deploy anywhere Go runs:

```go
func main() {
    app := smallapi.New()
    
    // Production middleware
    app.Use(smallapi.Logger())
    app.Use(smallapi.Recovery())
    app.Use(smallapi.Secure())
    app.Use(smallapi.RateLimit(100))
    
    // Your routes...
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    app.Run(":" + port)
}
```

Deploy to:
- **Docker** - Single binary deployment
- **Kubernetes** - Cloud-native scaling
- **Traditional servers** - Direct binary execution
- **Serverless** - Functions with minimal cold start

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

SmallAPI is MIT licensed. See [LICENSE](LICENSE) for details.

## ⭐ Star History

If you find SmallAPI useful, please consider giving it a star! ⭐

---

**Built with ❤️ for developers who value simplicity and performance.**
