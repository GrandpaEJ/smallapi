package smallapi

import (
        "crypto/rand"
        "encoding/base64"
        "net/http"
        "sync"
        "time"
)

// Session represents a user session
type Session struct {
        id     string
        data   map[string]interface{}
        mutex  sync.RWMutex
        maxAge time.Duration
}

// SessionManager manages user sessions
type SessionManager struct {
        sessions map[string]*Session
        mutex    sync.RWMutex
        maxAge   time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
        sm := &SessionManager{
                sessions: make(map[string]*Session),
                maxAge:   24 * time.Hour, // 24 hours default
        }
        
        // Start cleanup routine
        go sm.cleanup()
        
        return sm
}

// GetSession gets or creates a session for a request
func (sm *SessionManager) GetSession(r *http.Request, w http.ResponseWriter) *Session {
        // Try to get session ID from cookie
        cookie, err := r.Cookie("session_id")
        var sessionID string
        
        if err != nil || cookie.Value == "" {
                // Create new session
                sessionID = sm.generateSessionID()
                sm.setSessionCookie(w, sessionID)
        } else {
                sessionID = cookie.Value
        }
        
        sm.mutex.Lock()
        defer sm.mutex.Unlock()
        
        session, exists := sm.sessions[sessionID]
        if !exists {
                session = &Session{
                        id:     sessionID,
                        data:   make(map[string]interface{}),
                        maxAge: sm.maxAge,
                }
                sm.sessions[sessionID] = session
        }
        
        return session
}

// generateSessionID generates a unique session ID
func (sm *SessionManager) generateSessionID() string {
        bytes := make([]byte, 32)
        rand.Read(bytes)
        return base64.URLEncoding.EncodeToString(bytes)
}

// setSessionCookie sets the session cookie
func (sm *SessionManager) setSessionCookie(w http.ResponseWriter, sessionID string) {
        cookie := &http.Cookie{
                Name:     "session_id",
                Value:    sessionID,
                Path:     "/",
                MaxAge:   int(sm.maxAge.Seconds()),
                HttpOnly: true,
                Secure:   false, // Set to true in production with HTTPS
                SameSite: http.SameSiteLaxMode,
        }
        http.SetCookie(w, cookie)
}

// cleanup removes expired sessions
func (sm *SessionManager) cleanup() {
        ticker := time.NewTicker(time.Hour)
        defer ticker.Stop()
        
        for range ticker.C {
                sm.mutex.Lock()
                // In a real implementation, you'd track session creation times
                // and remove expired sessions
                sm.mutex.Unlock()
        }
}

// Set stores a value in the session
func (s *Session) Set(key string, value interface{}) {
        s.mutex.Lock()
        defer s.mutex.Unlock()
        s.data[key] = value
}

// Get retrieves a value from the session
func (s *Session) Get(key string) interface{} {
        s.mutex.RLock()
        defer s.mutex.RUnlock()
        return s.data[key]
}

// GetString retrieves a string value from the session
func (s *Session) GetString(key string) string {
        value := s.Get(key)
        if str, ok := value.(string); ok {
                return str
        }
        return ""
}

// GetInt retrieves an int value from the session
func (s *Session) GetInt(key string) int {
        value := s.Get(key)
        if i, ok := value.(int); ok {
                return i
        }
        return 0
}

// Delete removes a value from the session
func (s *Session) Delete(key string) {
        s.mutex.Lock()
        defer s.mutex.Unlock()
        delete(s.data, key)
}

// Clear removes all values from the session
func (s *Session) Clear() {
        s.mutex.Lock()
        defer s.mutex.Unlock()
        s.data = make(map[string]interface{})
}

// ID returns the session ID
func (s *Session) ID() string {
        return s.id
}

// Has checks if a key exists in the session
func (s *Session) Has(key string) bool {
        s.mutex.RLock()
        defer s.mutex.RUnlock()
        _, exists := s.data[key]
        return exists
}

// Keys returns all keys in the session
func (s *Session) Keys() []string {
        s.mutex.RLock()
        defer s.mutex.RUnlock()
        
        keys := make([]string, 0, len(s.data))
        for key := range s.data {
                keys = append(keys, key)
        }
        return keys
}

// SessionMiddleware returns a middleware that provides session support
func SessionMiddleware() MiddlewareFunc {
        return func(c *Context) bool {
                // Session is already initialized in NewContext
                return true
        }
}
