# Gowitness Fly.io Deployment Guide

## Prerequisites

1. **Install flyctl**: `curl -L https://fly.io/install.sh | sh`
2. **Login to Fly.io**: `flyctl auth login`
3. **Have your gowitness project data ready** (SQLite database + screenshots folder)

## Quick Setup Commands

### 1. Deploy to Fly.io
```bash
# Initialize the fly app (only first time)
fly launch --no-deploy

# Create a persistent volume for your data
fly volumes create gowitness_data --region iad --size 10

# Deploy the application
fly deploy
```

### 2. Upload Your Project Data

You have two options to get your local project data to fly.io:

#### Option A: Using fly ssh (Recommended)
```bash
# Connect to your app via SSH
fly ssh console

# From inside the container, you can upload files
# (You'll need to use scp or another method to get files there first)
```

#### Option B: Rebuild with data included
```bash
# Copy your project data to a staging directory
cp -r targets/test_run staging_data/

# Modify Dockerfile.fly to copy this data during build
# Then redeploy
fly deploy
```

### 3. Configure for Your Project

Edit `fly.toml` and change the `PROJECT_NAME` environment variable:
```toml
[env]
  PROJECT_NAME = "test_run"  # Change to your project name
```

## File Structure Expected on Fly.io

```
/data/
├── test_run.sqlite3       # Your SQLite database
└── screenshots/           # Your screenshots directory
    ├── http---example.com-80.jpeg
    └── ...
```

## Local Development Workflow

### 1. Run Scans Locally
```bash
# Initialize a new project
gowitness scan init --db-uri sqlite://./targets/myproject/myproject.sqlite3

# Scan from a file
gowitness scan file -f targets.txt --write-db --db-uri sqlite://./targets/myproject/myproject.sqlite3 --screenshot-path ./targets/myproject/screenshots

# Scan a single URL
gowitness scan single --url "https://example.com" --write-db --db-uri sqlite://./targets/myproject/myproject.sqlite3 --screenshot-path ./targets/myproject/screenshots

# Scan a CIDR range
gowitness scan cidr -t 20 -c 10.0.0.0/24 --write-db --db-uri sqlite://./targets/myproject/myproject.sqlite3 --screenshot-path ./targets/myproject/screenshots
```

### 2. Test Server Locally
```bash
# Test the server with your local data
gowitness report server --db-uri sqlite://./targets/myproject/myproject.sqlite3 --screenshot-path ./targets/myproject/screenshots --port 7171
```

### 3. Deploy to Fly.io
```bash
# Update PROJECT_NAME in fly.toml to match your project
# Upload your data to the volume
# Deploy
fly deploy
```

## Environment Variables

- `PROJECT_NAME`: Name of your project (matches the SQLite filename)
- Database path: `/data/${PROJECT_NAME}.sqlite3`
- Screenshots path: `/data/screenshots`

## Accessing Your Deployed App

Once deployed, your app will be available at:
`https://your-app-name.fly.dev`

## Volume Management

```bash
# List volumes
fly volumes list

# Create new volume
fly volumes create gowitness_data --size 20  # 20GB

# Extend existing volume
fly volumes extend <volume-id> --size 30

# Backup volume
fly volumes snapshots create <volume-id>
```

## Security

To add password protection, modify the startup script in Dockerfile.fly:
```bash
exec gowitness report server \
    --host 0.0.0.0 \
    --port 7171 \
    --db-uri "sqlite://$DB_PATH" \
    --screenshot-path "$SCREENSHOTS_PATH" \
    --password "your-secure-password"
```

## Troubleshooting

### Check logs
```bash
fly logs
```

### SSH into container
```bash
fly ssh console
```

### Check if files exist
```bash
fly ssh console -C "ls -la /data/"
```