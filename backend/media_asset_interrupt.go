// Media asset read interrupt manager

package main

import (
	"os"
	"sync"
)

// Media asset key
type MediaAssetKey struct {
	MediaId, AssetId uint64
}

// State of a media asset
type MediaAssetReadInterruptState struct {
	// True if interrupted
	interrupted int

	// Notifiers to interrupt pending reads
	readerNotifiers [](chan struct{})
}

// Manager to control read interruption of assets
type MediaAssetReadInterruptManager struct {
	// Mutex to control access to the struct
	lock *sync.Mutex

	// Map of states
	states map[MediaAssetKey]*MediaAssetReadInterruptState
}

func makeMediaAssetReadInterruptManager() *MediaAssetReadInterruptManager {
	return &MediaAssetReadInterruptManager{
		lock:   &sync.Mutex{},
		states: make(map[MediaAssetKey]*MediaAssetReadInterruptState),
	}
}

// Indicates a reader for an assert
// Returns true only if allowed
// If allowed, it also returns a channel to receive interrupt signals
// Important: Must call EndRead after the read is done
func (manager *MediaAssetReadInterruptManager) StartRead(mediaId uint64, assetId uint64) (bool, chan struct{}) {
	key := MediaAssetKey{
		MediaId: mediaId,
		AssetId: assetId,
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	state := manager.states[key]

	if state != nil {
		if state.interrupted > 0 {
			return false, nil
		}

		interruptChan := make(chan struct{})

		state.readerNotifiers = append(state.readerNotifiers, interruptChan)

		return true, interruptChan
	} else {
		interruptChan := make(chan struct{})

		manager.states[key] = &MediaAssetReadInterruptState{
			interrupted:     0,
			readerNotifiers: [](chan struct{}){interruptChan},
		}

		return true, interruptChan
	}
}

// Indicates the end of a read operation
func (manager *MediaAssetReadInterruptManager) EndRead(mediaId uint64, assetId uint64, notifier chan struct{}) {
	key := MediaAssetKey{
		MediaId: mediaId,
		AssetId: assetId,
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	state := manager.states[key]

	if state == nil {
		return
	}

	for i, n := range state.readerNotifiers {
		if n == notifier {
			state.readerNotifiers = append(state.readerNotifiers[:i], state.readerNotifiers[i+1:]...)
			break
		}
	}

	if len(state.readerNotifiers) == 0 && state.interrupted <= 0 {
		delete(manager.states, key)
	}
}

// Interrupts any read operations for a specific asset
// Important: The caller must call FinalizeInterrupt after it is done with the write operation
func (manager *MediaAssetReadInterruptManager) Interrupt(mediaId uint64, assetId uint64) {
	key := MediaAssetKey{
		MediaId: mediaId,
		AssetId: assetId,
	}

	var toNotify [](chan struct{}) = nil

	manager.lock.Lock()

	state := manager.states[key]

	if state == nil {
		manager.states[key] = &MediaAssetReadInterruptState{
			interrupted:     1,
			readerNotifiers: nil,
		}
	} else {
		state.interrupted++

		toNotify = state.readerNotifiers

		state.readerNotifiers = nil
	}

	manager.lock.Unlock()

	for _, n := range toNotify {
		n <- struct{}{}
	}
}

// Finalizes the read interruption for an asset
func (manager *MediaAssetReadInterruptManager) FinalizeInterrupt(mediaId uint64, assetId uint64) {
	key := MediaAssetKey{
		MediaId: mediaId,
		AssetId: assetId,
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	state := manager.states[key]

	if state == nil {
		return
	}

	state.interrupted--

	if len(state.readerNotifiers) == 0 && state.interrupted <= 0 {
		delete(manager.states, key)
	}
}

// Clears an old media asset
// Example use: Replacing the olf thumbnail for a bew one
func ClearOldMediaAsset(media *MediaAsset, oldAssetId uint64, oldAssetType string) {
	success, asset_path, asset_lock := media.AcquireAsset(oldAssetId, oldAssetType)

	if success {
		GetVault().media.readInterruptManager.Interrupt(media.id, oldAssetId)

		asset_lock.RequestWrite()
		asset_lock.StartWrite()

		err := os.Remove(asset_path)

		if err != nil {
			LogWarning("Could not delete the asset file '" + asset_path + "'. Cause: " + err.Error())
		}

		asset_lock.EndWrite()

		GetVault().media.readInterruptManager.FinalizeInterrupt(media.id, oldAssetId)

		media.ReleaseAsset(oldAssetId)
	}
}
