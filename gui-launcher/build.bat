@echo on

call cmake -DCMAKE_BUILD_TYPE=Release -A x64 -Ssrc -Brelease

call cmake --build release --config Release
