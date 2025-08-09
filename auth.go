package smallapi

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"
)

// User represents a user in the authentication system
type User struct {
	ID       string                 `json:"id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Password string                 `json:"-"` // Never include in JSON
	Data     map[string]interface{} `json:"data,omitempty"`
	Created  time.Time              `json:"created"`
}

// AuthManager handles authentication
type AuthManager struct {
	users    map[string]*User // In production, this would be a database
	sessions map[string]*User // Active sessions
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return &AuthManager{
		users:    make(map[string]*User),
		sessions: make(map[string]*User),
	}
}

// Register creates a new user account
func (am *AuthManager) Register(username, email, password string) (*User, error) {
	// Check if user already exists
	for _, user := range am.users {
		if user.Username == username || user.Email == email {
			return nil, errors.New("user already exists")
		}
	}
	
	// Generate user ID
	id, err := generateID()
	if err != nil {
		return nil, err
	}
	
	// Hash password (in production, use bcrypt)
	hashedPassword := hashPassword(password)
	
	user := &User{
		ID:       id,
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Data:     make(map[string]interface{}),
		Created:  time.Now(),
	}
	
	am.users[id] = user
	return user, nil
}

// Login authenticates a user and creates a session
func (am *AuthManager) Login(username, password string) (string, *User, error) {
	var user *User
	
	// Find user by username or email
	for _, u := range am.users {
		if u.Username == username || u.Email == username {
			user = u
			break
		}
	}
	
	if user == nil {
		return "", nil, errors.New("user not found")
	}
	
	// Verify password
	if !verifyPassword(user.Password, password) {
		return "", nil, errors.New("invalid password")
	}
	
	// Create session token
	token, err := generateID()
	if err != nil {
		return "", nil, err
	}
	
	am.sessions[token] = user
	return token, user, nil
}

// Logout removes a session
func (am *AuthManager) Logout(token string) {
	delete(am.sessions, token)
}

// GetUser returns a user by session token
func (am *AuthManager) GetUser(token string) *User {
	return am.sessions[token]
}

// ChangePassword changes a user's password
func (am *AuthManager) ChangePassword(userID, oldPassword, newPassword string) error {
	user, exists := am.users[userID]
	if !exists {
		return errors.New("user not found")
	}
	
	if !verifyPassword(user.Password, oldPassword) {
		return errors.New("invalid old password")
	}
	
	user.Password = hashPassword(newPassword)
	return nil
}

// UpdateUser updates user information
func (am *AuthManager) UpdateUser(userID string, updates map[string]interface{}) error {
	user, exists := am.users[userID]
	if !exists {
		return errors.New("user not found")
	}
	
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	
	if data, ok := updates["data"].(map[string]interface{}); ok {
		for k, v := range data {
			user.Data[k] = v
		}
	}
	
	return nil
}

// DeleteUser removes a user account
func (am *AuthManager) DeleteUser(userID string) error {
	if _, exists := am.users[userID]; !exists {
		return errors.New("user not found")
	}
	
	delete(am.users, userID)
	
	// Remove all sessions for this user
	for token, user := range am.sessions {
		if user.ID == userID {
			delete(am.sessions, token)
		}
	}
	
	return nil
}

// ListUsers returns all users (admin function)
func (am *AuthManager) ListUsers() []*User {
	users := make([]*User, 0, len(am.users))
	for _, user := range am.users {
		users = append(users, user)
	}
	return users
}

// generateID generates a random ID
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashPassword hashes a password (simplified - use bcrypt in production)
func hashPassword(password string) string {
	// In production, use bcrypt.GenerateFromPassword
	return base64.StdEncoding.EncodeToString([]byte(password))
}

// verifyPassword verifies a password against a hash
func verifyPassword(hash, password string) bool {
	// In production, use bcrypt.CompareHashAndPassword
	decoded, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}
	return string(decoded) == password
}

// Auth returns a middleware that provides authentication
func Auth(authManager *AuthManager) MiddlewareFunc {
	return func(c *Context) bool {
		token := c.Session().Get("auth_token")
		if token == nil {
			return true // Continue without authentication
		}
		
		user := authManager.GetUser(token.(string))
		if user != nil {
			c.Set("user", user)
			c.Set("user_id", user.ID)
		}
		
		return true
	}
}

// RequireUser returns a middleware that requires a logged-in user
func RequireUser(authManager *AuthManager) MiddlewareFunc {
	return func(c *Context) bool {
		token := c.Session().Get("auth_token")
		if token == nil {
			c.Status(401).JSON(map[string]string{
				"error": "Authentication required",
			})
			return false
		}
		
		user := authManager.GetUser(token.(string))
		if user == nil {
			c.Status(401).JSON(map[string]string{
				"error": "Invalid session",
			})
			return false
		}
		
		c.Set("user", user)
		c.Set("user_id", user.ID)
		return true
	}
}
