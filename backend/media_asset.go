// Media asset

package main

type MediaAsset struct {
	id uint64

	path string
	lock *ReadWriteLock

	use_count int32
}
