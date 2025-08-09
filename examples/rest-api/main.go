// REST API example demonstrating CRUD operations with validation
package main

import (
	"strconv"
	"time"
	"github.com/grandpaej/smallapi"
)

// Task represents a todo task
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title" validate:"required,min=1,max=100"`
	Description string    `json:"description" validate:"max=500"`
	Completed   bool      `json:"completed"`
	Priority    string    `json:"priority" validate:"regex=^(low|medium|high)$"`
	DueDate     *string   `json:"due_date,omitempty"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// TaskStore provides in-memory storage for tasks
type TaskStore struct {
	tasks  map[string]*Task
	nextID int
}

// NewTaskStore creates a new task store
func NewTaskStore() *TaskStore {
	store := &TaskStore{
		tasks:  make(map[string]*Task),
		nextID: 1,
	}
	
	// Add some sample tasks
	store.CreateTask(&Task{
		Title:       "Learn SmallAPI",
		Description: "Go through the documentation and examples",
		Priority:    "high",
		Completed:   false,
	})
	
	store.CreateTask(&Task{
		Title:       "Build REST API",
		Description: "Create a todo API using SmallAPI",
		Priority:    "medium",
		Completed:   true,
	})
	
	return store
}

// CreateTask creates a new task
func (s *TaskStore) CreateTask(task *Task) *Task {
	task.ID = strconv.Itoa(s.nextID)
	s.nextID++
	task.Created = time.Now()
	task.Updated = time.Now()
	s.tasks[task.ID] = task
	return task
}

// GetAllTasks returns all tasks with optional filtering
func (s *TaskStore) GetAllTasks(completed *bool, priority string) []*Task {
	var tasks []*Task
	
	for _, task := range s.tasks {
		// Filter by completion status
		if completed != nil && task.Completed != *completed {
			continue
		}
		
		// Filter by priority
		if priority != "" && task.Priority != priority {
			continue
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks
}

// GetTaskByID returns a task by its ID
func (s *TaskStore) GetTaskByID(id string) *Task {
	return s.tasks[id]
}

// UpdateTask updates an existing task
func (s *TaskStore) UpdateTask(id string, updates *Task) *Task {
	task := s.tasks[id]
	if task == nil {
		return nil
	}
	
	// Update fields
	if updates.Title != "" {
		task.Title = updates.Title
	}
	if updates.Description != "" {
		task.Description = updates.Description
	}
	if updates.Priority != "" {
		task.Priority = updates.Priority
	}
	if updates.DueDate != nil {
		task.DueDate = updates.DueDate
	}
	
	task.Completed = updates.Completed
	task.Updated = time.Now()
	
	return task
}

// DeleteTask deletes a task by ID
func (s *TaskStore) DeleteTask(id string) bool {
	if _, exists := s.tasks[id]; exists {
		delete(s.tasks, id)
		return true
	}
	return false
}

func main() {
	app := smallapi.New()
	store := NewTaskStore()
	
	// Middleware
	app.Use(smallapi.Logger())
	app.Use(smallapi.CORS())
	app.Use(smallapi.Recovery())
	
	// API routes
	api := app.Group("/api/v1")
	
	// Get all tasks with optional filters
	api.Get("/tasks", func(c *smallapi.Context) {
		// Parse query parameters
		var completed *bool
		if c.Query("completed") != "" {
			val := c.Query("completed") == "true"
			completed = &val
		}
		
		priority := c.Query("priority")
		
		tasks := store.GetAllTasks(completed, priority)
		
		c.JSON(map[string]interface{}{
			"tasks": tasks,
			"total": len(tasks),
		})
	})
	
	// Get task by ID
	api.Get("/tasks/:id", func(c *smallapi.Context) {
		id := c.Param("id")
		task := store.GetTaskByID(id)
		
		if task == nil {
			c.Status(404).JSON(map[string]string{
				"error": "Task not found",
			})
			return
		}
		
		c.JSON(task)
	})
	
	// Create new task
	api.Post("/tasks", func(c *smallapi.Context) {
		var task Task
		
		if err := c.JSON(&task); err != nil {
			c.Status(400).JSON(map[string]string{
				"error": "Invalid JSON format",
			})
			return
		}
		
		if err := c.Validate(&task); err != nil {
			c.Status(400).JSON(map[string]string{
				"error": err.Error(),
			})
			return
		}
		
		created := store.CreateTask(&task)
		c.Status(201).JSON(created)
	})
	
	// Update task
	api.Put("/tasks/:id", func(c *smallapi.Context) {
		id := c.Param("id")
		
		if store.GetTaskByID(id) == nil {
			c.Status(404).JSON(map[string]string{
				"error": "Task not found",
			})
			return
		}
		
		var updates Task
		if err := c.JSON(&updates); err != nil {
			c.Status(400).JSON(map[string]string{
				"error": "Invalid JSON format",
			})
			return
		}
		
		updated := store.UpdateTask(id, &updates)
		c.JSON(updated)
	})
	
	// Delete task
	api.Delete("/tasks/:id", func(c *smallapi.Context) {
		id := c.Param("id")
		
		if !store.DeleteTask(id) {
			c.Status(404).JSON(map[string]string{
				"error": "Task not found",
			})
			return
		}
		
		c.Status(204).String("")
	})
	
	// Mark task as completed
	api.Post("/tasks/:id/complete", func(c *smallapi.Context) {
		id := c.Param("id")
		task := store.GetTaskByID(id)
		
		if task == nil {
			c.Status(404).JSON(map[string]string{
				"error": "Task not found",
			})
			return
		}
		
		updates := &Task{Completed: true}
		updated := store.UpdateTask(id, updates)
		
		c.JSON(map[string]interface{}{
			"message": "Task marked as completed",
			"task":    updated,
		})
	})
	
	// Get task statistics
	api.Get("/tasks/stats", func(c *smallapi.Context) {
		allTasks := store.GetAllTasks(nil, "")
		
		stats := map[string]interface{}{
			"total":     len(allTasks),
			"completed": 0,
			"pending":   0,
			"priority": map[string]int{
				"high":   0,
				"medium": 0,
				"low":    0,
			},
		}
		
		for _, task := range allTasks {
			if task.Completed {
				stats["completed"] = stats["completed"].(int) + 1
			} else {
				stats["pending"] = stats["pending"].(int) + 1
			}
			
			priorityMap := stats["priority"].(map[string]int)
			priorityMap[task.Priority]++
		}
		
		c.JSON(stats)
	})
	
	// Health check endpoint
	app.Get("/health", func(c *smallapi.Context) {
		c.JSON(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		})
	})
	
	// Start server
	app.Run(":8080")
}
