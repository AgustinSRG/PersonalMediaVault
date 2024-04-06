// Tests for RW-Lock

package main

import (
	"sync"
	"testing"
	"time"
)

func threadReader(wg *sync.WaitGroup, lock *ReadWriteLock, msWaitBefore int, msWaitRead int) {
	time.Sleep(time.Duration(msWaitBefore) * time.Millisecond)

	lock.StartRead()

	time.Sleep(time.Duration(msWaitRead) * time.Millisecond)

	lock.EndRead()

	wg.Done()
}

func threadWriter(wg *sync.WaitGroup, lock *ReadWriteLock, msWaitBefore int, msWaitWrite int, msWaitSave int) {
	time.Sleep(time.Duration(msWaitBefore) * time.Millisecond)

	lock.RequestWrite()

	time.Sleep(time.Duration(msWaitWrite) * time.Millisecond)

	lock.StartWrite()

	time.Sleep(time.Duration(msWaitSave) * time.Millisecond)

	lock.EndWrite()

	wg.Done()
}

func TestReadWriteLock(t *testing.T) {
	var wg sync.WaitGroup
	lock := CreateReadWriteLock()

	wg.Add(5)

	go threadReader(&wg, lock, 10, 50)
	go threadReader(&wg, lock, 10, 50)
	go threadReader(&wg, lock, 10, 50)

	go threadWriter(&wg, lock, 20, 20, 50)
	go threadWriter(&wg, lock, 20, 20, 50)

	wg.Wait()
}
