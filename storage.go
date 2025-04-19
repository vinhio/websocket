package main

import (
	"sync"
	"ws/data"
)

// UserStore defines the interface for user storage
type UserStore interface {
	// AddUser adds a new user to the store
	AddUser(username, password string) error

	// GetUser retrieves a user by username
	GetUser(username string) (*data.User, bool)

	// Authenticate checks if the provided credentials are valid
	Authenticate(username, password string) bool

	// ListUsers returns a list of all usernames
	ListUsers() []string
}

// InMemoryUserStore implements UserStore with an in-memory map
type InMemoryUserStore struct {
	users map[string]string // map[username]password
	mu    sync.RWMutex
}

// NewInMemoryUserStore creates a new in-memory user store
func NewInMemoryUserStore() *InMemoryUserStore {
	store := &InMemoryUserStore{
		users: make(map[string]string),
	}

	// Add some default users
	_ = store.AddUser("admin", "admin123")
	_ = store.AddUser("user1", "password1")
	_ = store.AddUser("user2", "password2")

	return store
}

// AddUser adds a new user to the store
func (s *InMemoryUserStore) AddUser(username, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[username] = password
	return nil
}

// GetUser retrieves a user by username
func (s *InMemoryUserStore) GetUser(username string) (*data.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	password, exists := s.users[username]
	if !exists {
		return nil, false
	}

	return &data.User{
		Username: username,
		Password: password,
	}, true
}

// Authenticate checks if the provided credentials are valid
func (s *InMemoryUserStore) Authenticate(username, password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storedPassword, exists := s.users[username]
	if !exists {
		return false
	}

	return storedPassword == password
}

// ListUsers returns a list of all usernames
func (s *InMemoryUserStore) ListUsers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	usernames := make([]string, 0, len(s.users))
	for username := range s.users {
		usernames = append(usernames, username)
	}

	return usernames
}

// Global instance of the user store
var GlobalUserStore UserStore = NewInMemoryUserStore()
