// FFMPEG utils

package main

import (
	"errors"
	"strconv"
	"time"

	"github.com/vansante/go-ffprobe"
)

var (
	FFMPEG_BINARY_PATH  = "/usr/bin/ffmpeg"
	FFPROBE_BINARY_PATH = "/usr/bin/ffprobe"
)

func SetFFMPEGBinaries(ffmpeg_path string, ffprobe_path string) {
	FFMPEG_BINARY_PATH = ffmpeg_path
	FFPROBE_BINARY_PATH = ffprobe_path

	ffprobe.SetFFProbeBinPath(FFPROBE_BINARY_PATH)
}

type FFprobeMediaResult struct {
	Type     MediaType
	Format   string
	Duration float64
	Width    int32
	Height   int32
	Fps      int32
}

func ProbeMediaFileWithFFMprobe(file string) (*FFprobeMediaResult, error) {
	data, err := ffprobe.GetProbeData(file, 5*time.Second)

	if err != nil {
		return nil, err
	}

	format := data.Format.FormatName

	videoStream := data.GetFirstVideoStream()
	audioStream := data.GetFirstAudioStream()

	if videoStream != nil {
		if videoStream.Duration == "" {
			// Image
			result := FFprobeMediaResult{
				Type:     MediaTypeImage,
				Format:   format,
				Duration: 0,
				Width:    int32(videoStream.Width),
				Height:   int32(videoStream.Height),
				Fps:      0,
			}

			return &result, nil
		} else {
			// Video
			duration, err := strconv.ParseFloat(videoStream.Duration, 64)

			if err != nil {
				return nil, err
			}

			result := FFprobeMediaResult{
				Type:     MediaTypeVideo,
				Format:   format,
				Duration: duration,
				Width:    int32(videoStream.Width),
				Height:   int32(videoStream.Height),
				Fps:      0,
			}

			return &result, nil
		}
	} else if audioStream != nil {
		// Audio
		duration, err := strconv.ParseFloat(audioStream.Duration, 64)

		if err != nil {
			return nil, err
		}

		result := FFprobeMediaResult{
			Type:     MediaTypeAudio,
			Format:   format,
			Duration: duration,
			Width:    0,
			Height:   0,
			Fps:      0,
		}

		return &result, nil
	} else {
		return nil, errors.New("Invalid media file. No audio or video streams.")
	}
}
