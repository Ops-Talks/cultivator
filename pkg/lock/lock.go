package lock

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ParseDuration parses a duration string (like "10m", "1h")
func ParseDuration(s string) time.Duration {
	if s == "" {
		return 10 * time.Minute
	}
	
	d, err := time.ParseDuration(s)
	if err != nil {
		return 10 * time.Minute
	}
	
	return d
}

// Lock represents a lock on a module
type Lock struct {
	Module     string
	User       string
	PR         int
	AcquiredAt time.Time
}

// Manager manages locks for Terragrunt modules
type Manager struct {
	locks   map[string]*Lock
	mu      sync.RWMutex
	timeout time.Duration
}

// NewManager creates a new lock manager
func NewManager(timeout time.Duration) *Manager {
	return &Manager{
		locks:   make(map[string]*Lock),
		timeout: timeout,
	}
}

// Acquire attempts to acquire a lock for a module
func (m *Manager) Acquire(ctx context.Context, module, user string, pr int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already locked
	if lock, exists := m.locks[module]; exists {
		// Check if lock has expired
		if time.Since(lock.AcquiredAt) > m.timeout {
			delete(m.locks, module)
		} else {
			return fmt.Errorf("module %s is locked by %s (PR #%d)", module, lock.User, lock.PR)
		}
	}

	// Acquire lock
	m.locks[module] = &Lock{
		Module:     module,
		User:       user,
		PR:         pr,
		AcquiredAt: time.Now(),
	}

	return nil
}

// Release releases a lock for a module
func (m *Manager) Release(module string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.locks[module]; !exists {
		return fmt.Errorf("no lock found for module %s", module)
	}

	delete(m.locks, module)
	return nil
}

// IsLocked checks if a module is locked
func (m *Manager) IsLocked(module string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lock, exists := m.locks[module]
	if !exists {
		return false
	}

	// Check if lock has expired
	if time.Since(lock.AcquiredAt) > m.timeout {
		return false
	}

	return true
}

// GetLock returns the lock information for a module
func (m *Manager) GetLock(module string) (*Lock, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lock, exists := m.locks[module]
	if !exists {
		return nil, fmt.Errorf("no lock found for module %s", module)
	}

	return lock, nil
}

// CleanExpired removes all expired locks
func (m *Manager) CleanExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for module, lock := range m.locks {
		if time.Since(lock.AcquiredAt) > m.timeout {
			delete(m.locks, module)
		}
	}
}
