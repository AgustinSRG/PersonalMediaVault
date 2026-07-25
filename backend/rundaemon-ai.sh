#!/bin/sh

SEMANTIC_SEARCH_ENABLED=YES SSE_BIN_PATH=../semantic-search-engine/server/pmv-sse SSE_MODEL_PATH=../semantic-search-engine/server/models/small ./pmvd --clean --daemon --debug --log-requests --cors-insecure --port 8000 --bind 127.0.0.1
