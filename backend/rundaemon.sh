#!/bin/sh

./pmvd --clean --fix-consistency --daemon --debug --log-requests --cors-insecure --port 8000 --bind 127.0.0.1
