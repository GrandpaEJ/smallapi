# Migration from Flask to SmallAPI

This guide helps Flask developers transition to SmallAPI by showing equivalent patterns and highlighting the similarities between the two frameworks.

## Table of Contents

- [Philosophy](#philosophy)
- [Basic Application Structure](#basic-application-structure)
- [Routing](#routing)
- [Request Handling](#request-handling)
- [Response Types](#response-types)
- [Templates](#templates)
- [Sessions](#sessions)
- [Middleware](#middleware)
- [Error Handling](#error-handling)
- [Static Files](#static-files)
- [Configuration](#configuration)
- [Migration Checklist](#migration-checklist)

## Philosophy

SmallAPI is designed to bring Flask's developer experience to Go. If you love Flask's simplicity, you'll feel right at home with SmallAPI.

**Flask Philosophy:**
- Microframework with minimal boilerplate
- Explicit is better than implicit
- Simple things should be simple

**SmallAPI Philosophy:**
- Same Flask simplicity, Go performance
- Zero dependencies, maximum compatibility
- Python-like ease, Go-like speed

## Basic Application Structure

### Flask
```python
from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return {'hello': 'world'}

if __name__ == '__main__':
    app.run(debug=True, port=8080)
