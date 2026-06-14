@echo off

call cargo build --release
cp -f .\target\release\pmv-sse.exe pmv-sse.exe
