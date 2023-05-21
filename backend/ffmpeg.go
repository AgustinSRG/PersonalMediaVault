// FFMPEG utils
// This file contains the functions to call ffmpeg processes
// to encode media or do any other required tasks

package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/vansante/go-ffprobe"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
)

var (
	FFMPEG_BINARY_PATH  = "/usr/bin/ffmpeg"  // Location of FFMPEG binary
	FFPROBE_BINARY_PATH = "/usr/bin/ffprobe" // Location of FFPROBE binary
)

// Sets FFMPEG config
// ffmpeg_path - Location of FFMPEG binary
// ffprobe_path - Location of FFPROBE binary
func SetFFMPEGBinaries(ffmpeg_path string, ffprobe_path string) {
	FFMPEG_BINARY_PATH = ffmpeg_path
	FFPROBE_BINARY_PATH = ffprobe_path

	ffprobe.SetFFProbeBinPath(FFPROBE_BINARY_PATH)
}

// Result of FFPROBE (media description)
type FFprobeMediaResult struct {
	Type         MediaType // Type of media (video, audio, image)
	Format       string    // File format
	Duration     float64   // Duration
	Width        int32     // Width (px)
	Height       int32     // Height (px)
	Fps          int32     // Frames per second
	Encoded      bool      // True if already encoded to the expected format
	EncodedExt   string    // Extension of the encoded file
	CanCopyAudio bool      // True if the audio stream can be copied
	CanCopyVideo bool      // True if the video stream can be copied
}

// Parses frame rate from string returned by ffprobe
// fr - Frame rate in format 'f/t'
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

// Checks if the video format is valid (mp4)
// formatName - Video format name
func validateFormatNameVideo(formatName string) bool {
	parts := strings.Split(formatName, ",")

	for i := 0; i < len(parts); i++ {
		if parts[i] == "mp4" {
			return true
		}
	}

	return false
}

// Probes a file and returns its properties
// file - File path
func ProbeMediaFileWithFFProbe(file string) (*FFprobeMediaResult, error) {
	LogDebug("[FFPROBE] Probing " + file)
	data, err := ffprobe.GetProbeData(file, 5*time.Second)

	if err != nil {
		return nil, err
	}

	if data.Format == nil {
		return nil, errors.New("Invalid media file")
	}

	format := data.Format.FormatName

	videoStream := data.GetFirstVideoStream()
	audioStream := data.GetFirstAudioStream()

	if videoStream != nil {
		if data.Format.Duration().Seconds() < 0.5 || format == "image2" {
			// Image
			encoded := (format == "png_pipe")

			result := FFprobeMediaResult{
				Type:       MediaTypeImage,
				Format:     format,
				Duration:   0,
				Width:      int32(videoStream.Width),
				Height:     int32(videoStream.Height),
				Fps:        0,
				Encoded:    encoded,
				EncodedExt: "png",
			}

			return &result, nil
		} else {
			// Video
			duration := data.Format.Duration()

			if err != nil {
				return nil, err
			}

			encoded := validateFormatNameVideo(format)
			canCopyVideo := true

			if videoStream.CodecName != "h264" {
				encoded = false
				canCopyVideo = false
			}

			canCopyAudio := true

			if audioStream != nil && audioStream.CodecName != "aac" {
				encoded = false
				canCopyAudio = false
			}

			result := FFprobeMediaResult{
				Type:         MediaTypeVideo,
				Format:       format,
				Duration:     duration.Seconds(),
				Width:        int32(videoStream.Width),
				Height:       int32(videoStream.Height),
				Fps:          ParseFrameRate(videoStream.AvgFrameRate),
				Encoded:      encoded,
				EncodedExt:   "mp4",
				CanCopyVideo: canCopyVideo,
				CanCopyAudio: canCopyAudio,
			}

			return &result, nil
		}
	} else if audioStream != nil {
		// Audio
		duration := data.Format.Duration()

		encoded := ((format == "mp3") && (audioStream.CodecName == "mp3"))

		result := FFprobeMediaResult{
			Type:       MediaTypeAudio,
			Format:     format,
			Duration:   duration.Seconds(),
			Width:      0,
			Height:     0,
			Fps:        0,
			Encoded:    encoded,
			EncodedExt: "mp3",
		}

		return &result, nil
	} else {
		return nil, errors.New("Invalid media file. No audio or video streams.")
	}
}

