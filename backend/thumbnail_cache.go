// Album thumbnail cache

package main

import "sync"

// Album thumbnail cache entry
type ThumbnailCacheEntry struct {
	is_pending bool // True if pending

	pending_wait_group *sync.WaitGroup // Group to wait for pending

	invalid bool // True if invalid (marked when deleting, but pending)

	has_thumbnail bool   // Has thumbnail?
	media         uint64 // First media of the album
	asset         uint64 // Thumbnail asset
}

// Cache for album thumbnails
type ThumbnailCache struct {
	mutex *sync.Mutex
	cache map[uint64]*ThumbnailCacheEntry
}

// Creates a blank cache
func makeThumbnailCache() *ThumbnailCache {
	res := ThumbnailCache{
		mutex: &sync.Mutex{},
		cache: make(map[uint64]*ThumbnailCacheEntry),
	}

	return &res
}

// Tries to get an entry. If not found, creates one, marks it as pending
func (c *ThumbnailCache) getEntryOrMarkAsPending(album uint64) (result_entry *ThumbnailCacheEntry, wait_group *sync.WaitGroup) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := c.cache[album]

	var wg *sync.WaitGroup = nil

	if entry == nil {
		wg = &sync.WaitGroup{}
		wg.Add(1)

		// Mark as pending
		c.cache[album] = &ThumbnailCacheEntry{
			is_pending:         true,
			pending_wait_group: wg,
			invalid:            false,
			has_thumbnail:      false,
			media:              0,
			asset:              0,
		}
	}

	return entry, wg
}

// Removes cache entry
func (c *ThumbnailCache) removeEntry(album uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, album)
}

// Removes cache entry, or marks as invalid if pending
func (c *ThumbnailCache) RemoveEntryOrMarkInvalid(album uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := c.cache[album]

	if entry == nil {
		return
	}

	if entry.is_pending {
		entry.invalid = true
		return
	}

	delete(c.cache, album)
}

// Resolves cache entry
func (c *ThumbnailCache) resolveEntry(album uint64, has_thumbnail bool, media uint64, asset uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := c.cache[album]

	if entry == nil {
		return
	}

	entry.is_pending = false

	entry.has_thumbnail = has_thumbnail
	entry.media = media
	entry.asset = asset

	if entry.invalid {
		delete(c.cache, album)
	}
}

// Gets the album thumbnail asset from cache of the vault data
func (c *ThumbnailCache) GetAlbumThumbnail(album uint64, key []byte) (has_thumbnail bool, media uint64, asset uint64) {
	entry, wg := c.getEntryOrMarkAsPending(album)

	if entry != nil {
		c.mutex.Lock()
		if entry.is_pending {
			entry_waiter := entry.pending_wait_group

			c.mutex.Unlock()

			entry_waiter.Wait()

			// We waited, now it's ready

			c.mutex.Lock()

			has_thumbnail_tmp := entry.has_thumbnail
			media_tmp := entry.media
			asset_tmp := entry.asset

			c.mutex.Unlock()

			return has_thumbnail_tmp, media_tmp, asset_tmp
		} else {
			has_thumbnail_tmp := entry.has_thumbnail
			media_tmp := entry.media
			asset_tmp := entry.asset

			c.mutex.Unlock()

			return has_thumbnail_tmp, media_tmp, asset_tmp
		}
	}

	// There is no entry
	// So it was marked as pending
	// We have to fetch the data from the vault

	defer wg.Done()

	albumsData, err := GetVault().albums.readData(key)

	if err != nil {
		c.removeEntry(album)
		return false, 0, 0
	}

	albumData := albumsData.Albums[album]

	if albumData == nil {
		c.removeEntry(album)
		return false, 0, 0
	}

	if len(albumData.List) == 0 {
		// No thumbnail
		c.resolveEntry(album, false, 0, 0)
		return false, 0, 0
	}

	firstMediaId := albumData.List[0]

	firstMedia := GetVault().media.AcquireMediaResource(firstMediaId)

	meta, err := firstMedia.ReadMetadata(key)

	GetVault().media.ReleaseMediaResource(firstMediaId)

	if err != nil {
		LogError(err)
	}

	if meta == nil {
		// Not found media
		c.removeEntry(album)
		return false, 0, 0
	}

	has_thumbnail = meta.ThumbnailReady
	thumbnail_asset := meta.ThumbnailAsset

	c.resolveEntry(album, has_thumbnail, firstMediaId, thumbnail_asset)
	return has_thumbnail, firstMediaId, thumbnail_asset
}
