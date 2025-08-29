package bot

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProcessLock prevents multiple instances of the bot from running simultaneously
type ProcessLock struct {
	lockFile string
	file     *os.File
}

// NewProcessLock creates a new process lock instance
func NewProcessLock(lockFile string) *ProcessLock {
	// Ensure lock file is in a writable directory
	if lockFile == "" {
		lockFile = filepath.Join(os.TempDir(), "arcanus-vpn-bot.lock")
	}

	return &ProcessLock{lockFile: lockFile}
}

// Acquire attempts to acquire the process lock
func (pl *ProcessLock) Acquire() error {
	// Try to create the lock file exclusively
	file, err := os.OpenFile(pl.lockFile, os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("another instance is already running (lock file: %s)", pl.lockFile)
		}
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	// Write PID to lock file for debugging
	pid := fmt.Sprintf("%d\n", os.Getpid())
	if _, err := file.WriteString(pid); err != nil {
		_ = file.Close()
		_ = os.Remove(pl.lockFile)
		return fmt.Errorf("failed to write PID to lock file: %w", err)
	}

	pl.file = file
	return nil
}

// Release releases the process lock
func (pl *ProcessLock) Release() error {
	if pl.file != nil {
		_ = pl.file.Close()
		if err := os.Remove(pl.lockFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove lock file: %w", err)
		}
		pl.file = nil
	}
	return nil
}

// IsLocked checks if another instance is running
func (pl *ProcessLock) IsLocked() bool {
	_, err := os.Stat(pl.lockFile)
	return err == nil
}

// GetLockInfo returns information about the current lock
func (pl *ProcessLock) GetLockInfo() (string, error) {
	if !pl.IsLocked() {
		return "", nil
	}

	data, err := os.ReadFile(pl.lockFile)
	if err != nil {
		return "", fmt.Errorf("failed to read lock file: %w", err)
	}

	return string(data), nil
}
