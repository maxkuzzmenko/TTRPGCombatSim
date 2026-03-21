#!/bin/bash
# Compile dnd.go inside Docker and output the Linux binary to shared/
set -e
cd "$(dirname "$0")"

docker run --rm \
  -v "$PWD/dnd.go:/build/dnd.go" \
  -v "$PWD/shared:/output" \
  golang:1.22-alpine \
  sh -c "cd /build && CGO_ENABLED=0 go build -o /output/dnd dnd.go"

echo "Built shared/dnd successfully."
