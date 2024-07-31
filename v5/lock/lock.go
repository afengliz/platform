package lock

import "sync"

type Table struct {
	mu       sync.RWMutex
	rowLocks map[string]*sync.RWMutex
}

func NewTable() *Table {
	return &Table{
		rowLocks: make(map[string]*sync.RWMutex),
	}
}

func (t *Table) getRowLock(rowID string) *sync.RWMutex {
	t.mu.RLock()
	lock, exists := t.rowLocks[rowID]
	t.mu.RUnlock()

	if !exists {
		t.mu.Lock()
		lock, exists = t.rowLocks[rowID]
		if !exists {
			lock = &sync.RWMutex{}
			t.rowLocks[rowID] = lock
		}
		t.mu.Unlock()
	}
	return lock
}

func (t *Table) LockRowForRead(rowID string) {
	lock := t.getRowLock(rowID)
	lock.RLock()
}

func (t *Table) UnlockRowForRead(rowID string) {
	lock := t.getRowLock(rowID)
	lock.RUnlock()
}

func (t *Table) LockRowForWrite(rowID string) {
	lock := t.getRowLock(rowID)
	lock.Lock()
}

func (t *Table) UnlockRowForWrite(rowID string) {
	lock := t.getRowLock(rowID)
	lock.Unlock()
}
