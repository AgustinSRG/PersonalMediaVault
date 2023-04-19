// Tests for block-encrypted files

package main

import (
	"crypto/rand"
	"os"
	"path"
	"testing"
)

func TestFileBlockEncrypt(t *testing.T) {
	test_path_base := "./temp"

	err := os.MkdirAll(test_path_base, FOLDER_PERMISSION)

	if err != nil {
		t.Error(err)
		return
	}

	test_file := path.Join(test_path_base, "test_block_file_enc.pmv")
	size := int64(48 * 1024 * 1024)
	key := make([]byte, 32)
	rand.Read(key) //nolint:errcheck

	// Write file
	ws, err := CreateFileBlockEncryptWriteStream(test_file)

	if err != nil {
		t.Error(err)
		return
	}

	err = ws.Initialize(size, key)

	if err != nil {
		t.Error(err)
		return
	}

	buf := make([]byte, 269*1024)

	for j := 0; j < len(buf); j++ {
		buf[j] = 'A'
	}

	for i := 0; i < 48; i++ {
		err = ws.Write(buf)
		if err != nil {
			t.Error(err)
			return
		}
	}

	err = ws.Close()
	if err != nil {
		t.Error(err)
		return
	}

	// Read file

	rs, err := CreateFileBlockEncryptReadStream(test_file, key)

	if err != nil {
		t.Error(err)
		return
	}

	if rs.file_size != size {
		t.Errorf("Expected file_size = (%d), but got (%d)", size, rs.file_size)
	}

	if rs.block_size != FILE_ENC_BLOCK_SIZE_BYTES {
		t.Errorf("Expected block_size = (%d), but got (%d)", FILE_ENC_BLOCK_SIZE_BYTES, rs.block_size)
	}

	if rs.block_count != ws.block_count {
		t.Errorf("Expected block_count = (%d), but got (%d)", ws.block_count, rs.block_count)
	}

	buf2 := make([]byte, 269*1024)

	for i := 0; i < 48; i++ {
		n, err := rs.Read(buf2)

		if err != nil {
			t.Error(err)
			return
		}

		if n != len(buf) {
			t.Errorf("Expected n = (%d), but got (%d)", len(buf), n)
		}

		for j := 0; j < len(buf2); j++ {
			if buf2[j] != 'A' {
				t.Errorf("Expected buf[%d] = (A), but got (%c)", j, buf2[j])
			}
		}
	}

	rs.Close()

	// Remove temp file

	os.Remove(test_file)
}
