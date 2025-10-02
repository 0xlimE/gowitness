#!/bin/bash

# Generic Docker run script for gowitness server
# Usage: ./docker-run-generic.sh <target_dir> <password> <port>
#
# Example: ./docker-run-generic.sh targets/nemlig testpass 8080

set -e

if [ "$#" -lt 3 ]; then
    echo "Usage: $0 <target_dir> <password> <port>"
    echo "Example: $0 targets/nemlig testpass 8080"
    exit 1
fi

TARGET_DIR="$1"
PASSWORD="$2"
PORT="$3"
HOST="${4:-0.0.0.0}"

# Get the absolute path
ABSOLUTE_TARGET_DIR="$(cd "$(dirname "$TARGET_DIR")" && pwd)/$(basename "$TARGET_DIR")"

# Extract target name from path
TARGET_NAME=$(basename "$TARGET_DIR")

# Check if target directory exists
if [ ! -d "$ABSOLUTE_TARGET_DIR" ]; then
    echo "Error: Target directory '$ABSOLUTE_TARGET_DIR' does not exist"
    exit 1
fi

# Check if SQLite database exists
DB_PATH="$ABSOLUTE_TARGET_DIR/${TARGET_NAME}.sqlite3"
if [ ! -f "$DB_PATH" ]; then
    echo "Warning: Database file '$DB_PATH' does not exist"
    echo "Looking for any .sqlite3 file in the directory..."
    DB_FILE=$(find "$ABSOLUTE_TARGET_DIR" -maxdepth 1 -name "*.sqlite3" | head -n 1)
    if [ -z "$DB_FILE" ]; then
        echo "Error: No .sqlite3 database found in '$ABSOLUTE_TARGET_DIR'"
        exit 1
    fi
    TARGET_NAME=$(basename "$DB_FILE" .sqlite3)
    echo "Found database: $DB_FILE"
fi

# Check if screenshots directory exists
SCREENSHOTS_PATH="$ABSOLUTE_TARGET_DIR/screenshots"
if [ ! -d "$SCREENSHOTS_PATH" ]; then
    echo "Warning: Screenshots directory '$SCREENSHOTS_PATH' does not exist"
    echo "Creating it now..."
    mkdir -p "$SCREENSHOTS_PATH"
fi

echo "Starting gowitness server..."
echo "  Target: $TARGET_NAME"
echo "  Database: targets/$TARGET_NAME/${TARGET_NAME}.sqlite3"
echo "  Screenshots: targets/$TARGET_NAME/screenshots"
echo "  Port: $PORT"
echo "  Access at: http://127.0.0.1:$PORT"
echo ""

docker run -it --rm \
  -p "$PORT:$PORT" \
  -v "$ABSOLUTE_TARGET_DIR:/app/targets/$TARGET_NAME:ro" \
  gowitness:latest \
  report server \
  --db-uri "sqlite://targets/$TARGET_NAME/${TARGET_NAME}.sqlite3" \
  --screenshot-path "targets/$TARGET_NAME/screenshots" \
  --password "$PASSWORD" \
  --port "$PORT" \
  --host "$HOST"
