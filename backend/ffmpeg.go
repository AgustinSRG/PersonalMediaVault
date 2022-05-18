// FFMPEG utils

package main

import "github.com/vansante/go-ffprobe"

var (
	FFMPEG_BINARY_PATH  = "/usr/bin/ffmpeg"
	FFPROBE_BINARY_PATH = "/usr/bin/ffprobe"
)

func SetFFMPEGBinaries(ffmpeg_path string, ffprobe_path string) {
	FFMPEG_BINARY_PATH = ffmpeg_path
	FFPROBE_BINARY_PATH = ffprobe_path

	ffprobe.SetFFProbeBinPath(FFPROBE_BINARY_PATH)
}
