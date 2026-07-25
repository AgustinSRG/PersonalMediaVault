#!/bin/sh

set -e

cargo build --release
cp -f ./target/release/pmv-sse pmv-sse
