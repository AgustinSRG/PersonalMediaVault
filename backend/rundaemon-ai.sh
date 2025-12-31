#!/bin/sh

SEMANTIC_SEARCH_ENABLED=YES QDRANT_API_KEY=change_me CLIP_API_AUTH=change_me ./pmvd --clean --fix-consistency --daemon --debug --log-requests --cors-insecure --port 8000 --bind 127.0.0.1
