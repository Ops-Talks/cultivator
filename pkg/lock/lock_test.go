package lock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLockManager_AcquireAndRelease(t *testing.T) {
	manager := NewManager(5 * time.Minute)
	ctx := context.Background()

	// Acquire lock
	err := manager.Acquire(ctx, "module1", "user1", 123)
	assert.NoError(t, err)

	// Check if locked
	assert.True(t, manager.IsLocked("module1"))

	// Try to acquire again (should fail)
	err = manager.Acquire(ctx, "module1", "user2", 456)
	assert.Error(t, err)

	// Release lock
	err = manager.Release("module1")
	assert.NoError(t, err)

	// Check if unlocked
	assert.False(t, manager.IsLocked("module1"))
}

func TestLockManager_Timeout(t *testing.T) {
	manager := NewManager(1 * time.Millisecond)
	ctx := context.Background()

	// Acquire lock
	err := manager.Acquire(ctx, "module1", "user1", 123)
	assert.NoError(t, err)

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	// Lock should be expired
	assert.False(t, manager.IsLocked("module1"))

	// Should be able to acquire again
	err = manager.Acquire(ctx, "module1", "user2", 456)
	assert.NoError(t, err)
}
