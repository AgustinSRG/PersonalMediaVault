// Tests for multi file pack

package main

import (
	"os"
	"path"
	"testing"
)

func TestMultiFilePack(t *testing.T) {
	test_path_base := "./temp"

	err := os.MkdirAll(test_path_base, FOLDER_PERMISSION)

	if err != nil {
		t.Error(err)
	}

	test_file := path.Join(test_path_base, "test_multi_file_pack.pmv")

	fileContents1 := "File contents 1 (AABB)"
	fileContents2 := "File contents 2"
	fileContents3 := "File contents 3 (AAAAAAAA)"
	fileContents4 := "File contents 4 (AABBCC)"
	fileContents5 := "File contents 5 (AABBCCDDEEFFGG)"

	// Write file

	file, err := CreateMultiFilePackWriteStream(test_file)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = file.Initialize(5)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = file.PutFile([]byte(fileContents1))

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = file.PutFile([]byte(fileContents2))

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = file.PutFile([]byte(fileContents3))

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = file.PutFile([]byte(fileContents4))

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = file.PutFile([]byte(fileContents5))

	if err != nil {
		t.Error(err)
		panic(err)
	}

	file.Close()

	// Read file

	rf, err := CreateMultiFilePackReadStream(test_file)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if rf.file_count != 5 {
		t.Errorf("Expected file_count = %d, but got %d", 5, rf.file_count)
	}

	// Check files

	b, err := rf.GetFile(0)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if string(b) != fileContents1 {
		t.Errorf("Expected GetFile(0) = (%s), but got (%s)", fileContents1, string(b))
	}

	b, err = rf.GetFile(1)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if string(b) != fileContents2 {
		t.Errorf("Expected GetFile(1) = (%s), but got (%s)", fileContents2, string(b))
	}

	b, err = rf.GetFile(2)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if string(b) != fileContents3 {
		t.Errorf("Expected GetFile(2) = (%s), but got (%s)", fileContents3, string(b))
	}

	b, err = rf.GetFile(3)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if string(b) != fileContents4 {
		t.Errorf("Expected GetFile(3) = (%s), but got (%s)", fileContents4, string(b))
	}

	b, err = rf.GetFile(4)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if string(b) != fileContents5 {
		t.Errorf("Expected GetFile(4) = (%s), but got (%s)", fileContents5, string(b))
	}

	rf.Close()

	// Remove temp file

	os.Remove(test_file)
}
