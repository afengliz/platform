package main

import (
	"fmt"
	"sync"
)

type Table struct {
	mu       sync.RWMutex
	rowLocks map[int]*sync.RWMutex
}

func NewTable() *Table {
	return &Table{
		rowLocks: make(map[int]*sync.RWMutex),
	}
}

func (t *Table) getRowLock(rowID int) *sync.RWMutex {
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

func (t *Table) LockRowForRead(rowID int) {
	lock := t.getRowLock(rowID)
	lock.RLock()
}

func (t *Table) UnlockRowForRead(rowID int) {
	lock := t.getRowLock(rowID)
	lock.RUnlock()
}

func (t *Table) LockRowForWrite(rowID int) {
	lock := t.getRowLock(rowID)
	lock.Lock()
}

func (t *Table) UnlockRowForWrite(rowID int) {
	lock := t.getRowLock(rowID)
	lock.Unlock()
}

func main() {
	table := NewTable()
	var wg sync.WaitGroup

	wg.Add(3)

	// Goroutine 1 - Read Lock
	go func() {
		defer wg.Done()
		table.LockRowForRead(1)
		fmt.Println("Goroutine 1: Read locked row 1")
		table.UnlockRowForRead(1)
		fmt.Println("Goroutine 1: Read unlocked row 1")
	}()

	// Goroutine 2 - Write Lock
	go func() {
		defer wg.Done()
		table.LockRowForWrite(1)
		fmt.Println("Goroutine 2: Write locked row 1")
		table.UnlockRowForWrite(1)
		fmt.Println("Goroutine 2: Write unlocked row 1")
	}()

	// Goroutine 3 - Read Lock
	go func() {
		defer wg.Done()
		table.LockRowForRead(1)
		fmt.Println("Goroutine 3: Read locked row 1")
		table.UnlockRowForRead(1)
		fmt.Println("Goroutine 3: Read unlocked row 1")
	}()

	wg.Wait()
}
