// Vault scanner

package main

import (
	"sort"
)

const MAX_TAGS_FILTER_COUNT = 16

type TagFilterMode int

const (
	TagFilterAllOf  TagFilterMode = 0
	TagFilterAnyOf  TagFilterMode = 1
	TagFilterNoneOf TagFilterMode = 2
)

// Entry to store a tag id within its media count
type TagIndexMetadata struct {
	// Tag ID
	id uint64

	// Size of the index
	count int64

	// Index manager
	index *VaultMainIndex

	// Index file
	indexFile *IndexedListFile
}

// Vault scanner
type VaultScanner struct {
	// Tag filter mode
	tagFilterMode TagFilterMode

	// Indexes of tags
	tagIndexes []*TagIndexMetadata

	// Main index of the vault
	mainIndexFile *IndexedListFile

	// Index being scanned
	scanningIndex *IndexedListFile

	// Count pre-calc if the scanning index
	scanningIndexCount int64

	// True for reverse order
	reversed bool

	// Current index
	currentIndex int64
}

// Setups a vault scanner
// tagMode - Tag filtering mode
// reversed - True if reverse order
// tagNames - List of tag names
// continueRef - Reference to continue
// key - Vault decryption key
// Returns a reference to the created VaultScanner
// May return null if an error occurs, or if some tags are not found, meaning the list is empty
func NewVaultScanner(tagMode TagFilterMode, reversed bool, tagNames []string, continueRef int64, key []byte) (*VaultScanner, error) {
	tagList, err := GetVault().tags.ReadList(key)

	if err != nil {
		return nil, err
	}

	// Find tags

	if len(tagNames) > MAX_TAGS_FILTER_COUNT {
		tagNames = tagNames[:MAX_TAGS_FILTER_COUNT]
	}

	tags := make([]uint64, 0)

	for _, tagName := range tagNames {
		found, tagId := tagList.FindTag(tagName)

		if !found {
			if tagMode == TagFilterAllOf {
				// If we require all tags, but one does not exists, the list is empty
				return nil, nil
			}

			continue
		}

		tags = append(tags, tagId)
	}

	if tagMode == TagFilterAnyOf && len(tags) == 0 {
		// If we require at least one tag, but no tags are provided, the list is empty
		return nil, nil
	}

	// Create vault scanner

	vaultScanner := &VaultScanner{
		tagFilterMode: tagMode,
		reversed:      reversed,
	}

	// Load main index if necessary

	if tagMode == TagFilterNoneOf || (tagMode == TagFilterAllOf && len(tags) == 0) || (tagMode == TagFilterAnyOf && len(tags) > 1) {
		indexFile, err := GetVault().index.StartRead()

		if err != nil {
			return nil, err
		}

		vaultScanner.mainIndexFile = indexFile
		vaultScanner.scanningIndex = indexFile

		count, err := indexFile.Count()

		if err != nil {
			GetVault().index.EndRead(vaultScanner.mainIndexFile)
			return nil, err
		}

		vaultScanner.scanningIndexCount = count
	}

	// Load and sort tag indexes

	sortedTags := make([]*TagIndexMetadata, len(tags))

	for i := 0; i < len(tags); i++ {
		meta := &TagIndexMetadata{
			id: tags[i],
		}

		tagIndex, err := GetVault().tags.AcquireIndexFile(meta.id)

		if err != nil {
			// Release all acquired tags
			for j := i - 1; j >= 0; j-- {
				sortedTags[j].index.EndRead(sortedTags[j].indexFile)
				GetVault().tags.ReleaseIndexFile(sortedTags[j].id, false, key)
			}

			// Release main index if acquired
			if vaultScanner.mainIndexFile != nil {
				GetVault().index.EndRead(vaultScanner.mainIndexFile)
			}

			return nil, err
		}

		meta.index = tagIndex

		tagIndexFile, err := tagIndex.StartRead()

		if err != nil {
			// Release current tag
			GetVault().tags.ReleaseIndexFile(meta.id, false, key)

			// Release all acquired tags
			for j := i - 1; j >= 0; j-- {
				sortedTags[j].index.EndRead(sortedTags[j].indexFile)
				GetVault().tags.ReleaseIndexFile(sortedTags[j].id, false, key)
			}

			// Release main index if acquired
			if vaultScanner.mainIndexFile != nil {
				GetVault().index.EndRead(vaultScanner.mainIndexFile)
			}

			return nil, err
		}

		meta.indexFile = tagIndexFile

		count, err := tagIndexFile.Count()

		if err != nil {
			// Release current tag
			tagIndex.EndRead(tagIndexFile)
			GetVault().tags.ReleaseIndexFile(meta.id, false, key)

			// Release all acquired tags
			for j := i - 1; j >= 0; j-- {
				sortedTags[j].index.EndRead(sortedTags[j].indexFile)
				GetVault().tags.ReleaseIndexFile(sortedTags[j].id, false, key)
			}

			// Release main index if acquired
			if vaultScanner.mainIndexFile != nil {
				GetVault().index.EndRead(vaultScanner.mainIndexFile)
			}

			return nil, err
		}

		meta.count = count

		sortedTags[i] = meta
	}

	if tagMode == TagFilterAllOf {
		sort.Slice(sortedTags, func(i, j int) bool {
			return sortedTags[i].count < sortedTags[j].count
		})
	} else {
		sort.Slice(sortedTags, func(i, j int) bool {
			return sortedTags[i].count > sortedTags[j].count
		})
	}

	vaultScanner.tagIndexes = sortedTags

	if tagMode == TagFilterAllOf && len(sortedTags) > 0 {
		vaultScanner.scanningIndex = sortedTags[0].indexFile
		vaultScanner.scanningIndexCount = sortedTags[0].count
	} else if tagMode == TagFilterAnyOf && len(sortedTags) == 1 {
		vaultScanner.scanningIndex = sortedTags[len(sortedTags)-1].indexFile
		vaultScanner.scanningIndexCount = sortedTags[len(sortedTags)-1].count
	}

	// Position the scanner properly based on the continue value

	if vaultScanner.scanningIndex == nil {
		vaultScanner.Release(key)
		return nil, nil
	}

	if continueRef >= 0 {
		// Find continue value in the index file

		found, foundIndex, err := vaultScanner.scanningIndex.BinarySearchWithCountPreCalc(uint64(continueRef), vaultScanner.scanningIndexCount)

		if err != nil {
			vaultScanner.Release(key)
			return nil, err
		}

		if reversed {
			if found {
				vaultScanner.currentIndex = foundIndex - 1
			} else {
				vaultScanner.currentIndex = foundIndex
			}
		} else {
			vaultScanner.currentIndex = foundIndex + 1
		}
	} else if reversed {
		vaultScanner.currentIndex = vaultScanner.scanningIndexCount - 1
	} else {
		vaultScanner.currentIndex = 0
	}

	return vaultScanner, nil
}

