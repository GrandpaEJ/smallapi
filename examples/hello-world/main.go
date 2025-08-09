// Hello World example - the simplest SmallAPI application
package main

import "github.com/grandpaej/smallapi"

func main() {
        // Create a new SmallAPI application
        app := smallapi.New()

        // Add a simple route that returns JSON
        app.Get("/", func(c *smallapi.Context) {
                c.JSON(map[string]string{
                        "message": "Hello, World!",
                        "framework": "SmallAPI",
                })
        })

        // Add a route with a parameter
        app.Get("/hello/:name", func(c *smallapi.Context) {
                name := c.Param("name")
                c.JSON(map[string]string{
                        "message": "Hello, " + name + "!",
                })
        })

        // Add a route that returns plain text
        app.Get("/text", func(c *smallapi.Context) {
                c.String("Hello from SmallAPI!")
        })

        // Add a route that demonstrates query parameters
        app.Get("/greet", func(c *smallapi.Context) {
                name := c.QueryDefault("name", "World")
                format := c.QueryDefault("format", "json")
                
                message := "Hello, " + name + "!"
                
                switch format {
                case "json":
                        c.JSON(map[string]string{"message": message})
                case "text":
                        c.String(message)
                case "html":
                        c.HTML("<h1>" + message + "</h1>")
                default:
                        c.Status(400).JSON(map[string]string{
                                "error": "Unsupported format. Use: json, text, or html",
                        })
                }
        })

        // Start the server on port 5000
        // Use RunDev for development with hot reload
        app.Run(":5000")
}
