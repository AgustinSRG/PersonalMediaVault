@echo off

call go build -ldflags="-s -w" -o pmvd.exe
