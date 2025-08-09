// Authentication example demonstrating user registration, login, and protected routes
package main

import (
        "time"
        "github.com/grandpaej/smallapi"
)

// LoginRequest represents a login request
type LoginRequest struct {
        Username string `json:"username" validate:"required,min=3"`
        Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
        Username string `json:"username" validate:"required,min=3,alphanum"`
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=6"`
        Name     string `json:"name" validate:"required,min=2"`
}

// UserProfile represents a user's public profile
type UserProfile struct {
        ID       string    `json:"id"`
        Username string    `json:"username"`
        Email    string    `json:"email"`
        Name     string    `json:"name"`
        Role     string    `json:"role"`
        Created  time.Time `json:"created"`
}

func main() {
        app := smallapi.New()
        
        // Create authentication manager
        authManager := smallapi.NewAuthManager()
        
        // Create some demo users
        createDemoUsers(authManager)
        
        // Middleware
        app.Use(smallapi.Logger())
        app.Use(smallapi.CORS())
        app.Use(smallapi.Recovery())
        app.Use(smallapi.SessionMiddleware())
        app.Use(smallapi.Auth(authManager)) // Add user info to context if logged in
        
        // Public routes
        app.Post("/register", func(c *smallapi.Context) {
                var req RegisterRequest
                
                if err := c.JSON(&req); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": "Invalid JSON format",
                        })
                        return
                }
                
                if err := c.Validate(&req); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": err.Error(),
                        })
                        return
                }
                
                user, err := authManager.Register(req.Username, req.Email, req.Password)
                if err != nil {
                        c.Status(409).JSON(map[string]string{
                                "error": err.Error(),
                        })
                        return
                }
                
                // Auto-login after registration
                token, _, err := authManager.Login(req.Username, req.Password)
                if err != nil {
                        c.Status(500).JSON(map[string]string{
                                "error": "Registration successful but login failed",
                        })
                        return
                }
                
                c.Session().Set("auth_token", token)
                
                profile := UserProfile{
                        ID:       user.ID,
                        Username: user.Username,
                        Email:    user.Email,
                        Name:     req.Name,
                        Role:     "user",
                        Created:  user.Created,
                }
                
                c.Status(201).JSON(map[string]interface{}{
                        "message": "Registration successful",
                        "user":    profile,
                })
        })
        
        app.Post("/login", func(c *smallapi.Context) {
                var req LoginRequest
                
                if err := c.JSON(&req); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": "Invalid JSON format",
                        })
                        return
                }
                
                if err := c.Validate(&req); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": err.Error(),
                        })
                        return
                }
                
                token, user, err := authManager.Login(req.Username, req.Password)
                if err != nil {
                        c.Status(401).JSON(map[string]string{
                                "error": "Invalid username or password",
                        })
                        return
                }
                
                c.Session().Set("auth_token", token)
                
                profile := UserProfile{
                        ID:       user.ID,
                        Username: user.Username,
                        Email:    user.Email,
                        Role:     "user",
                        Created:  user.Created,
                }
                
                c.JSON(map[string]interface{}{
                        "message": "Login successful",
                        "user":    profile,
                })
        })
        
        app.Post("/logout", func(c *smallapi.Context) {
                token := c.Session().Get("auth_token")
                if token != nil {
                        authManager.Logout(token.(string))
                        c.Session().Delete("auth_token")
                }
                
                c.JSON(map[string]string{
                        "message": "Logout successful",
                })
        })
        
        // Check authentication status
        app.Get("/auth/status", func(c *smallapi.Context) {
                user := c.Get("user")
                if user == nil {
                        c.JSON(map[string]interface{}{
                                "authenticated": false,
                        })
                        return
                }
                
                u := user.(*smallapi.User)
                profile := UserProfile{
                        ID:       u.ID,
                        Username: u.Username,
                        Email:    u.Email,
                        Role:     "user",
                        Created:  u.Created,
                }
                
                c.JSON(map[string]interface{}{
                        "authenticated": true,
                        "user":          profile,
                })
        })
        
        // Protected routes - require authentication
        protected := app.Group("/api")
        protected.Use(smallapi.RequireUser(authManager))
        
        // Get user profile
        protected.Get("/profile", func(c *smallapi.Context) {
                user := c.Get("user").(*smallapi.User)
                
                profile := UserProfile{
                        ID:       user.ID,
                        Username: user.Username,
                        Email:    user.Email,
                        Role:     "user",
                        Created:  user.Created,
                }
                
                c.JSON(profile)
        })
        
        // Update user profile
        protected.Put("/profile", func(c *smallapi.Context) {
                user := c.Get("user").(*smallapi.User)
                
                var updates map[string]interface{}
                if err := c.JSON(&updates); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": "Invalid JSON format",
                        })
                        return
                }
                
                if err := authManager.UpdateUser(user.ID, updates); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": err.Error(),
                        })
                        return
                }
                
                c.JSON(map[string]string{
                        "message": "Profile updated successfully",
                })
        })
        
        // Change password
        protected.Post("/change-password", func(c *smallapi.Context) {
                user := c.Get("user").(*smallapi.User)
                
                var req struct {
                        OldPassword string `json:"old_password" validate:"required"`
                        NewPassword string `json:"new_password" validate:"required,min=6"`
                }
                
                if err := c.JSON(&req); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": "Invalid JSON format",
                        })
                        return
                }
                
                if err := c.Validate(&req); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": err.Error(),
                        })
                        return
                }
                
                if err := authManager.ChangePassword(user.ID, req.OldPassword, req.NewPassword); err != nil {
                        c.Status(400).JSON(map[string]string{
                                "error": err.Error(),
                        })
                        return
                }
                
                c.JSON(map[string]string{
                        "message": "Password changed successfully",
                })
        })
        
        // Get all users (admin-like functionality)
        protected.Get("/users", func(c *smallapi.Context) {
                users := authManager.ListUsers()
                
                var profiles []UserProfile
                for _, user := range users {
                        profile := UserProfile{
                                ID:       user.ID,
                                Username: user.Username,
                                Email:    user.Email,
                                Role:     "user",
                                Created:  user.Created,
                        }
                        profiles = append(profiles, profile)
                }
                
                c.JSON(map[string]interface{}{
                        "users": profiles,
                        "total": len(profiles),
                })
        })
        
        // Protected resource example
        protected.Get("/dashboard", func(c *smallapi.Context) {
                user := c.Get("user").(*smallapi.User)
                
                dashboardData := map[string]interface{}{
                        "welcome_message": "Welcome to your dashboard, " + user.Username + "!",
                        "user_stats": map[string]interface{}{
                                "login_count":    42, // In real app, track this
                                "last_login":     time.Now().Add(-2 * time.Hour),
                                "account_status": "active",
                        },
                        "recent_activity": []map[string]interface{}{
                                {
                                        "action":    "Profile updated",
                                        "timestamp": time.Now().Add(-1 * time.Hour),
                                },
                                {
                                        "action":    "Password changed",
                                        "timestamp": time.Now().Add(-24 * time.Hour),
                                },
                                {
                                        "action":    "Account created",
                                        "timestamp": user.Created,
                                },
                        },
                }
                
                c.JSON(dashboardData)
        })
        
        // Health check
        app.Get("/health", func(c *smallapi.Context) {
                c.JSON(map[string]interface{}{
                        "status":    "healthy",
                        "timestamp": time.Now(),
                })
        })
        
        app.Run(":8080")
}

// createDemoUsers creates some demo users for testing
func createDemoUsers(authManager *smallapi.AuthManager) {
        // Create admin user
        authManager.Register("admin", "admin@example.com", "admin123")
        
        // Create regular user
        authManager.Register("john", "john@example.com", "password123")
        
        // Create another user
        authManager.Register("jane", "jane@example.com", "secret456")
}
