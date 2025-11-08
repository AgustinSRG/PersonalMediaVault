@echo off

call update-translations.bat

call cargo build --release
