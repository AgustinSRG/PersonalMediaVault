// Tool to encrypt large files in blocks

package main

import (
	"encoding/binary"
	"errors"
	"os"
)

//////////////////////////
//     WRITE STREAM    //
/////////////////////////

const FILE_ENC_BLOCK_SIZE_BYTES = 5 * 1024 * 1024

type FileBlockEncryptWriteStream struct {
	f                   *os.File
	file_size           int64
	key                 []byte
	block_count         int64
	current_write_index int64
	current_write_pt    int64
	buf                 []byte
}

func CreateFileBlockEncryptWriteStream(file string) (*FileBlockEncryptWriteStream, error) {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FILE_PERMISSION)

	if err != nil {
		return nil, err
	}

	i := FileBlockEncryptWriteStream{
		f: f,
	}

	return &i, nil
}

func (file *FileBlockEncryptWriteStream) Initialize(file_size int64, key []byte) error {
	blockCount := file_size / FILE_ENC_BLOCK_SIZE_BYTES

	if file_size%FILE_ENC_BLOCK_SIZE_BYTES != 0 {
		blockCount++
	}

	file.file_size = file_size
	file.block_count = blockCount
	file.key = key

	// Set the size of the file
	err := file.f.Truncate(16 + (16)*blockCount)
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

	binary.BigEndian.PutUint64(b, uint64(file_size))
	_, err = file.f.Write(b)
	if err != nil {
		return err
	}

	// Write block size
	binary.BigEndian.PutUint64(b, uint64(FILE_ENC_BLOCK_SIZE_BYTES))
	_, err = file.f.Write(b)
	if err != nil {
		return err
	}

	// Write default block values
	binary.BigEndian.PutUint64(b, 0)
	for i := int64(0); i < blockCount; i++ {
		_, err = file.f.Write(b)
		if err != nil {
			return err
		}

		_, err = file.f.Write(b)
		if err != nil {
			return err
		}
	}

	file.current_write_index = 0
	file.current_write_pt = 16 + (16)*blockCount
	file.buf = make([]byte, 0)

	return nil
}

func (file *FileBlockEncryptWriteStream) Write(data []byte) error {
	if file.current_write_index >= file.block_count {
		return errors.New("Exceeded file size limit")
	}

	file.buf = append(file.buf, data...)

	for len(file.buf) >= FILE_ENC_BLOCK_SIZE_BYTES {
		if file.current_write_index >= file.block_count {
			return errors.New("Exceeded file size limit")
		}

		blockData := file.buf[:FILE_ENC_BLOCK_SIZE_BYTES]
		file.buf = file.buf[FILE_ENC_BLOCK_SIZE_BYTES:]

		content, err := encryptFileContents(blockData, AES256_ZIP, file.key)

		if err != nil {
			return err
		}

		// Save metadata
		_, err = file.f.Seek(16+file.current_write_index*16, 0)

		if err != nil {
			return err
		}

		b := make([]byte, 8)

		// Write start pointer
		binary.BigEndian.PutUint64(b, uint64(file.current_write_pt))
		_, err = file.f.Write(b)
		if err != nil {
			return err
		}

		// Write length
		binary.BigEndian.PutUint64(b, uint64(len(content)))
		_, err = file.f.Write(b)
		if err != nil {
			return err
		}

		// Write data

		_, err = file.f.Seek(file.current_write_pt, 0)

		if err != nil {
			return err
		}

		_, err = file.f.Write(content)
		if err != nil {
			return err
		}

		file.current_write_index++
		file.current_write_pt += int64(len(content))
	}

	return nil
}

func (file *FileBlockEncryptWriteStream) Close() error {
	if len(file.buf) > 0 {
		if file.current_write_index >= file.block_count {
			return errors.New("Exceeded file size limit")
		}

		content, err := encryptFileContents(file.buf, AES256_ZIP, file.key)

		if err != nil {
			return err
		}

		// Save metadata
		_, err = file.f.Seek(16+file.current_write_index*16, 0)

		if err != nil {
			return err
		}

		b := make([]byte, 8)

		// Write start pointer
		binary.BigEndian.PutUint64(b, uint64(file.current_write_pt))
		_, err = file.f.Write(b)
		if err != nil {
			return err
		}

		// Write length
		binary.BigEndian.PutUint64(b, uint64(len(content)))
		_, err = file.f.Write(b)
		if err != nil {
			return err
		}

		// Write data

		_, err = file.f.Seek(file.current_write_pt, 0)

		if err != nil {
			return err
		}

		_, err = file.f.Write(content)
		if err != nil {
			return err
		}

		file.current_write_index++
		file.current_write_pt += int64(len(content))
	}

	file.f.Close()

	return nil
}

