// Template example demonstrating server-side rendering with SmallAPI
package main

import (
        "log"
        "time"
        "github.com/grandpaej/smallapi"
)

// PageData represents data passed to templates
type PageData struct {
        Title       string
        CurrentTime string
        User        *User
        Items       []Item
}

// User represents a user
type User struct {
        Name  string
        Email string
        Role  string
}

// Item represents a list item
type Item struct {
        ID          int
        Name        string
        Description string
        Price       float64
        InStock     bool
}

func main() {
        app := smallapi.New()
        
        // Configure templates directory
        err := app.Templates("./views")
        if err != nil {
                log.Fatal("Template loading error:", err)
        }
        
        // Serve static files for CSS, JS, images
        app.Static("/static", "./static")
        
        // Sample data
        sampleUser := &User{
                Name:  "John Doe",
                Email: "john@example.com",
                Role:  "Admin",
        }
        
        sampleItems := []Item{
                {ID: 1, Name: "Laptop", Description: "High-performance laptop", Price: 999.99, InStock: true},
                {ID: 2, Name: "Mouse", Description: "Wireless optical mouse", Price: 29.99, InStock: true},
                {ID: 3, Name: "Keyboard", Description: "Mechanical keyboard", Price: 149.99, InStock: false},
                {ID: 4, Name: "Monitor", Description: "4K display monitor", Price: 299.99, InStock: true},
        }
        
        // Home page
        app.Get("/", func(c *smallapi.Context) {
                data := PageData{
                        Title:       "SmallAPI Template Example",
                        CurrentTime: time.Now().Format("January 2, 2006 15:04:05"),
                        User:        sampleUser,
                        Items:       sampleItems,
                }
                
                if err := c.Render("index.html", data); err != nil {
                        log.Printf("Template render error: %v", err)
                        c.Status(500).String("Template error: " + err.Error())
                }
        })
        
        // About page
        app.Get("/about", func(c *smallapi.Context) {
                data := PageData{
                        Title:       "About - SmallAPI",
                        CurrentTime: time.Now().Format("January 2, 2006 15:04:05"),
                        User:        sampleUser,
                }
                
                if err := c.Render("about.html", data); err != nil {
                        log.Printf("Template render error: %v", err)
                        c.Status(500).String("Template error: " + err.Error())
                }
        })
        
        // Items page with filtering
        app.Get("/items", func(c *smallapi.Context) {
                inStockOnly := c.Query("in_stock") == "true"
                filteredItems := sampleItems
                
                if inStockOnly {
                        filteredItems = []Item{}
                        for _, item := range sampleItems {
                                if item.InStock {
                                        filteredItems = append(filteredItems, item)
                                }
                        }
                }
                
                data := PageData{
                        Title:       "Items - SmallAPI",
                        CurrentTime: time.Now().Format("January 2, 2006 15:04:05"),
                        User:        sampleUser,
                        Items:       filteredItems,
                }
                
                c.Render("items.html", data)
        })
        
        // Item detail page
        app.Get("/items/:id", func(c *smallapi.Context) {
                id, err := c.ParamInt("id")
                if err != nil {
                        c.Status(400).HTML("<h1>Invalid item ID</h1>")
                        return
                }
                
                var foundItem *Item
                for _, item := range sampleItems {
                        if item.ID == id {
                                foundItem = &item
                                break
                        }
                }
                
                if foundItem == nil {
                        c.Status(404).HTML("<h1>Item not found</h1>")
                        return
                }
                
                data := struct {
                        PageData
                        Item Item
                }{
                        PageData: PageData{
                                Title:       foundItem.Name + " - SmallAPI",
                                CurrentTime: time.Now().Format("January 2, 2006 15:04:05"),
                                User:        sampleUser,
                        },
                        Item: *foundItem,
                }
                
                c.Render("item-detail.html", data)
        })
        
        // Contact form (GET - show form, POST - handle submission)
        app.Get("/contact", func(c *smallapi.Context) {
                data := PageData{
                        Title:       "Contact Us - SmallAPI",
                        CurrentTime: time.Now().Format("January 2, 2006 15:04:05"),
                        User:        sampleUser,
                }
                
                c.Render("contact.html", data)
        })
        
        app.Post("/contact", func(c *smallapi.Context) {
                name := c.Form("name")
                email := c.Form("email")
                message := c.Form("message")
                
                // In a real app, you'd save this to a database or send an email
                responseData := struct {
                        PageData
                        Success bool
                        Message string
                        FormData map[string]string
                }{
                        PageData: PageData{
                                Title:       "Contact Submitted - SmallAPI",
                                CurrentTime: time.Now().Format("January 2, 2006 15:04:05"),
                                User:        sampleUser,
                        },
                        Success: true,
                        Message: "Thank you for your message! We'll get back to you soon.",
                        FormData: map[string]string{
                                "name":    name,
                                "email":   email,
                                "message": message,
                        },
                }
                
                c.Render("contact-success.html", responseData)
        })
        
        // API endpoint that returns JSON (for AJAX calls)
        app.Get("/api/items", func(c *smallapi.Context) {
                c.JSON(map[string]interface{}{
                        "items": sampleItems,
                        "total": len(sampleItems),
                })
        })
        
        // Start server in development mode
        app.RunDev(":8080")
}
