# Overview

SmallAPI is a Go web framework designed to provide Flask-like simplicity with Go's performance benefits. It's a zero-dependency framework built entirely on Go's standard library, offering production-ready features including hot reload for development, built-in authentication, WebSocket support, templating, and API documentation generation. The framework aims to make Go web development as intuitive as Python Flask while maintaining Go's type safety and performance characteristics.

# User Preferences

Preferred communication style: Simple, everyday language.

# System Architecture

## Core Framework Design
- **Zero-dependency architecture**: Built entirely on Go standard library to ensure maximum compatibility and minimal external dependencies
- **Flask-inspired API design**: Uses familiar patterns like `app.Get("/", handler)` to reduce learning curve for developers coming from Python
- **Context-based request handling**: Centralizes request/response operations through a `Context` object similar to modern web frameworks

## Application Structure
- **Modular application instance**: Core `App` struct manages routing, middleware, and server lifecycle
- **Handler-based routing**: Uses function-based handlers that receive a context object for request processing
- **Middleware chain**: Supports composable middleware for cross-cutting concerns like authentication, logging, and rate limiting

## Development Features
- **Hot reload system**: Automatically reloads the application when source files change during development
- **Built-in templating**: Auto-loading HTML templates with Go's template engine integration
- **Static file serving**: One-line configuration for serving static assets

## Production Features
- **Built-in security**: Includes security middleware and best practices out of the box
- **Rate limiting**: Built-in request rate limiting to protect against abuse
- **Monitoring capabilities**: Health checks and metrics collection for production deployment
- **Session management**: Built-in session handling and authentication middleware

## API and Communication
- **RESTful API support**: Full CRUD operations with JSON request/response handling
- **WebSocket integration**: Easy WebSocket upgrade capabilities for real-time communication
- **Auto-generated documentation**: Automatic API documentation generation from route definitions
- **Request validation**: Built-in validation system for incoming data

## Template and Frontend
- **Template engine integration**: Uses Go's built-in template system for server-side rendering
- **Static asset management**: Efficient static file serving with proper caching headers
- **Responsive design support**: Template examples include modern CSS and responsive layouts

# External Dependencies

## Core Dependencies
- **Go standard library only**: No external Go modules required for core functionality
- **Go 1.19+**: Minimum Go version requirement for modern language features

## Optional Integrations
- **Database systems**: Framework-agnostic design allows integration with any Go database driver
- **External authentication providers**: Extensible authentication system supports OAuth and other external auth systems
- **Monitoring services**: Compatible with standard Go monitoring and observability tools
- **Load balancers**: Standard HTTP server implementation works with any reverse proxy or load balancer

## Development Tools
- **File watching system**: Uses OS-level file system notifications for hot reload functionality
- **Template compilation**: Leverages Go's built-in template compilation for server-side rendering
- **Static analysis**: Compatible with standard Go tooling for linting, testing, and code analysis