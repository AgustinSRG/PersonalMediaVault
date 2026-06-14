#!/bin/sh

cargo build --release
cp -f ./target/release/pmv-sse pmv-sse
