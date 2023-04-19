// Tests for indexed lists

package main

import (
	"os"
	"path"
	"testing"
)

func TestIndexedList(t *testing.T) {
	err := InitializeTestVault()

	if err != nil {
		t.Error(err)
		return
	}

	test_path_base := path.Join(temp_files_path, "unit")

	err = os.MkdirAll(test_path_base, FOLDER_PERMISSION)

	if err != nil {
		t.Error(err)
	}

	test_file := path.Join(test_path_base, "test_indexed_list.index")

	// Create a list
	f, err := OpenIndexedListForWriting(test_file)

	if err != nil {
		t.Error(err)
	}

	// Initialize

	err = f.Initialize()

	if err != nil {
		t.Error(err)
	}

	// Add first value

	success, index, err := f.AddValue(5)

	if err != nil {
		t.Error(err)
	}

	if !success {
		t.Errorf("Add value was not successful")
	}

	if index != 0 {
		t.Errorf("Expected index %d, but got %d", 0, index)
	}

	// Add value larger

	success, index, err = f.AddValue(6)

	if err != nil {
		t.Error(err)
	}

	if !success {
		t.Errorf("Add value was not successful")
	}

	if index != 1 {
		t.Errorf("Expected index %d, but got %d", 1, index)
	}

	// Add value lower

	success, index, err = f.AddValue(2)

	if err != nil {
		t.Error(err)
	}

	if !success {
		t.Errorf("Add value was not successful")
	}

	if index != 0 {
		t.Errorf("Expected index %d, but got %d", 0, index)
	}

	// Try adding duplicate value

	success, _, err = f.AddValue(2)

	if err != nil {
		t.Error(err)
	}

	if success {
		t.Errorf("Add value was successful")
	}

	// Count

	c, err := f.Count()

	if err != nil {
		t.Error(err)
	}

	if c != 3 {
		t.Errorf("Expected count = %d, but got %d", 3, c)
	}

	// Listing

	l, err := f.ListValues(0, 10)

	if err != nil {
		t.Error(err)
	}

	if len(l) != 3 {
		t.Errorf("Expected len(l) = %d, but got %d", 3, len(l))
	}

	if l[0] != 2 {
		t.Errorf("Expected l[0] = %d, but got %d", 2, l[0])
	}

	if l[1] != 5 {
		t.Errorf("Expected l[1] = %d, but got %d", 5, l[1])
	}

	if l[2] != 6 {
		t.Errorf("Expected l[2] = %d, but got %d", 6, l[2])
	}

	f.Close()

	// Open again for reading

	f, err = OpenIndexedListForReading(test_file)

	if err != nil {
		t.Error(err)
	}

	// Count

	c, err = f.Count()

	if err != nil {
		t.Error(err)
	}

	if c != 3 {
		t.Errorf("Expected count = %d, but got %d", 3, c)
	}

	// Listing

	l, err = f.ListValues(0, 10)

	if err != nil {
		t.Error(err)
	}

	if len(l) != 3 {
		t.Errorf("Expected len(l) = %d, but got %d", 3, len(l))
	}

	if l[0] != 2 {
		t.Errorf("Expected l[0] = %d, but got %d", 2, l[0])
	}

	if l[1] != 5 {
		t.Errorf("Expected l[1] = %d, but got %d", 5, l[1])
	}

	if l[2] != 6 {
		t.Errorf("Expected l[2] = %d, but got %d", 6, l[2])
	}

	// Search

	found, index, err := f.BinarySearch(6)

	if err != nil {
		t.Error(err)
	}

	if !found {
		t.Errorf("Value was not found")
	}

	if index != 2 {
		t.Errorf("Expected index = %d, but got %d", 2, index)
	}

	// Search (not found)

	found, _, err = f.BinarySearch(4)

	if err != nil {
		t.Error(err)
	}

	if found {
		t.Errorf("Value was found")
	}

	f.Close()

	os.Remove(test_file)
}
