# Getting Started with SmallAPI

Welcome to SmallAPI! This guide will help you get up and running with SmallAPI in just a few minutes.

## Prerequisites

- Go 1.19 or later
- Basic understanding of Go programming
- Familiarity with web development concepts

## Installation

First, make sure you have Go 1.19 or later installed. Then create a new project:

```bash
mkdir my-smallapi-app
cd my-smallapi-app
go mod init my-smallapi-app
go get github.com/grandpaej/smallapi
```

## Your First SmallAPI Application

Create a file called `main.go`:

```go
package main

import "github.com/grandpaej/smallapi"

func main() {
    app := smallapi.New()
    
    app.Get("/", func(c *smallapi.Context) {
        c.JSON(map[string]string{
            "message": "Hello, SmallAPI!",
            "version": "1.0.0",
        })
    })
    
    app.Run(":8080")
}
```

Run your application:

```bash
go run main.go
```

Visit `http://localhost:8080` and you should see:
```json
{"message":"Hello, SmallAPI!","version":"1.0.0"}
```

## Core Concepts

### 1. Application Instance

Every SmallAPI application starts with creating an app instance:

```go
app := smallapi.New()
```

### 2. Routing

SmallAPI supports all standard HTTP methods:

```go
app.Get("/users", getUsersHandler)
app.Post("/users", createUserHandler) 
app.Put("/users/:id", updateUserHandler)
app.Delete("/users/:id", deleteUserHandler)
app.Patch("/users/:id", patchUserHandler)
app.Options("/users", optionsHandler)
```

### 3. Route Parameters

Capture dynamic segments in URLs:

```go
app.Get("/users/:id", func(c *smallapi.Context) {
    userID := c.Param("id")
    c.JSON(map[string]string{"user_id": userID})
})

app.Get("/users/:id/posts/:postId", func(c *smallapi.Context) {
    userID := c.Param("id")
    postID := c.Param("postId")
    // Handle nested parameters
})
```

### 4. Query Parameters

Access URL query parameters:

```go
app.Get("/search", func(c *smallapi.Context) {
    query := c.Query("q")
    page := c.QueryDefault("page", "1")
    limit, err := c.QueryInt("limit")
    // Handle query parameters
})
```

### 5. Request Body

Parse JSON and form data:

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

app.Post("/users", func(c *smallapi.Context) {
    var user User
    if err := c.JSON(&user); err != nil {
        c.Status(400).JSON(map[string]string{"error": "Invalid JSON"})
        return
    }
    
    // Handle the user data
    c.Status(201).JSON(user)
})
```

### 6. Form Data

Handle HTML form submissions:

```go
app.Post("/contact", func(c *smallapi.Context) {
    name := c.Form("name")
    email := c.Form("email")
    message := c.FormDefault("message", "No message provided")
    
    // Process form data
    c.JSON(map[string]string{"status": "submitted"})
})
```

## Middleware

Middleware functions run before your route handlers:

```go
app.Use(smallapi.Logger())        // Request logging
app.Use(smallapi.CORS())          // Cross-origin requests
app.Use(smallapi.Recovery())      // Panic recovery
app.Use(smallapi.RateLimit(60))   // Rate limiting (60 req/min)
app.Use(smallapi.Secure())        // Security headers
```

### Custom Middleware

Create your own middleware:

```go
func customMiddleware() smallapi.MiddlewareFunc {
    return func(c *smallapi.Context) bool {
        // Do something before the request
        start := time.Now()
        
        // Continue to next middleware/handler
        // Return true to continue, false to stop
        
        // Do something after the request
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
        
        return true
    }
}

app.Use(customMiddleware())
```

## Route Groups

Organize related routes with common prefixes and middleware:

```go
// API v1 routes
v1 := app.Group("/api/v1")
v1.Use(someMiddleware())

v1.Get("/users", getAllUsers)
v1.Post("/users", createUser)
v1.Get("/users/:id", getUser)

// API v2 routes
v2 := app.Group("/api/v2")
v2.Use(anotherMiddleware())

