// Indexed list of identifiers, for the main index and tags
// This kind of file is a sorted list of Media IDs
// Header:
//   - List size (uint64) (Big endian) (8 bytes)
// List (always sorted):
//   - Each item is an uint64 (Big endian) (8 bytes)

package main

import (
	"encoding/binary"
	"math/rand"
	"os"
)

// Indexed list file
type IndexedListFile struct {
	f *os.File // File descriptor
}

// Opens index file for writing
// file - Path to the file to open
func OpenIndexedListForWriting(file string) (*IndexedListFile, error) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, FILE_PERMISSION)

	if err != nil {
		return nil, err
	}

	i := IndexedListFile{
		f: f,
	}

	return &i, nil
}

// Opens index file for reading
// file - Path to the file to open
func OpenIndexedListForReading(file string) (*IndexedListFile, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, FILE_PERMISSION)

	if err != nil {
		return nil, err
	}

	i := IndexedListFile{
		f: f,
	}

	return &i, nil
}

// Closes the file
func (file *IndexedListFile) Close() {
	_ = file.f.Close()
}

// Returns the number of items in the index
func (file *IndexedListFile) Count() (int64, error) {
	// Rewind to the start of the file
	_, err := file.f.Seek(0, 0)

	if err != nil {
		return 0, err
	}

	b := make([]byte, 8)

	_, err = file.f.Read(b)

	if err != nil {
		return 0, err
	}

	return int64(binary.BigEndian.Uint64(b)), nil
}

// Initializes the index. Call once when the index file does not exists
func (file *IndexedListFile) Initialize() error {
	// Set the size of the file
	err := file.f.Truncate(8)
	if err != nil {
		return err
	}

	// Rewind to the start of the file
	_, err = file.f.Seek(0, 0)

	if err != nil {
		return err
	}

	// Write size
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, 0)
	_, err = file.f.Write(b)

	if err != nil {
		return err
	}

	return nil
}

// Reads a value given an index
// index - Index in the sorted list
func (file *IndexedListFile) ReadValue(index int64) (uint64, error) {
	_, err := file.f.Seek(8+(index*8), 0)

	if err != nil {
		return 0, err
	}

	b := make([]byte, 8)

	_, err = file.f.Read(b)

	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(b), nil
}

// Writes a value
// index - Index in the sorted list
// value - Value to write
func (file *IndexedListFile) WriteValue(index int64, value uint64) error {
	_, err := file.f.Seek(8+(index*8), 0)

	if err != nil {
		return err
	}

	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, value)

	_, err = file.f.Write(b)

	if err != nil {
		return err
	}

	return nil
}

// Searches in the file for a value
// val - Value to search in the list
// Returns:
//
//	1 - A boolean value = true if the exact value was found
//	2 - The closest index to that value
func (file *IndexedListFile) BinarySearch(val uint64) (bool, int64, error) {
	count, err := file.Count()

	if err != nil {
		return false, 0, err
	}

	if count == 0 {
		return false, 0, nil
	}

	var low int64
	var high int64
	var mVal uint64

	low = 0
	high = int64(count - 1)

	for low <= high {
		m := (low + high) / 2

		mVal, err = file.ReadValue(m)

		if err != nil {
			return false, 0, err
		}

		if mVal < val {
			low = m + 1
		} else if mVal > val {
			high = m - 1
		} else {
			low = m
			high = m - 1
		}
	}

	return mVal == val, low, nil
}

// Searches in the file for a value
// This method requires you to provide the list length.
// A wrong value will result in undefined behavior
// val - Value to search in the list
// count - List size
// Returns:
//
//	1 - A boolean value = true if the exact value was found
//	2 - The closest index to that value
func (file *IndexedListFile) BinarySearchWithCountPreCalc(val uint64, count int64) (bool, int64, error) {
	if count == 0 {
		return false, 0, nil
	}

	var low int64
	var high int64
	var mVal uint64
	var err error

	low = 0
	high = count - 1

	for low <= high {
		m := (low + high) / 2

		mVal, err = file.ReadValue(m)

		if err != nil {
			return false, 0, err
		}

		if mVal < val {
			low = m + 1
		} else if mVal > val {
			high = m - 1
		} else {
			low = m
			high = m - 1
		}
	}

	return mVal == val, low, nil
}

