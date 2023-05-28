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

// Read / Write lock
// Controls access to a resource
type ReadWriteLock struct {
	lock *sync.Mutex

	// Read

	read_count int

	read_wait_count int

	read_wait_group *sync.WaitGroup

	// Write

	writing bool

	write_mutex *sync.Mutex

	write_wait       bool
	write_wait_group *sync.WaitGroup
}

// Creates a lock to manage a resource
func CreateReadWriteLock() *ReadWriteLock {
	return &ReadWriteLock{
		lock: &sync.Mutex{},

		read_count:      0,
		read_wait_count: 0,
		read_wait_group: nil,

		writing:          false,
		write_wait:       false,
		write_mutex:      &sync.Mutex{},
		write_wait_group: nil,
	}
}

// Request a write operation
// This locks the resource from writing
// Only one write thread is allowed
func (lock *ReadWriteLock) RequestWrite() {
	lock.write_mutex.Lock()
}

// Starts a write operation
// Waits for pending read threads to finish
// Locks the resource so only the write thread can use it
func (lock *ReadWriteLock) StartWrite() {
	lock.lock.Lock()

	var waitGroup *sync.WaitGroup = nil

	if lock.read_count > 0 {
		lock.write_wait = true
		if lock.write_wait_group == nil {
			lock.write_wait_group = &sync.WaitGroup{}
			lock.write_wait_group.Add(1)
		}
		waitGroup = lock.write_wait_group
	} else {
		lock.writing = true
	}

	lock.lock.Unlock()

	if waitGroup != nil {
		// Wait for read threads
		waitGroup.Wait()
	}
}

// Finish a write operation, unlocking the resource
func (lock *ReadWriteLock) EndWrite() {
	lock.lock.Lock()

	lock.writing = false

	if lock.read_wait_count > 0 {
		lock.read_count += lock.read_wait_count
		lock.read_wait_count = 0

		// Release read threads
		if lock.read_wait_group != nil {
			lock.read_wait_group.Done()
			lock.read_wait_group = nil
		}
	}

	lock.lock.Unlock()

	// Unlock write mutex, so other write threads can continue
	lock.write_mutex.Unlock()
}

// Starts a read operation
func (lock *ReadWriteLock) StartRead() {
	lock.lock.Lock()

	var waitGroup *sync.WaitGroup = nil

	if lock.writing || lock.write_wait {
		// We cannot read, we have to wait for a write thread
		lock.read_wait_count++
		if lock.read_wait_group == nil {
			lock.read_wait_group = &sync.WaitGroup{}
			lock.read_wait_group.Add(1)
		}
		waitGroup = lock.read_wait_group
	} else {
		lock.read_count++
	}

	lock.lock.Unlock()

	if waitGroup != nil {
		// Wait for the write threads to finish
		waitGroup.Wait()
	}
}

// Ends a read operation
func (lock *ReadWriteLock) EndRead() {
	lock.lock.Lock()

	lock.read_count--

	if lock.write_wait && lock.read_count <= 0 {
		lock.write_wait = false
		lock.writing = true

		if lock.write_wait_group != nil {
			// Release write wait group
			lock.write_wait_group.Done()
			lock.write_wait_group = nil
		}
	}

	lock.lock.Unlock()
}
