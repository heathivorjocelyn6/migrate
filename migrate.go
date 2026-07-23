package migrate

import (
	"errors"
	"sync"
	"time"
)

var ErrLockTimeout = errors.New("lock timeout")

type mockDB interface {
	Lock() error
	Unlock() error
	SetVersion(version int, dirty bool) error
}

type Migrator struct {
	mu      sync.Mutex
	db      mockDB
	version int
}

func New(db mockDB, version int) *Migrator {
	return &Migrator{db: db, version: version}
}

func (m *Migrator) AcquireLock(timeout time.Duration) (bool, error) {
	ch := make(chan struct{}, 1)
	go func() {
		m.mu.Lock()
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		return true, nil
	case <-time.After(timeout):
		return false, ErrLockTimeout
	}
}

func (m *Migrator) Run(migration func() error) error {
	acquired, err := m.AcquireLock(5 * time.Second)
	if err != nil || !acquired {
		return ErrLockTimeout
	}
	defer m.mu.Unlock()

	if err := m.db.SetVersion(m.version, true); err != nil {
		return err
	}

	if err := migration(); err != nil {
		return err
	}

	return m.db.SetVersion(m.version, false)
}
