@echo off

goi18n extract

sh.exe -c "goi18n merge active.*.toml"

echo Translate the translate.*.toml files and run locales-done.bat
