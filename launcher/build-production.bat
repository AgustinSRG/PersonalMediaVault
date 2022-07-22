@echo off

call windres -o main-res.syso main.rc

call go build -ldflags="-s -w" -o pmv.exe