// Releases all the resources acquired by this scanner
func (scanner *VaultScanner) Release(key []byte) {
	if scanner.mainIndexFile != nil {
		GetVault().index.EndRead(scanner.mainIndexFile)
	}

	for _, tie := range scanner.tagIndexes {
		tie.index.EndRead(tie.indexFile)
		GetVault().tags.ReleaseIndexFile(tie.id, false, key)
	}
}

// Finds the next value
// Returns:
// - found: true only if the scanner has a value to return, false means the scanner has reached the end
// - id: The id of the next media element
// - err: Error
func (scanner *VaultScanner) Next() (found bool, id uint64, err error) {
	for (!scanner.reversed && scanner.currentIndex < scanner.scanningIndexCount) || (scanner.reversed && scanner.currentIndex >= 0) {
		// Read value

		current, err := scanner.scanningIndex.ReadValue(scanner.currentIndex)

		if err != nil {
			return false, 0, err
		}

		// Increment index

		if scanner.reversed {
			scanner.currentIndex--
		} else {
			scanner.currentIndex++
		}

		// Filter

		switch scanner.tagFilterMode {
		case TagFilterAllOf:
			p, err := scanner.checkAllTags(current)

			if err != nil {
				return false, 0, err
			}

			if !p {
				continue
			}
		case TagFilterAnyOf:
			p, err := scanner.checkAnyTag(current)

			if err != nil {
				return false, 0, err
			}

			if !p {
				continue
			}
		case TagFilterNoneOf:
			p, err := scanner.checkNoneTag(current)

			if err != nil {
				return false, 0, err
			}

			if !p {
				continue
			}
		}

		// Return

		return true, current, nil
	}

	// End reached
	return false, 0, nil
}

// Checks all tags
// Returns true only if the media ID is present in all the indexed lists
func (scanner *VaultScanner) checkAllTags(id uint64) (passes bool, err error) {
	if len(scanner.tagIndexes) < 2 {
		return true, nil
	}
	for _, ie := range scanner.tagIndexes[1:] {
		found, _, err := ie.indexFile.BinarySearchWithCountPreCalc(id, ie.count)

		if err != nil {
			return false, err
		}

		if !found {
			return false, nil
		}
	}

	return true, nil
}

// Checks any tag
// Returns true only if the media ID is present in at least one of the indexed lists
func (scanner *VaultScanner) checkAnyTag(id uint64) (passes bool, err error) {
	for _, ie := range scanner.tagIndexes {
		found, _, err := ie.indexFile.BinarySearchWithCountPreCalc(id, ie.count)

		if err != nil {
			return false, err
		}

		if found {
			return true, nil
		}
	}

	return false, nil
}

// Checks none of the tags contain the media
// Returns true only if the media ID is not present in any of the indexed lists
func (scanner *VaultScanner) checkNoneTag(id uint64) (passes bool, err error) {
	for _, ie := range scanner.tagIndexes {
		found, _, err := ie.indexFile.BinarySearchWithCountPreCalc(id, ie.count)

		if err != nil {
			return false, err
		}

		if found {
			return false, nil
		}
	}

	return true, nil
}

// Gets current scanner progress
func (scanner *VaultScanner) GetProgress() (scanned int64, total int64) {
	if scanner.reversed {
		if scanner.currentIndex < 0 {
			return scanner.scanningIndexCount, scanner.scanningIndexCount
		}

		return scanner.scanningIndexCount - scanner.currentIndex - 1, scanner.scanningIndexCount
	} else {
		if scanner.currentIndex >= scanner.scanningIndexCount {
			return scanner.scanningIndexCount, scanner.scanningIndexCount
		}

		return scanner.currentIndex, scanner.scanningIndexCount
	}
}
