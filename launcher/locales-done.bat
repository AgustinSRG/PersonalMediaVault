@echo off

sh.exe -c "goi18n merge active.*.toml translate.*.toml"

sh.exe -c "rm translate.*.toml"
