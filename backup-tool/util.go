// Utilities

package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	FILE_PERMISSION   = 0600 // Read/Write
	FOLDER_PERMISSION = 0700 // Read/Write/Run
)

func CheckFileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return false
	} else {
		return false
	}
}

// Renames and replaces file (Atomic)
// If it fails, tries again up to 3 times, waiting 500 ms (this is to wait for any other program to unlock the file)
// tmpFile - The temporal file to move
// destFile - The destination file name
// returns the error
func RenameAndReplace(tmpFile string, destFile string) error {
	retriesLeft := 3
	var err error = nil

	for retriesLeft > 0 {
		err = os.Rename(tmpFile, destFile)

		if err == nil {
			return nil
		}

		retriesLeft--

		time.Sleep(500 * time.Millisecond)
	}

	return err
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

var PreviousLine = ""

func printLineOverWrite(line string) {
	paddedLine := line
	for len(PreviousLine) > len(paddedLine) {
		paddedLine = paddedLine + " "
	}

	PreviousLine = line

	fmt.Fprint(os.Stderr, "\r"+paddedLine)
}

func resetLineOverWrite() {
	PreviousLine = ""
}
