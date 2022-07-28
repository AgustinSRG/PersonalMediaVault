@echo off

call powershell < configure.ps1

call make-wix.bat

