@echo off

SET FFMPEG_PATH=/ffmpeg/bin/ffmpeg.exe
SET FFPROBE_PATH=/ffmpeg/bin/ffprobe.exe

call personal-media-vault.exe --clean --daemon --debug --log-requests --cors-insecure
