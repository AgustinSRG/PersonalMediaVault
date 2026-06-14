#!/bin/bash

protoc --go_out=./sse-proto-go --go_opt=paths=source_relative --go-grpc_out=./sse-proto-go --go-grpc_opt=paths=source_relative ./sse.proto
