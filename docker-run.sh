#!/bin/bash

# Example script to build and run gowitness in Docker
# This assumes you've already built the gowitness binary

set -e

# Build the Docker image
echo "Building Docker image..."
docker build -t gowitness:latest .

# Example: Run gowitness server for nemlig target
echo "Running gowitness server..."
docker run -it --rm \
  -p 8080:8080 \
  -v "$(pwd)/targets/nemlig:/app/targets/nemlig:ro" \
  gowitness:latest \
  report server \
  --db-uri sqlite://targets/nemlig/nemlig.sqlite3 \
  --screenshot-path targets/nemlig/screenshots \
  --password testpass \
  --port 8080 \
  --host 0.0.0.0

# Note: Inside container, we bind to 0.0.0.0 so it's accessible from host
# Access via http://127.0.0.1:8080 on your host machine