v2.Get("/users", getAllUsersV2)
v2.Post("/users", createUserV2)
```

## Response Types

SmallAPI supports multiple response types:

### JSON Response
```go
c.JSON(map[string]interface{}{
    "users": users,
    "total": len(users),
})
```

### Text Response
```go
c.String("Hello, World!")
```

### HTML Response
```go
c.HTML("<h1>Welcome</h1>")
```

### File Response
```go
c.File("./uploads/document.pdf")
```

### Redirect
```go
c.Redirect("/login")
```

### Custom Status Codes
```go
c.Status(201).JSON(newUser)
c.Status(404).JSON(map[string]string{"error": "Not found"})
```

## Data Validation

Use struct tags for automatic validation:

```go
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"min=18,max=120"`
}

app.Post("/users", func(c *smallapi.Context) {
    var req CreateUserRequest
    
    if err := c.JSON(&req); err != nil {
        c.Status(400).JSON(map[string]string{"error": "Invalid JSON"})
        return
    }
    
    if err := c.Validate(&req); err != nil {
        c.Status(400).JSON(map[string]string{"error": err.Error()})
        return
    }
    
    // Process valid data
})
```

## Static Files

Serve static assets like CSS, JavaScript, and images:

```go
// Serve files from "./public" at "/static" URL path
app.Static("/static", "./public")

// Multiple static directories
app.Static("/css", "./assets/css")
app.Static("/js", "./assets/js")
app.Static("/images", "./assets/images")
```

## Templates

Render HTML templates with data:

```go
// Configure template directory
app.Templates("./views")

app.Get("/", func(c *smallapi.Context) {
    data := map[string]interface{}{
        "title": "Welcome",
        "user":  currentUser,
        "items": []string{"Item 1", "Item 2", "Item 3"},
    }
    
    c.Render("index.html", data)
})
```

## Sessions

Handle user sessions:

```go
app.Use(smallapi.SessionMiddleware())

app.Post("/login", func(c *smallapi.Context) {
    // Validate credentials...
    
    // Store data in session
    c.Session().Set("user_id", userID)
    c.Session().Set("username", username)
    
    c.JSON(map[string]string{"status": "logged in"})
})

app.Get("/profile", func(c *smallapi.Context) {
    userID := c.Session().Get("user_id")
    if userID == nil {
        c.Status(401).JSON(map[string]string{"error": "Not logged in"})
        return
    }
    
    // Get user profile...
})
```

## Error Handling

SmallAPI provides automatic panic recovery, but you can also handle errors explicitly:

```go
app.Get("/users/:id", func(c *smallapi.Context) {
    id := c.Param("id")
    
    user, err := getUserFromDatabase(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.Status(404).JSON(map[string]string{
                "error": "User not found",
            })
            return
        }
        
        // Log the error and return generic message
        log.Printf("Database error: %v", err)
        c.Status(500).JSON(map[string]string{
            "error": "Internal server error",
        })
        return
    }
    
    c.JSON(user)
})
```

## Development Mode

Use development mode for automatic restart on file changes:

```go
// Instead of app.Run(":8080")
app.RunDev(":8080")
```

## Production Deployment

For production, use the regular Run method with proper configuration:

```go
func main() {
    app := smallapi.New()
    
    // Production middleware
    app.Use(smallapi.Logger())
    app.Use(smallapi.Recovery())
    app.Use(smallapi.Secure())
    app.Use(smallapi.RateLimit(100)) // 100 requests per minute
    
    // Your routes...
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    app.Run(":" + port)
}
```

## Best Practices

1. **Structure your application**: Organize routes in separate files/packages
2. **Use middleware**: Apply common functionality like logging and authentication
3. **Validate input**: Always validate and sanitize user input
4. **Handle errors**: Provide meaningful error messages
5. **Use route groups**: Organize related routes together
6. **Environment variables**: Use environment variables for configuration
7. **Graceful shutdown**: SmallAPI handles this automatically

## Next Steps

- Explore the [API Reference](api-reference.md) for detailed documentation
- Check out [Examples](examples.md) for real-world use cases
- Learn about [Migration from Flask](migration-from-flask.md) if you're coming from Python