// Validates subtitles file
// file - Subtitles file
func ValidateSubtitlesFile(file string) bool {
	LogDebug("[FFPROBE] Probing " + file)
	data, err := ffprobe.GetProbeData(file, 5*time.Second)

	if err != nil {
		return false
	}

	if data.Format == nil {
		return false
	}

	format := data.Format.FormatName

	return format == "srt"
}

// Encodes video to MP4 for playback
// originalFilePath - Original video file
// originalFileFormat - Original file format
// originalFileDuration - Original video duration (seconds)
// tempPath - Temporal path to use for the encoding
// resolution - Resolution for re-scaling
// originalWidth - Original width
// originalHeight - Original height
// config - User configuration
// The encoded file will be tempfile/video.mp4
func MakeFFMpegEncodeToMP4Command(originalFilePath string, originalFileFormat string, originalFileDuration float64, tempPath string, resolution *UserConfigResolution, originalWidth int32, originalHeight int32, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

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

	// Resize

	var width = originalWidth
	var height = originalHeight

	resWidth := resolution.Width

	if resWidth%2 != 0 {
		resWidth++
	}

	resHeight := resolution.Height

	if resHeight%2 != 0 {
		resHeight++
	}

	if width > height {
		proportionalHeight := int32(math.Ceil((float64(height)*float64(resWidth)/float64(width))/2) * 2)

		if proportionalHeight > resolution.Height {
			width = int32(math.Ceil((float64(width)*float64(resHeight)/float64(height))/2) * 2)
			height = resHeight
		} else {
			width = resWidth
			height = proportionalHeight
		}
	} else {
		proportionalWidth := int32(math.Ceil((float64(width)*float64(resHeight)/float64(height))/2) * 2)

		if proportionalWidth > resolution.Width {
			height = int32(math.Ceil((float64(height)*float64(resWidth)/float64(width))/2) * 2)
			width = resWidth
		} else {
			height = resHeight
			width = proportionalWidth
		}
	}

	videoFilter += "scale=" + fmt.Sprint(width) + ":" + fmt.Sprint(height)

	args = append(args, "-vf", videoFilter)

	// Force duration
	args = append(args, "-t", fmt.Sprint(originalFileDuration))

	// MP4
	args = append(args, "-max_muxing_queue_size", "9999", "-vcodec", "libx264", "-acodec", "aac", "-pix_fmt", "yuv420p", tempPath+"/video.mp4")

	cmd.Args = args

	return cmd
}

// Encodes video to MP4 for playback (No re-scaling)
// originalFilePath - Original video file
// originalFileFormat - Original file format
// originalFileDuration - Original video duration (seconds)
// tempPath - Temporal path to use for the encoding
// config - User configuration
// canCopyVideo - True if video stream can be copied
// canCopyAudio - True if audio stream can be copied
// The encoded file will be tempfile/video.mp4
func MakeFFMpegEncodeToMP4OriginalCommand(originalFilePath string, originalFileFormat string, originalFileDuration float64, width int32, height int32, tempPath string, config *UserConfig, canCopyVideo bool, canCopyAudio bool) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	// Force duration
	args = append(args, "-t", fmt.Sprint(originalFileDuration))

	// Fix odd dimensions
	if (width%2 != 0) || (height%2 != 0) {
		args = append(args, "-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2")
		canCopyVideo = false
	}

	var vCodec string

	if canCopyVideo {
		vCodec = "copy"
	} else {
		vCodec = "libx264"
		args = append(args, "-pix_fmt", "yuv420p")
	}

	var aCodec string

	if canCopyAudio {
		aCodec = "copy"
	} else {
		aCodec = "aac"
	}

	// MP4
	args = append(args, "-max_muxing_queue_size", "9999", "-vcodec", vCodec, "-acodec", aCodec, tempPath+"/video.mp4")

	cmd.Args = args

	return cmd
}

