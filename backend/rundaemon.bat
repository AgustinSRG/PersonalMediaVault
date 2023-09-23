@echo off

SET FFMPEG_PATH=/ffmpeg/bin/ffmpeg.exe
SET FFPROBE_PATH=/ffmpeg/bin/ffprobe.exe
SET FRONTEND_PATH=../frontend/dist/

call pmvd.exe --clean --fix-consistency --daemon --debug --log-requests --cors-insecure --port 80 --bind 127.0.0.1 --cache-size 50