// Adds a value
// val - Value to add
// Returns
//
//	1 - true if it was added, false if it was already in the file
//	2 - the index where the value was added
func (file *IndexedListFile) AddValue(val uint64) (bool, int64, error) {
	count, err := file.Count()

	if err != nil {
		return false, 0, err
	}

	found, index, err := file.BinarySearchWithCountPreCalc(val, count)

	if err != nil {
		return false, 0, err
	}

	if found {
		return false, index, nil
	}

	// Resize file to make space for the new value
	err = file.f.Truncate(8 + 8*(count+1))
	if err != nil {
		return false, 0, err
	}

	// Move down 1 all items below
	tempVal := val
	for i := index; i < count+1; i++ {
		rVal, err := file.ReadValue(i)
		if err != nil {
			return false, 0, err
		}

		err = file.WriteValue(i, tempVal)
		if err != nil {
			return false, 0, err
		}

		tempVal = rVal
	}

	// Write the count value

	// Rewind to the start of the file
	_, err = file.f.Seek(0, 0)

	if err != nil {
		return false, 0, err
	}

	// Write size
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(count+1))
	_, err = file.f.Write(b)

	if err != nil {
		return false, 0, err
	}

	return true, index, nil
}

// Removes a value
// val - Value to remove
// Returns
//
//	1 - true if it was removed, false if it was not present in the index
//	2 - The new count value
func (file *IndexedListFile) RemoveValue(val uint64) (bool, int64, error) {
	count, err := file.Count()

	if err != nil {
		return false, 0, err
	}

	found, index, err := file.BinarySearchWithCountPreCalc(val, count)

	if err != nil {
		return false, 0, err
	}

	if !found {
		return false, count, nil
	}

	new_count, err := file.RemoveIndex(index, count)

	if err != nil {
		return false, 0, err
	}

	return true, new_count, nil
}

// Removes a value given the index and the count
// index - Index to remove
// count - List size
// Returns the new count value
func (file *IndexedListFile) RemoveIndex(index int64, count int64) (int64, error) {
	// Move instances 1 above
	lastIndex := int64(count - 1)
	for i := index; i < lastIndex; i++ {
		val, err := file.ReadValue(i + 1)
		if err != nil {
			return 0, err
		}

		err = file.WriteValue(i, val)

		if err != nil {
			return 0, err
		}
	}

	// Change file size
	err := file.f.Truncate(int64(8 + 8*(count-1)))
	if err != nil {
		return 0, err
	}

	// Write the count value

	// Rewind to the start of the file
	_, err = file.f.Seek(0, 0)

	if err != nil {
		return 0, err
	}

	// Write size
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(count-1))
	_, err = file.f.Write(b)

	if err != nil {
		return 0, err
	}

	return count - 1, err
}

// List values inside the index in order
// skip - Number of items to skip
// limit - Max number of items to return
func (file *IndexedListFile) ListValues(skip int64, limit int64) ([]uint64, error) {
	count, err := file.Count()

	if err != nil {
		return nil, err
	}

	if count <= 0 || limit <= 0 || skip >= count {
		return make([]uint64, 0), nil
	}

	resultSize := count - skip

	if resultSize > limit {
		resultSize = limit
	}

	result := make([]uint64, resultSize)

	for i := int64(0); i < resultSize; i++ {
		val, err := file.ReadValue(skip + i)
		if err != nil {
			return nil, err
		}

		result[i] = val
	}

	return result, nil
}

// List values inside the index in reverse order
// skip - Number of items to skip
// limit - Max number of items to return
func (file *IndexedListFile) ListValuesReverse(skip int64, limit int64) ([]uint64, error) {
	count, err := file.Count()

	if err != nil {
		return nil, err
	}

	if count <= 0 || limit <= 0 || skip >= count {
		return make([]uint64, 0), nil
	}

	resultSize := count - skip

	if resultSize > limit {
		resultSize = limit
	}

	result := make([]uint64, resultSize)

	for i := int64(0); i < resultSize; i++ {
		val, err := file.ReadValue((count - 1) - skip - i)
		if err != nil {
			return nil, err
		}

		result[i] = val
	}

	return result, nil
}

// Returns a random list of values from the list
// seed - Seed for the P-RNG
// limit - Max number of items to return
func (file *IndexedListFile) RandomValues(seed int64, limit int64) ([]uint64, error) {
	count, err := file.Count()

	if err != nil {
		return nil, err
	}

	if count <= 0 || limit <= 0 {
		return make([]uint64, 0), nil
	}

	pseudoRandom := rand.New(rand.NewSource(seed))

	result := make([]uint64, limit)

	index_set := make(map[int64]struct{})

	resultCount := int64(0)

	for i := int64(0); i < limit; i++ {
		if resultCount >= count {
			break
		}

		index := pseudoRandom.Int63()

		if index < 0 {
			index = 0
		}

		index = index % count

		_, repeated := index_set[index]

		if repeated {
			repeatedIndex := index
			index++

			if index >= count {
				index = 0
			}

			for index != repeatedIndex && repeated {
				_, repeated = index_set[index]

				if repeated {
					index++

					if index >= count {
						index = 0
					}
				}
			}

			if repeated {
				break
			}
		}

		index_set[index] = struct{}{}

		val, err := file.ReadValue(index)

		if err != nil {
			return nil, err
		}

		result[i] = val
		resultCount++
	}

	return result[0:resultCount], nil
}
