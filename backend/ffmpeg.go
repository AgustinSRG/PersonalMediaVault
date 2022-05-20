// FFMPEG utils

package main

import (
	"errors"
	"fmt"
	"math"
	"os/exec"
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

func ProbeMediaFileWithFFProbe(file string) (*FFprobeMediaResult, error) {
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

func MakeFFMpegEncodeToHLSCommand(originalFilePath string, originalFileFormat string, tempPath string, definition *TaskDefinition, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	cmd.Dir = tempPath

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	args = append(args, "-f", "hls") // Output format

	// Video filter
	if !definition.UseOriginalResolution {
		videoFilter := ""

		if definition.Fps > 0 {
			videoFilter += "fps=" + fmt.Sprint(definition.Fps) + ","
		}

		videoFilter += "scale=" + fmt.Sprint(definition.Width) + ":" + fmt.Sprint(definition.Height) +
			":force_original_aspect_ratio=decrease,pad=" + fmt.Sprint(definition.Width) + ":" + fmt.Sprint(definition.Height) +
			":(ow-iw)/2:(oh-ih)/2"

		args = append(args, "-vf", videoFilter)
	}

	// HLS encoder hidden options
	args = append(args, "-profile:v", "baseline")
	args = append(args, "-level", "3.0")
	args = append(args, "-pix_fmt", "yuv420p")
	args = append(args, "-strict", "-2")

	// Force key frames every 3 seconds
	args = append(args, "-force_key_frames", "expr:gte(t,n_forced*3)")

	// HLS playlist options
	args = append(args, "-hls_time", "6")
	args = append(args, "-hls_list_size", "0")
	args = append(args, "-hls_playlist_type", "vod")
	args = append(args, "-hls_segment_filename", "f_%d.ts")

	// Playlist name
	args = append(args, "video.m3u8")

	return cmd
}

func MakeFFMpegEncodeToMP3Command(originalFilePath string, originalFileFormat string, tempPath string, definition *TaskDefinition, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	cmd.Dir = tempPath

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	args = append(args, "-f", "mp3", "-vn") // Output format

	// Playlist name
	args = append(args, "audio.mp3")

	return cmd
}

func MakeFFMpegEncodeToPNGCommand(originalFilePath string, originalFileFormat string, tempPath string, definition *TaskDefinition, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	cmd.Dir = tempPath

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	// Video filter
	if !definition.UseOriginalResolution {
		videoFilter := "scale=" + fmt.Sprint(definition.Width) + ":" + fmt.Sprint(definition.Height) +
			":force_original_aspect_ratio=decrease,format=rgba,pad=" + fmt.Sprint(definition.Width) + ":" + fmt.Sprint(definition.Height) +
			":(ow-iw)/2:(oh-ih)/2:color=#00000000"
		args = append(args, "-vf", videoFilter)
	}

	// Playlist name
	args = append(args, "image.png")

	return cmd
}

const (
	THUMBNAIL_SIZE       = 250
	THUMBNAIL_VIDEO_TIME = 30
)

func GenerateThumbnailFromMedia(originalFilePath string, probedata *FFprobeMediaResult) (string, error) {
	if probedata.Type == MediaTypeVideo {
		tmpFile := GetTemporalFileName("jpg")
		cmd := exec.Command(FFMPEG_BINARY_PATH)

		args := make([]string, 1)

		args[0] = FFMPEG_BINARY_PATH

		args = append(args, "-y") // Overwrite

		args = append(args, "-f", probedata.Format, "-i", originalFilePath) // Input file

		// Setting for image
		args = append(args, "-vframes", "1", "-an")

		// Thumbnail time
		if probedata.Duration > (THUMBNAIL_VIDEO_TIME * 2) {
			args = append(args, "-ss", fmt.Sprint(THUMBNAIL_VIDEO_TIME))
		} else {
			midPoint := math.Floor(probedata.Duration / 2)
			args = append(args, "-ss", fmt.Sprint(midPoint))
		}

		// Video filter
		videoFilter := "scale=" + fmt.Sprint(THUMBNAIL_SIZE) + ":" + fmt.Sprint(THUMBNAIL_SIZE) +
			":force_original_aspect_ratio=decrease,format=rgba,pad=" + fmt.Sprint(THUMBNAIL_SIZE) + ":" + fmt.Sprint(THUMBNAIL_SIZE) +
			":(ow-iw)/2:(oh-ih)/2:color=#00000000"
		args = append(args, "-vf", videoFilter)

		// Playlist name
		args = append(args, tmpFile)

		err := cmd.Run()

		if err != nil {
			return "", err
		}

		return tmpFile, nil
	} else if probedata.Type == MediaTypeImage {
		tmpFile := GetTemporalFileName("jpg")
		cmd := exec.Command(FFMPEG_BINARY_PATH)

		args := make([]string, 1)

		args[0] = FFMPEG_BINARY_PATH

		args = append(args, "-y") // Overwrite

		args = append(args, "-f", probedata.Format, "-i", originalFilePath) // Input file

		// Setting for image
		args = append(args, "-vframes", "1", "-an")

		// Crop image
		x := int32(0)
		y := int32(0)
		w := probedata.Width
		h := probedata.Height

		if w > h {
			x = int32(math.Floor(float64(w-h) / 2))
			w = h
		} else if h > w {
			y = int32(math.Floor(float64(h-w) / 2))
			h = w
		}

		// Video filter
		videoFilter := "crop=" + fmt.Sprint(w) + ":" + fmt.Sprint(h) + ":" + fmt.Sprint(x) + ":" + fmt.Sprint(y) +
			",scale=" + fmt.Sprint(THUMBNAIL_SIZE) + ":" + fmt.Sprint(THUMBNAIL_SIZE) +
			":force_original_aspect_ratio=decrease,format=rgba,pad=" + fmt.Sprint(THUMBNAIL_SIZE) + ":" + fmt.Sprint(THUMBNAIL_SIZE) +
			":(ow-iw)/2:(oh-ih)/2:color=#00000000"
		args = append(args, "-vf", videoFilter)

		// Playlist name
		args = append(args, tmpFile)

		err := cmd.Run()

		if err != nil {
			return "", err
		}

		return tmpFile, nil
	} else {
		// Cant generate a thumbnail
		return "", nil
	}
}

const (
	PREVIEWS_INTERVAL_SECONDS = 3
	PREVIEWS_IMAGE_WIDTH      = 256
	PREVIEWS_IMAGE_HEIGHT     = 144
)

func MakeFFMpegEncodeToPreviewsCommand(originalFilePath string, originalFileFormat string, tempPath string, definition *TaskDefinition, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	cmd.Dir = tempPath

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	// Video filter
	videoFilter := "fps=1/" + fmt.Sprint(PREVIEWS_INTERVAL_SECONDS) +
		",scale=" + fmt.Sprint(PREVIEWS_IMAGE_WIDTH) + ":" + fmt.Sprint(PREVIEWS_IMAGE_HEIGHT) +
		":force_original_aspect_ratio=decrease,pad=" + fmt.Sprint(PREVIEWS_IMAGE_WIDTH) + ":" + fmt.Sprint(PREVIEWS_IMAGE_HEIGHT) +
		":(ow-iw)/2:(oh-ih)/2"
	args = append(args, "-vf", videoFilter)

	// Playlist name
	args = append(args, "thumb_%d.jpg")

	return cmd
}
