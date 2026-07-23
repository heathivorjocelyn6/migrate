package main

import (
	"context"
	"fmt"
)

// Mocking the migration structure described in the issue
type Database interface {
	Lock() error
	Unlock() error
	SetVersion(version int, dirty bool) error
}

func RunMigration(ctx context.Context, db Database, version int) error {
	if err := db.Lock(); err != nil {
		return err
	}
	defer db.Unlock()

	// Check if context was cancelled while waiting for the lock
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Only now is it safe to mark the database as dirty
	if err := db.SetVersion(version, true); err != nil {
		return err
	}

	return nil
}

func main() {
	fmt.Println("Hello, Bounty Hunter!")
}