// Encodes audio to MP3 for playback
// originalFilePath - Original audio file
// originalFileFormat - Original file format
// tempPath - Temporal path to use for the encoding
// config - User configuration
// The encoded file will be tempfile/audio.mp3
func MakeFFMpegEncodeToMP3Command(originalFilePath string, originalFileFormat string, tempPath string, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	args = append(args, "-f", "mp3", "-vn") // Output format

	// Playlist name
	args = append(args, tempPath+"/audio.mp3")

	cmd.Args = args

	return cmd
}

// Encodes image to PNG for display
// originalFilePath - Original image file
// originalFileFormat - Original file format
// tempPath - Temporal path to use for the encoding
// resolution - Resolution for re-scaling
// originalWidth - Original width
// originalHeight - Original height
// config - User configuration
// The encoded file will be tempfile/image.png
func MakeFFMpegEncodeToPNGCommand(originalFilePath string, originalFileFormat string, tempPath string, resolution *UserConfigResolution, originalWidth int32, originalHeight int32, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	// Resize

	var width = originalWidth
	var height = originalHeight

	if width > height {
		proportionalHeight := int32(math.Round(float64(height) * float64(resolution.Width) / float64(width)))

		if proportionalHeight > resolution.Height {
			width = int32(math.Round(float64(width) * float64(resolution.Height) / float64(height)))
			height = resolution.Height
		} else {
			width = resolution.Width
			height = proportionalHeight
		}
	} else {
		proportionalWidth := int32(math.Round(float64(width) * float64(resolution.Height) / float64(height)))

		if proportionalWidth > resolution.Width {
			height = int32(math.Round(float64(height) * float64(resolution.Width) / float64(width)))
			width = resolution.Width
		} else {
			height = resolution.Height
			width = proportionalWidth
		}
	}

	videoFilter := "scale=" + fmt.Sprint(width) + ":" + fmt.Sprint(height)
	args = append(args, "-vf", videoFilter)

	// Setting for image
	args = append(args, "-vframes", "1", "-an")

	// Playlist name
	args = append(args, tempPath+"/image.png")

	cmd.Args = args

	return cmd
}

// Encodes image to PNG for display (no re-scaling)
// originalFilePath - Original image file
// originalFileFormat - Original file format
// tempPath - Temporal path to use for the encoding
// config - User configuration
// The encoded file will be tempfile/image.png
func MakeFFMpegEncodeOriginalToPNGCommand(originalFilePath string, originalFileFormat string, tempPath string, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	if config.EncodingThreads > 0 {
		args = append(args, "-threads", fmt.Sprint(config.EncodingThreads)) // Max threads
	}

	args = append(args, "-f", originalFileFormat, "-i", originalFilePath) // Input file

	// Setting for image
	args = append(args, "-vframes", "1", "-an")

	// Playlist name
	args = append(args, tempPath+"/image.png")

	cmd.Args = args

	return cmd
}

const (
	THUMBNAIL_SIZE       = 250 // Thumbnail height and width (px)
	THUMBNAIL_VIDEO_TIME = 30  // Max thumbnail frame start time
)

