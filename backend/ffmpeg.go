// FFMPEG utils

package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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

func ParseFrameRate(fr string) int32 {
	if fr == "" {
		return 0
	}
	parts := strings.Split(fr, "/")
	if len(parts) == 2 {
		n, err := strconv.Atoi(parts[0])

		if err != nil {
			return 0
		}

		n2, err := strconv.Atoi(parts[1])

		if err != nil || n2 == 0 {
			return 0
		}

		return int32(n) / int32(n2)
	} else if len(parts) == 1 {
		n, err := strconv.Atoi(parts[0])

		if err != nil {
			return 0
		}

		return int32(n)
	} else {
		return 0
	}
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
		if videoStream.Duration == "" || videoStream.DurationTs <= 1 {
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
				Fps:      ParseFrameRate(videoStream.AvgFrameRate),
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

func MakeFFMpegEncodeToMP4Command(originalFilePath string, originalFileFormat string, tempPath string, resolution *UserConfigResolution, config *UserConfig) *exec.Cmd {
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
	videoFilter := ""

	if resolution.Fps > 0 {
		videoFilter += "fps=" + fmt.Sprint(resolution.Fps) + ","
	}

	videoFilter += "scale=" + fmt.Sprint(resolution.Width) + ":" + fmt.Sprint(resolution.Height) +
		":force_original_aspect_ratio=decrease,pad=" + fmt.Sprint(resolution.Width) + ":" + fmt.Sprint(resolution.Height) +
		":(ow-iw)/2:(oh-ih)/2"

	args = append(args, "-vf", videoFilter)

	// MP4
	args = append(args, "-vcodec", "libx264", "-acodec", "aac", "video.mp4")

	cmd.Args = args

	return cmd
}

func MakeFFMpegEncodeToMP4OriginalCommand(originalFilePath string, originalFileFormat string, tempPath string, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	cmd.Dir = tempPath

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	// MP4
	args = append(args, "-vcodec", "libx264", "-acodec", "aac", "video.mp4")

	cmd.Args = args

	return cmd
}

func MakeFFMpegEncodeToMP3Command(originalFilePath string, originalFileFormat string, tempPath string, config *UserConfig) *exec.Cmd {
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

	cmd.Args = args

	return cmd
}

func MakeFFMpegEncodeToPNGCommand(originalFilePath string, originalFileFormat string, tempPath string, resolution *UserConfigResolution, config *UserConfig) *exec.Cmd {
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
	videoFilter := "scale=" + fmt.Sprint(resolution.Width) + ":" + fmt.Sprint(resolution.Height) +
		":force_original_aspect_ratio=decrease,format=rgba,pad=" + fmt.Sprint(resolution.Width) + ":" + fmt.Sprint(resolution.Height) +
		":(ow-iw)/2:(oh-ih)/2:color=#00000000"
	args = append(args, "-vf", videoFilter)

	// Playlist name
	args = append(args, "image.png")

	cmd.Args = args

	return cmd
}

func MakeFFMpegEncodeOriginalToPNGCommand(originalFilePath string, originalFileFormat string, tempPath string, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	cmd.Dir = tempPath

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	// Playlist name
	args = append(args, "image.png")

	cmd.Args = args

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

		// Outout
		args = append(args, tmpFile)

		cmd.Args = args

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

		// Output
		args = append(args, tmpFile)

		cmd.Args = args

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

func MakeFFMpegEncodeToPreviewsCommand(originalFilePath string, originalFileFormat string, tempPath string, config *UserConfig) *exec.Cmd {
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

	cmd.Args = args

	return cmd
}

func RunFFMpegCommandAsync(cmd *exec.Cmd, input_duration float64, progress_reporter func(progress float64) bool) error {
	// Create a pipe to read StdErr
	pipe, err := cmd.StderrPipe()

	if err != nil {
		return err
	}

	// Start the command

	err = cmd.Start()

	if err != nil {
		return err
	}

	// Read stderr line by line

	reader := bufio.NewReader(pipe)

	var finished bool = false

	for !finished {
		line, err := reader.ReadString('\r')

		if err != nil {
			finished = true
			continue
		}

		line = strings.ReplaceAll(line, "\r", "")

		if !strings.HasPrefix(line, "frame=") {
			continue // Not a progress line
		}

		parts := strings.Split(line, "time=")

		if len(parts) < 2 {
			continue
		}

		parts = strings.Split(strings.Trim(parts[1], " "), " ")

		if len(parts) < 1 {
			continue
		}

		parts = strings.Split(parts[0], ":")

		if len(parts) != 3 {
			continue
		}

		hours, _ := strconv.Atoi(parts[0])
		minutes, _ := strconv.Atoi(parts[1])
		seconds, _ := strconv.ParseFloat(parts[2], 64)

		out_duration := float64(hours)*3600 + float64(minutes)*60 + seconds

		if out_duration > 0 && out_duration < input_duration {
			shouldKill := progress_reporter(out_duration * 100 / input_duration)

			if shouldKill {
				cmd.Process.Signal(syscall.SIGINT)
			}
		}
	}

	// Wait for ending

	err = cmd.Wait()

	if err != nil {
		return err
	}

	return nil
}
