@echo off

call go-winres make

call go build -ldflags="-s -w" -o pmv-backup.exe
