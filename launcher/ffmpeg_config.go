// FFmpeg config file

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type FFmpegConfig struct {
	FFProbePath string `json:"ffmpeg_path"`
	FFMpegPath  string `json:"ffprobe_path"`
	H264Codec   string `json:"h264_codec"`
}

func loadFFmpegConfigFromFile() (*FFmpegConfig, error) {
	userConfigDir, err := os.UserConfigDir()

	if err != nil {
		return nil, err
	}

	vaultConfigDir := path.Join(userConfigDir, "PersonalMediaVault")

	err = os.MkdirAll(vaultConfigDir, FOLDER_PERMISSION)

	if err != nil {
		return nil, err
	}

	file := path.Join(vaultConfigDir, "ffmpeg.json")

	b, err := os.ReadFile(file)

	if err != nil {
		return nil, err
	}

	config := FFmpegConfig{}

	// Parse
	err = json.Unmarshal(b, &config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func writeFFmpegConfigToFile(config *FFmpegConfig) error {
	userConfigDir, err := os.UserConfigDir()

	if err != nil {
		return err
	}

	vaultConfigDir := path.Join(userConfigDir, "PersonalMediaVault")

	err = os.MkdirAll(vaultConfigDir, FOLDER_PERMISSION)

	if err != nil {
		return err
	}

	file := path.Join(vaultConfigDir, "ffmpeg.json")

	b, err := json.Marshal(config)

	if err != nil {
		return err
	}

	return os.WriteFile(file, b, FILE_PERMISSION)
}

const H264_CODEC_DEFAULT = "libx264"
const H264_CODEC_FREE = "libopenh264"

func detectH264Codec(config *FFmpegConfig) (string, error) {
	out, err := exec.Command(config.FFMpegPath, "-encoders").Output()

	if err != nil {
		return "", err
	}

	availableCodecs := make(map[string]bool)

	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		lineParts := strings.Split(strings.TrimSpace(line), " ")

		if len(lineParts) < 2 {
			continue
		}

		codecName := lineParts[1]

		availableCodecs[codecName] = true
	}

	if availableCodecs[H264_CODEC_DEFAULT] {
		return H264_CODEC_DEFAULT, nil
	}

	if availableCodecs[H264_CODEC_FREE] {
		return H264_CODEC_FREE, nil
	}

	return "", errors.New("unavailable codec")
}

func checkFFmpegCodec(config *FFmpegConfig) bool {
	out, err := exec.Command(config.FFMpegPath, "-encoders").Output()

	if err != nil {
		return false
	}

	availableCodecs := make(map[string]bool)

	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		lineParts := strings.Split(strings.TrimSpace(line), " ")

		if len(lineParts) < 2 {
			continue
		}

		codecName := lineParts[1]

		availableCodecs[codecName] = true
	}

	return availableCodecs[config.H264Codec]
}

func loadFFmpegConfig() *FFmpegConfig {
	result := &FFmpegConfig{
		FFProbePath: "",
		FFMpegPath:  "",
		H264Codec:   "",
	}

	configFromFile, err := loadFFmpegConfigFromFile()

	if err == nil {
		result = configFromFile
	}

	// Check

	if !fileExists(result.FFMpegPath) {
		result.FFMpegPath = path.Join(getDirName(), "bin", getBinaryFileName("ffmpeg"))

		if !fileExists(result.FFMpegPath) {

			if runtime.GOOS == "windows" {
				result.FFMpegPath = path.Join("C:\\ffmpeg\\bin\\", getBinaryFileName("ffmpeg"))
			} else {
				result.FFMpegPath = path.Join("/usr/bin", getBinaryFileName("ffmpeg"))
			}

			if !fileExists(result.FFMpegPath) {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorCodecBin",
						Other: "Error: Could not find the ffmpeg binary (ffmpeg)",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "FilesMissing",
						Other: "Seems like some required files are missing.",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ReinstallFix",
						Other: "In order to fix this error, you could re-install PersonalMediaVault.",
					},
				})
				fmt.Println(msg)
				os.Exit(1)
			}

		}
	}

	if !fileExists(result.FFProbePath) {
		result.FFProbePath = path.Join(getDirName(), "bin", getBinaryFileName("ffprobe"))

		if !fileExists(result.FFProbePath) {

			if runtime.GOOS == "windows" {
				result.FFProbePath = path.Join("C:\\ffmpeg\\bin\\", getBinaryFileName("ffprobe"))
			} else {
				result.FFProbePath = path.Join("/usr/bin", getBinaryFileName("ffprobe"))
			}

			if !fileExists(result.FFProbePath) {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorProbeBin",
						Other: "Error: Could not find the ffprobe binary (ffprobe)",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "FilesMissing",
						Other: "Seems like some required files are missing.",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ReinstallFix",
						Other: "In order to fix this error, you could re-install PersonalMediaVault.",
					},
				})
				fmt.Println(msg)
				os.Exit(1)
			}
		}
	}

	if result.H264Codec == "" {
		detectedCodec, err := detectH264Codec(result)

		if err == nil {
			result.H264Codec = detectedCodec
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CodecWarning",
					Other: "Warning: Could not detect an encoder for H.264 in your FFMpeg installation. The will lead to errors when trying to encode videos.",
				},
			})
			fmt.Println(msg)

			result.H264Codec = H264_CODEC_DEFAULT
		}
	}

	return result
}
