# GoWitness Docker Setup

This directory contains a simplified Dockerfile that runs a pre-built `gowitness` binary in a container.

## Prerequisites

1. Build the `gowitness` binary first (outside the container):
   ```bash
   go build -o gowitness
   ```

2. Ensure you have Docker installed and running.

## Building the Docker Image

```bash
docker build -t gowitness:latest .
```

## Running the Container

### Quick Start with Helper Scripts

**Using the generic script (recommended):**
```bash
./docker-run-generic.sh targets/nemlig testpass 8080
```

This will:
- Mount the `targets/nemlig` directory into the container
- Start the gowitness server on port 8080
- Use password `testpass` for authentication
- Automatically detect the database file

### Manual Docker Run

```bash
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
```

**Important:** 
- The container binds to `0.0.0.0` internally (so Docker can expose it)
- Access the server from your host at `http://127.0.0.1:8080`
- The volume is mounted as read-only (`:ro`) for safety

### Running for Different Targets

For almbrand:
```bash
./docker-run-generic.sh targets/almbrand mypassword 8081
```

For defenddenmark:
```bash
./docker-run-generic.sh targets/defenddenmark mypassword 8082
```

## Directory Structure Expected

Your target directory should have this structure:
```
targets/nemlig/
├── nemlig.sqlite3          # The database file
└── screenshots/            # Directory containing screenshot files
```

## Notes

- The Dockerfile uses `ghcr.io/go-rod/rod` as the base image (includes Chrome/Chromium)
- The gowitness binary must be built before building the Docker image
- Port 8080 is exposed by default, but you can map it to any host port
- The `/app/targets` directory is set up as a volume mount point
