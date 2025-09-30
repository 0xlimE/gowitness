#!/bin/bash

echo "🛑 Stopping gowitness server..."
killall gowitness || true

echo "🔨 Building server frontend..."
cd web/ui
npm run build
cd ../..

echo "🔨 Building Go binary..."
go build

echo "You can start the report server with:"
echo "./gowitness report server --db-uri sqlite://<DB_FILE> --host 0.0.0.0 --port 7171"
