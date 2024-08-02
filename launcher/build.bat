@echo off

call go-winres make

call go build -o pmv.exe