//////////////////////////
//     READ STREAM     //
/////////////////////////

type FileBlockEncryptReadStream struct {
	f          *os.File
	file_size  int64
	block_size int64

	key         []byte
	block_count int64

	cur_pos int64

	cur_block      int64
	cur_block_data []byte
}

func CreateFileBlockEncryptReadStream(file string, key []byte) (*FileBlockEncryptReadStream, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, FILE_PERMISSION)

	if err != nil {
		return nil, err
	}

	i := FileBlockEncryptReadStream{
		f: f,
	}

	b := make([]byte, 8)

	// Read original file size

	_, err = f.Read(b)

	if err != nil {
		f.Close()
		return nil, err
	}

	i.file_size = int64(binary.BigEndian.Uint64(b))

	// Read block size

	_, err = f.Read(b)

	if err != nil {
		f.Close()
		return nil, err
	}

	i.block_size = int64(binary.BigEndian.Uint64(b))

	i.key = key

	i.block_count = i.file_size / i.block_size

	if i.file_size%i.block_size != 0 {
		i.block_count++
	}

	i.cur_block = -1
	i.cur_pos = 0

	return &i, nil
}

func (file *FileBlockEncryptReadStream) FetchBlock(block_num int64) error {
	if block_num < 0 || block_num >= file.block_count {
		return errors.New("Block index out of bounds")
	}

	// Read block metadata

	_, err := file.f.Seek(16+block_num*16, 0)

	if err != nil {
		return err
	}

	ptBytes := make([]byte, 8)
	lenBytes := make([]byte, 8)

	_, err = file.f.Read(ptBytes)

	if err != nil {
		return err
	}

	_, err = file.f.Read(lenBytes)

	if err != nil {
		return err
	}

	pt := int64(binary.BigEndian.Uint64(ptBytes))

	_, err = file.f.Seek(pt, 0)

	if err != nil {
		return err
	}

	l := int64(binary.BigEndian.Uint64(lenBytes))

	// Read encrypted data

	data := make([]byte, l)

	_, err = file.f.Read(data)

	if err != nil {
		return err
	}

	// Descrypt block data

	data, err = decryptFileContents(data, file.key)

	if err != nil {
		return err
	}

	// Assign current block
	file.cur_block = block_num
	file.cur_block_data = data

	return nil
}

// Reads from stream, returns the amount of bytes obtained
// Normally reads until the buffer is full, unless the file ends
func (file *FileBlockEncryptReadStream) Read(buf []byte) (int, error) {
	if file.cur_pos >= file.file_size {
		return 0, nil
	}

	filedLength := 0

	for filedLength < len(buf) && file.cur_pos < file.file_size {
		blockIndex := file.cur_pos / file.block_size
		blockOffset := int(file.cur_pos % file.block_size)

		if blockIndex != file.cur_block {
			err := file.FetchBlock(blockIndex)

			if err != nil {
				return 0, err
			}
		}

		blockLen := len(file.cur_block_data)
		bytesToCopy := blockLen - blockOffset
		bytesCanFit := len(buf) - filedLength

		if bytesToCopy > bytesCanFit {
			bytesToCopy = bytesCanFit
		}

		// Copy data into the buffer
		copy(buf[filedLength:filedLength+bytesToCopy], file.cur_block_data[blockOffset:blockOffset+bytesToCopy])

		filedLength += bytesToCopy

		// Seek
		file.cur_pos += int64(bytesToCopy)
	}

	return filedLength, nil
}

func (file *FileBlockEncryptReadStream) Seek(pos int64, whence int) (int64, error) {
	switch whence {
	case 1:
		pos = file.cur_pos + pos
	case 2:
		pos = file.file_size - pos
	}

	if pos < 0 || pos > file.file_size {
		return file.cur_pos, errors.New("Cursor position out of bounds")
	}

	file.cur_pos = pos

	return file.cur_pos, nil
}

func (file *FileBlockEncryptReadStream) Close() {
	file.f.Close()
}
