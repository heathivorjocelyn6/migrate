package migrate

import (
	"errors"
	"sync"
	"testing"
)

type mockDBImpl struct {
	mu         sync.Mutex
	locked     bool
	version    int
	dirty      bool
	versionSet bool
}

func (m *mockDBImpl) Lock() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.locked {
		return errors.New("already locked")
	}
	m.locked = true
	return nil
}

func (m *mockDBImpl) Unlock() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.locked = false
	return nil
}

func (m *mockDBImpl) SetVersion(version int, dirty bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.version = version
	m.dirty = dirty
	m.versionSet = true
	return nil
}

func TestNormalMigration(t *testing.T) {
	db := &mockDBImpl{}
	m := New(db, 1)

	err := m.Run(func() error { return nil })
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if db.dirty {
		t.Fatal("expected dirty to be false after successful migration")
	}
	if db.version != 1 {
		t.Fatalf("expected version 1, got %d", db.version)
	}
}

func TestLockTimeoutDoesNotMarkDirty(t *testing.T) {
	db := &mockDBImpl{}
	m := New(db, 1)

	// Hold the lock so AcquireLock times out
	m.mu.Lock()

	err := m.Run(func() error { return nil })
	if err != ErrLockTimeout {
		t.Fatalf("expected ErrLockTimeout, got %v", err)
	}
	if db.versionSet {
		t.Fatal("expected SetVersion to not be called on lock timeout")
	}
	m.mu.Unlock()
}

func TestLockAcquiredMigrationFailsMarksDirty(t *testing.T) {
	db := &mockDBImpl{}
	m := New(db, 1)

	err := m.Run(func() error { return errors.New("migration failed") })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !db.dirty {
		t.Fatal("expected dirty to be true after failed migration")
	}
	if db.version != 1 {
		t.Fatalf("expected version 1, got %d", db.version)
	}
}
