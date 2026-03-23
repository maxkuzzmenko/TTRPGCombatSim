#!/bin/bash
# Compile sim.go inside Docker and output the Linux binary to shared/
set -e
cd "$(dirname "$0")"

docker run --rm \
  -v "$PWD/sim.go:/build/sim.go" \
  -v "$PWD/shared:/output" \
  golang:1.22-alpine \
  sh -c "cd /build && CGO_ENABLED=0 go build -o /output/sim sim.go"

echo "Built shared/sim successfully."
