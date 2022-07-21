@echo off

call windres -o main-res.syso main.rc

call go build -o pmv.exe
