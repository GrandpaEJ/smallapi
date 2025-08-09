# SmallAPI Examples

This document contains practical examples demonstrating various SmallAPI features.

## Table of Contents

- [REST API with CRUD Operations](#rest-api-with-crud-operations)
- [File Upload and Download](#file-upload-and-download)
- [Real-time Chat with WebSockets](#real-time-chat-with-websockets)
- [Authentication System](#authentication-system)
- [Template-based Web Application](#template-based-web-application)
- [API with Rate Limiting](#api-with-rate-limiting)
- [Microservice with Health Checks](#microservice-with-health-checks)
- [Blog Application](#blog-application)

## REST API with CRUD Operations

Complete REST API for managing users with validation and error handling.

```go
package main

import (
    "strconv"
    "time"
    "github.com/grandpaej/smallapi"
)

type User struct {
    ID       string    `json:"id"`
    Name     string    `json:"name" validate:"required,min=2,max=50"`
    Email    string    `json:"email" validate:"required,email"`
    Age      int       `json:"age" validate:"min=1,max=150"`
    Created  time.Time `json:"created"`
    Modified time.Time `json:"modified"`
}

type UserStore struct {
    users   map[string]*User
    nextID  int
}

func NewUserStore() *UserStore {
    return &UserStore{
        users:  make(map[string]*User),
        nextID: 1,
    }
}

func (s *UserStore) Create(user *User) *User {
    user.ID = strconv.Itoa(s.nextID)
    s.nextID++
    user.Created = time.Now()
    user.Modified = time.Now()
    s.users[user.ID] = user
    return user
}

func (s *UserStore) GetAll() []*User {
    users := make([]*User, 0, len(s.users))
    for _, user := range s.users {
        users = append(users, user)
    }
    return users
}

func (s *UserStore) GetByID(id string) *User {
    return s.users[id]
}

func (s *UserStore) Update(id string, updates *User) *User {
    user := s.users[id]
    if user == nil {
        return nil
    }
    
    if updates.Name != "" {
        user.Name = updates.Name
    }
    if updates.Email != "" {
        user.Email = updates.Email
    }
    if updates.Age > 0 {
        user.Age = updates.Age
    }
    user.Modified = time.Now()
    
    return user
}

func (s *UserStore) Delete(id string) bool {
    if _, exists := s.users[id]; exists {
        delete(s.users, id)
        return true
    }
    return false
}

func main() {
    app := smallapi.New()
    store := NewUserStore()
    
    // Middleware
    app.Use(smallapi.Logger())
    app.Use(smallapi.CORS())
    app.Use(smallapi.Recovery())
    
    // API routes
    api := app.Group("/api/v1")
    
    // Get all users
    api.Get("/users", func(c *smallapi.Context) {
        page := c.QueryDefault("page", "1")
        limit := c.QueryDefault("limit", "10")
        
        users := store.GetAll()
        
        c.JSON(map[string]interface{}{
            "users": users,
            "total": len(users),
            "page":  page,
            "limit": limit,
        })
    })
    
    // Get user by ID
    api.Get("/users/:id", func(c *smallapi.Context) {
        id := c.Param("id")
        user := store.GetByID(id)
        
        if user == nil {
            c.Status(404).JSON(map[string]string{
                "error": "User not found",
            })
            return
        }
        
        c.JSON(user)
    })
    
    // Create user
    api.Post("/users", func(c *smallapi.Context) {
        var user User
        
        if err := c.JSON(&user); err != nil {
            c.Status(400).JSON(map[string]string{
                "error": "Invalid JSON format",
            })
            return
        }
        
        if err := c.Validate(&user); err != nil {
            c.Status(400).JSON(map[string]string{
                "error": err.Error(),
            })
            return
        }
        
        created := store.Create(&user)
        c.Status(201).JSON(created)
    })
    
    // Update user
    api.Put("/users/:id", func(c *smallapi.Context) {
        id := c.Param("id")
        
        if store.GetByID(id) == nil {
            c.Status(404).JSON(map[string]string{
                "error": "User not found",
            })
            return
        }
        
        var updates User
        if err := c.JSON(&updates); err != nil {
            c.Status(400).JSON(map[string]string{
                "error": "Invalid JSON format",
            })
            return
        }
        
        if err := c.Validate(&updates); err != nil {
            c.Status(400).JSON(map[string]string{
                "error": err.Error(),
            })
            return
        }
        
        updated := store.Update(id, &updates)
        c.JSON(updated)
    })
    
    // Delete user
    api.Delete("/users/:id", func(c *smallapi.Context) {
        id := c.Param("id")
        
        if !store.Delete(id) {
            c.Status(404).JSON(map[string]string{
                "error": "User not found",
            })
            return
        }
        
        c.Status(204)
    })
    
    app.Run(":8080")
}
