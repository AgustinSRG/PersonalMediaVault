@echo off

SET FFMPEG_PATH=/ffmpeg/bin/ffmpeg.exe
SET FFPROBE_PATH=/ffmpeg/bin/ffprobe.exe
SET FRONTEND_PATH=../frontend/dist/

call pmv.exe --clean --fix-consistency --daemon --debug --log-requests --cors-insecure
