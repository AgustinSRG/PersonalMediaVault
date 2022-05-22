// Read / Write lock for storage files

// Read sequence:
//  1 - StartRead()
//  2 - Perform read
//  3 - EndRead()

// Write sequence:
//  1 - RequestWrite()
//  2 - Copy resource to temp file
//  3 - Perform changes
//  4 - StartWrite()
//  5 - Copy temp files into original files
//  6 - EndWrite()

package main

import (
	"sync"
)

type ReadWriteLock struct {
	lock *sync.Mutex

	// Read

	read_count int

	read_wait_count int

	read_wait_locks [](*sync.Mutex)

	// Write

	writing bool

	write_sem *sync.Mutex

	write_wait     bool
	write_wait_sem *sync.Mutex
}

// Creates a lock to manage a resource
func CreateReadWriteLock() *ReadWriteLock {
	return &ReadWriteLock{
		lock: &sync.Mutex{},

		read_count:      0,
		read_wait_count: 0,
		read_wait_locks: make([](*sync.Mutex), 0),

		writing:        false,
		write_wait:     false,
		write_sem:      &sync.Mutex{},
		write_wait_sem: &sync.Mutex{},
	}
}

func (lock *ReadWriteLock) RequestWrite() {
	lock.write_sem.Lock()
}

func (lock *ReadWriteLock) StartWrite() {
	lock.lock.Lock()

	mustWait := false

	if lock.read_count > 0 {
		lock.write_wait = true
		mustWait = true
		lock.write_wait_sem.Lock()
	} else {
		lock.writing = true
	}

	lock.lock.Unlock()

	if mustWait {
		// Lock 2 times, so it's always blocking until other thread unlocks it
		lock.write_wait_sem.Lock()

		lock.write_wait_sem.Unlock()
	}
}

func (lock *ReadWriteLock) EndWrite() {
	lock.lock.Lock()

	lock.writing = false

	if lock.read_wait_count > 0 {
		lock.read_count += lock.read_wait_count
		lock.read_wait_count = 0

		// Release all locks

		for i := 0; i < len(lock.read_wait_locks); i++ {
			lock.read_wait_locks[i].Unlock()
		}

		lock.read_wait_locks = make([]*sync.Mutex, 0) // Clear list
	}

	lock.lock.Unlock()

	lock.write_sem.Unlock()
}

func (lock *ReadWriteLock) StartRead() {
	lock.lock.Lock()

	mustWait := false
	var readLock *sync.Mutex

	if lock.writing || lock.write_wait {
		lock.read_wait_count++
		mustWait = true
		// Create a mutex and append it to the waiter list
		readLock = &sync.Mutex{}
		readLock.Lock()
		lock.read_wait_locks = append(lock.read_wait_locks, readLock)
	} else {
		lock.read_count++
	}

	lock.lock.Unlock()

	if mustWait {
		// Lock 2 times to ensure blocking until the write thread releases it
		readLock.Lock()

		readLock.Unlock()
	}
}

func (lock *ReadWriteLock) EndRead() {
	lock.lock.Lock()

	lock.read_count--

	if lock.write_wait && lock.read_count <= 0 {
		lock.write_wait = false
		lock.writing = true
		lock.write_wait_sem.Unlock()
	}

	lock.lock.Unlock()
}
