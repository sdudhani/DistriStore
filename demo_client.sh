#!/bin/bash

echo "🚀 GoDFS Client Demo"
echo "===================="
echo ""

# Check if master server is running
echo "🔍 Checking if master server is running..."
if lsof -i :9000 > /dev/null 2>&1; then
    echo "✅ Master server is running on port 9000"
else
    echo "❌ Master server is not running!"
    echo "Please start the master server first:"
    echo "  go run ./cmd/master/main.go"
    exit 1
fi

# Check if chunkserver is running
echo "🔍 Checking if chunkserver is running..."
if lsof -i :9001 > /dev/null 2>&1; then
    echo "✅ Chunkserver is running on port 9001"
else
    echo "❌ Chunkserver is not running!"
    echo "Please start a chunkserver first:"
    echo "  go run ./cmd/chunkserver/main.go --port=9001 --data-dir=./chunkserver_data_1 --master=localhost:9000"
    exit 1
fi

echo ""
echo "🎯 Starting GoDFS Client Demo..."
echo "================================="
echo ""

# Run the client
go run ./cmd/client/main.go
