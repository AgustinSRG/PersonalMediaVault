// Media preview cache

package main

import (
	"fmt"
	"sync"
)

// Media preview cache entry
type MediaPreviewCacheEntry struct {
	is_pending bool // True if pending

	pending_wait_group *sync.WaitGroup // Group to wait for pending

	invalid bool // True if invalid (marked when deleting, but pending)

	prev *MediaPreviewCacheEntry // Previous
	next *MediaPreviewCacheEntry // Next

	media_id uint64            // Media ID
	data     *MediaListAPIItem // Cache actual data
}

// Cache for media previews
type MediaPreviewCache struct {
	mutex *sync.Mutex

	cache map[uint64]*MediaPreviewCacheEntry

	first *MediaPreviewCacheEntry
	last  *MediaPreviewCacheEntry

	max_size int
}

// Creates a blank cache
// max_size Max number of entries
func makeMediaPreviewCache(max_size int) *MediaPreviewCache {
	res := MediaPreviewCache{
		mutex:    &sync.Mutex{},
		cache:    make(map[uint64]*MediaPreviewCacheEntry),
		first:    nil,
		last:     nil,
		max_size: max_size,
	}

	return &res
}

// Tries to get an entry. If not found, creates one, marks it as pending
func (c *MediaPreviewCache) getEntryOrMarkAsPending(media_id uint64) (result_entry *MediaPreviewCacheEntry, wait_group *sync.WaitGroup) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := c.cache[media_id]

	var wg *sync.WaitGroup = nil

	if entry == nil {
		// Remove the last element if size has been reached
		if len(c.cache) >= c.max_size {
			if c.last != nil {
				entryToRemove := c.last
				c.removeArbitraryEntry(entryToRemove)
				delete(c.cache, entryToRemove.media_id)
			}
		}

		// Add the new entry at the top

		wg = &sync.WaitGroup{}
		wg.Add(1)

		// Mark as pending
		newEntry := &MediaPreviewCacheEntry{
			is_pending:         true,
			pending_wait_group: wg,
			invalid:            false,
			prev:               nil,
			next:               nil,
			media_id:           media_id,
			data: &MediaListAPIItem{
				Id:          media_id,
				Type:        MediaTypeDeleted,
				Title:       "",
				Description: "",
				Thumbnail:   "",
				Duration:    0,
				Tags:        make([]uint64, 0),
			},
		}

		// Add to the cache

		if c.max_size > 0 {
			c.cache[media_id] = newEntry
			c.insertOnTop(newEntry)
		}
	} else {
		// Entry exists, but we used the entry, so move it to the first
		c.removeArbitraryEntry(entry)
		c.insertOnTop(entry)
	}

	return entry, wg
}

// Internal method (thread unsafe)
// Removes arbitrary entry from the linked list
func (c *MediaPreviewCache) removeArbitraryEntry(entry *MediaPreviewCacheEntry) {
	next := entry.next
	prev := entry.prev

	if next != nil {
		next.prev = prev
	} else {
		// Is the last element
		c.last = prev
	}

	if prev != nil {
		prev.next = next
	} else {
		// Is the first element
		c.first = next
	}
}

// Internal method (thread unsafe)
// Inserts arbitrary entry in the top of the list
func (c *MediaPreviewCache) insertOnTop(entry *MediaPreviewCacheEntry) {
	if c.first == nil {
		// Empty
		entry.next = nil
		entry.prev = nil
		c.first = entry
		c.last = entry
	} else {
		entry.next = c.first
		entry.prev = nil
		c.first.prev = entry
		c.first = entry
	}
}

// Removes cache entry
func (c *MediaPreviewCache) removeEntry(media_id uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := c.cache[media_id]

	if entry != nil {
		c.removeArbitraryEntry(entry)
		delete(c.cache, media_id)
	}
}

// Removes cache entry, or marks as invalid if pending
func (c *MediaPreviewCache) RemoveEntryOrMarkInvalid(media_id uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := c.cache[media_id]

	if entry == nil {
		return
	}

	if entry.is_pending {
		entry.invalid = true
		return
	}

	c.removeArbitraryEntry(entry)
	delete(c.cache, media_id)
}

// Resolves cache entry
func (c *MediaPreviewCache) resolveEntry(media_id uint64, data *MediaListAPIItem) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := c.cache[media_id]

	if entry == nil {
		return
	}

	entry.is_pending = false
	entry.data = data

	if entry.invalid {
		c.removeArbitraryEntry(entry)
		delete(c.cache, media_id)
	}
}

// Gets media preview from cache of from the vault data
func (c *MediaPreviewCache) GetMediaPreview(media_id uint64, key []byte) *MediaListAPIItem {
	entry, wg := c.getEntryOrMarkAsPending(media_id)

	if entry != nil {
		c.mutex.Lock()
		if entry.is_pending {
			entry_waiter := entry.pending_wait_group

			c.mutex.Unlock()

			entry_waiter.Wait()

			// We waited, now it's ready

			c.mutex.Lock()

			data_tmp := entry.data

			c.mutex.Unlock()

			return data_tmp
		} else {
			data_tmp := entry.data

			c.mutex.Unlock()

			return data_tmp
		}
	}

	// There is no entry
	// So it was marked as pending
	// We have to fetch the data from the vault

	defer wg.Done()

	result := MediaListAPIItem{
		Id:          media_id,
		Type:        MediaTypeDeleted,
		Title:       "",
		Description: "",
		Thumbnail:   "",
		Duration:    0,
		Tags:        make([]uint64, 0),
	}

	media := GetVault().media.AcquireMediaResource(media_id)

	meta, err := media.ReadMetadata(key)

	if err != nil {
		LogError(err)
	}

	if meta == nil {
		c.removeEntry(media_id)
		return &result
	}

	result.Type = meta.Type

	result.Title = meta.Title
	result.Description = meta.Description

	if meta.ThumbnailReady {
		result.Thumbnail = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.ThumbnailAsset) + "/thumbnail.jpg" + "?fp=" + GetVault().credentials.GetFingerprint()
	} else {
		result.Thumbnail = ""
	}

	result.Duration = meta.MediaDuration

	result.Tags = meta.Tags

	GetVault().media.ReleaseMediaResource(media_id)

	c.resolveEntry(media_id, &result)

	return &result
}
