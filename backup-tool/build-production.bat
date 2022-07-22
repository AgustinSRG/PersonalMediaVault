@echo off

call go build -ldflags="-s -w" -o pmv-backup.exe