// Generates a thumbnail from a video or image file
// originalFilePath - Original media file
// probedata - Media file properties
// Returns the path to a temp file containing the thumbnail
func GenerateThumbnailFromMedia(originalFilePath string, probedata *FFprobeMediaResult) (string, error) {
	if probedata.Type == MediaTypeVideo {
		tmpFile := GetTemporalFileName("jpg", false)
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
		tmpFile := GetTemporalFileName("jpg", false)
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
	PREVIEWS_INTERVAL_SECONDS = 3   // Number of seconds between each preview
	PREVIEWS_IMAGE_WIDTH      = 256 // Preview width (px)
	PREVIEWS_IMAGE_HEIGHT     = 144 // Preview height (px)
)

// Generates previews for a video
// originalFilePath - Original video file
// originalFileFormat - Original file format
// originalFileDuration - Original video duration (seconds)
// tempPath - Temporal path to use for the encoding
// config - User configuration
// The previews will be generated with the following format: tempPath/thumb_%d.jpg
func MakeFFMpegEncodeToPreviewsCommand(originalFilePath string, originalFileFormat string, originalFileDuration float64, tempPath string, config *UserConfig) *exec.Cmd {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

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

	// Force duration
	args = append(args, "-t", fmt.Sprint(originalFileDuration))

	// Playlist name
	args = append(args, tempPath+"/thumb_%d.jpg")

	cmd.Args = args

	return cmd
}

// Runs FFMPEG command asynchronously (the child process can be managed)
// cmd - Command to run
// input_duration - Duration in seconds (used to calculate progress)
// progress_reporter - Function called each time ffmpeg reports progress vie standard error
// Note: If you return true in progress_reporter, the process will be killed (use this to interrupt tasks)
func RunFFMpegCommandAsync(cmd *exec.Cmd, input_duration float64, progress_reporter func(progress float64) bool) error {
	// Configure command
	err := child_process_manager.ConfigureCommand(cmd)
	if err != nil {
		return err
	}

	// Create a pipe to read StdErr
	pipe, err := cmd.StderrPipe()

	if err != nil {
		return err
	}

	// Start the command

	LogDebug("Running command: " + cmd.String())

	err = cmd.Start()

	if err != nil {
		return err
	}

	// Add process as a child process
	child_process_manager.AddChildProcess(cmd.Process) //nolint:errcheck

	// Read stderr line by line

	reader := bufio.NewReader(pipe)

	var finished bool = false

	for !finished {
		line, err := reader.ReadString('\r')

		if err != nil {
			finished = true
		}

		line = strings.ReplaceAll(line, "\r", "")

		LogDebug("[FFMPEG] " + line)

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
				cmd.Process.Kill() //nolint:errcheck
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

// Extracted subtitles file (.srt)
type ExtractedSubtitlesFile struct {
	Id   string // Language identifier
	Name string // Name (for display)
	file string // File path
}

// Extracts subtitles files from media file (usually .mkv)
// originalFilePath - Original media path
// probedata - Media properties
// Returns:
//
//	1 - error
//	2 - Temporal path created, where the files where stored
//	3 - List of files
func ExtractSubtitlesFiles(originalFilePath string, probedata *FFprobeMediaResult) (error, string, []ExtractedSubtitlesFile) {
	result := make([]ExtractedSubtitlesFile, 0)
	addedMap := make(map[string]bool)

	LogDebug("[FFPROBE] Probing " + originalFilePath)
	data, err := ffprobe.GetProbeData(originalFilePath, 5*time.Second)

	if err != nil {
		return err, "", nil
	}

	if data.Format == nil {
		return errors.New("Invalid media file"), "", nil
	}

	subtitleStreams := data.GetStreams(ffprobe.StreamSubtitle)

	tmpFolder, err := GetTemporalFolder(false)

	if err != nil {
		return err, "", nil
	}

	err = os.MkdirAll(tmpFolder, FOLDER_PERMISSION)

	if err != nil {
		return err, "", nil
	}

	for i := 0; i < len(subtitleStreams); i++ {
		stream := subtitleStreams[i]

		lang := stream.Tags.Language

		if lang == "" {
			continue
		}

		if addedMap[lang] {
			continue
		}

		srtPath := path.Join(tmpFolder, lang+".srt")

		LogDebug("Extracting subtitles file for lang: " + lang)

		err = ExtractSubtitlesFromMedia(originalFilePath, probedata, srtPath, i)

		if err != nil {
			LogError(err)
			continue
		}

		entry := ExtractedSubtitlesFile{
			Id:   lang,
			Name: strings.ToUpper(lang),
			file: srtPath,
		}

		result = append(result, entry)
		addedMap[lang] = true
	}

	return nil, tmpFolder, result
}

// Extracts a subtitles file from a media file (usually .mkv)
// originalFilePath - Original media path
// probedata - Media properties
// dest - Destination for the subtitles file
// streamIndex - Subtitles stream index
func ExtractSubtitlesFromMedia(originalFilePath string, probedata *FFprobeMediaResult, dest string, streamIndex int) error {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	args = append(args, "-f", probedata.Format, "-i", originalFilePath) // Input file

	// Setting the stream map
	args = append(args, "-map", "0:s:"+fmt.Sprint(streamIndex))

	// Output
	args = append(args, dest)

	cmd.Args = args

	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

// Extracted audio file (.mp3)
type ExtractedAudioFile struct {
	Id   string // Language identifier
	Name string // Name (for display)
	file string // File path
}

// Extracts audio tracks from media file (usually .mkv)
// originalFilePath - Original media path
// probedata - Media properties
// Returns:
//
//	1 - error
//	2 - Temporal path created, where the files where stored
//	3 - List of files
func ExtractAudioTracks(originalFilePath string, probedata *FFprobeMediaResult) (error, string, []ExtractedAudioFile) {
	result := make([]ExtractedAudioFile, 0)
	addedMap := make(map[string]bool)

	LogDebug("[FFPROBE] Probing " + originalFilePath)
	data, err := ffprobe.GetProbeData(originalFilePath, 5*time.Second)

	if err != nil {
		return err, "", nil
	}

	if data.Format == nil {
		return errors.New("Invalid media file"), "", nil
	}

	audioStreams := data.GetStreams(ffprobe.StreamAudio)

	tmpFolder, err := GetTemporalFolder(false)

	if err != nil {
		return err, "", nil
	}

	err = os.MkdirAll(tmpFolder, FOLDER_PERMISSION)

	if err != nil {
		return err, "", nil
	}

	if len(audioStreams) > 1 {
		for i := 0; i < len(audioStreams); i++ {
			stream := audioStreams[i]

			if stream.Disposition.Default != 0 {
				continue
			}

			lang := stream.Tags.Language

			if lang == "" {
				continue
			}

			if addedMap[lang] {
				continue
			}

			srtPath := path.Join(tmpFolder, lang+".mp3")

			LogDebug("Extracting audio track file for lang: " + lang)

			err = ExtractAudioFromMedia(originalFilePath, probedata, srtPath, i)

			if err != nil {
				LogError(err)
				continue
			}

			entry := ExtractedAudioFile{
				Id:   lang,
				Name: strings.ToUpper(lang),
				file: srtPath,
			}

			result = append(result, entry)
			addedMap[lang] = true
		}
	}

	return nil, tmpFolder, result
}

// Extracts audio track from a media file (usually .mkv)
// originalFilePath - Original media path
// probedata - Media properties
// dest - Destination for the audio file
// streamIndex - Audio stream index
func ExtractAudioFromMedia(originalFilePath string, probedata *FFprobeMediaResult, dest string, streamIndex int) error {
	cmd := exec.Command(FFMPEG_BINARY_PATH)

	args := make([]string, 1)

	args[0] = FFMPEG_BINARY_PATH

	args = append(args, "-y") // Overwrite

	args = append(args, "-f", probedata.Format, "-i", originalFilePath) // Input file

	// Setting the stream map
	args = append(args, "-vn", "-map", "0:a:"+fmt.Sprint(streamIndex))

	// Output
	args = append(args, dest)

	cmd.Args = args

	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}
