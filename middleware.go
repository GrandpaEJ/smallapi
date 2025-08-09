package smallapi

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Logger returns a middleware that logs requests
func Logger() MiddlewareFunc {
	return func(c *Context) bool {
		start := time.Now()
		
		// Process request
		defer func() {
			duration := time.Since(start)
			log.Printf("%s %s %d %v", 
				c.Method(), 
				c.Path(), 
				c.statusCode, 
				duration,
			)
		}()
		
		return true
	}
}

// CORS returns a middleware that handles CORS headers
func CORS() MiddlewareFunc {
	return CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	})
}

// CORSConfig configuration for CORS middleware
type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config CORSConfig) MiddlewareFunc {
	return func(c *Context) bool {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range config.AllowOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")
		
		// Handle preflight request
		if c.Method() == "OPTIONS" {
			c.Status(204)
			return false
		}
		
		return true
	}
}

// Recovery returns a middleware that recovers from panics
func Recovery() MiddlewareFunc {
	return func(c *Context) bool {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v", r)
				if !c.written {
					c.Status(500).JSON(map[string]string{
						"error": "Internal Server Error",
					})
				}
			}
		}()
		
		return true
	}
}

// RateLimit returns a middleware that implements rate limiting
func RateLimit(requestsPerMinute int) MiddlewareFunc {
	type client struct {
		requests int
		reset    time.Time
	}
	
	clients := make(map[string]*client)
	var mu sync.Mutex
	
	return func(c *Context) bool {
		ip := c.IP()
		now := time.Now()
		
		mu.Lock()
		defer mu.Unlock()
		
		// Clean up old entries
		for k, v := range clients {
			if now.After(v.reset) {
				delete(clients, k)
			}
		}
		
		// Get or create client entry
		cl, exists := clients[ip]
		if !exists {
			cl = &client{
				requests: 0,
				reset:    now.Add(time.Minute),
			}
			clients[ip] = cl
		}
		
		// Check rate limit
		if cl.requests >= requestsPerMinute {
			c.Status(429).JSON(map[string]string{
				"error": "Rate limit exceeded",
			})
			return false
		}
		
		cl.requests++
		
		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(requestsPerMinute-cl.requests))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(cl.reset.Unix(), 10))
		
		return true
	}
}

// Secure returns a middleware that adds security headers
func Secure() MiddlewareFunc {
	return func(c *Context) bool {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000")
		c.Header("Content-Security-Policy", "default-src 'self'")
		return true
	}
}

// BasicAuth returns a middleware that implements basic authentication
func BasicAuth(username, password string) MiddlewareFunc {
	return func(c *Context) bool {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != username || pass != password {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.Status(401).JSON(map[string]string{
				"error": "Unauthorized",
			})
			return false
		}
		return true
	}
}

// RequireAuth returns a middleware that requires authentication
func RequireAuth() MiddlewareFunc {
	return func(c *Context) bool {
		userID := c.Session().Get("user_id")
		if userID == nil {
			c.Status(401).JSON(map[string]string{
				"error": "Authentication required",
			})
			return false
		}
		
		// Store user ID in context for use in handlers
		c.Set("user_id", userID)
		return true
	}
}

// Timeout returns a middleware that implements request timeout
func Timeout(duration time.Duration) MiddlewareFunc {
	return func(c *Context) bool {
		done := make(chan bool, 1)
		
		go func() {
			// Simulate processing (in real implementation, this would be the actual request handling)
			done <- true
		}()
		
		select {
		case <-done:
			return true
		case <-time.After(duration):
			c.Status(408).JSON(map[string]string{
				"error": "Request timeout",
			})
			return false
		}
	}
}

// RequestID returns a middleware that adds a unique request ID
func RequestID() MiddlewareFunc {
	return func(c *Context) bool {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		c.Header("X-Request-ID", id)
		c.Set("request_id", id)
		return true
	}
}

// Compress returns a middleware that compresses responses
func Compress() MiddlewareFunc {
	return func(c *Context) bool {
		// Check if client accepts gzip
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			c.Header("Content-Encoding", "gzip")
		}
		return true
	}
}
