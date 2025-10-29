#!/bin/bash

echo "🚀 Starting GoDFS Distributed File System"
echo "=========================================="

# Kill any existing processes on our ports
echo "🧹 Cleaning up existing processes..."
lsof -ti :9000 | xargs kill -9 2>/dev/null || true
lsof -ti :9001 | xargs kill -9 2>/dev/null || true
lsof -ti :9002 | xargs kill -9 2>/dev/null || true
lsof -ti :9003 | xargs kill -9 2>/dev/null || true
lsof -ti :8080 | xargs kill -9 2>/dev/null || true

sleep 2

# Create data directories
echo "📁 Creating data directories..."
mkdir -p chunkserver_data_1 chunkserver_data_2 chunkserver_data_3

# Start master server
echo "🎯 Starting Master Server (port 9000)..."
go run ./cmd/master/main.go &
MASTER_PID=$!

# Wait for master to start
sleep 3

# Start chunkservers
echo "💾 Starting Chunkservers..."
go run ./cmd/chunkserver/main.go --port=9001 --data-dir=./chunkserver_data_1 --master=localhost:9000 &
CHUNKSERVER1_PID=$!

go run ./cmd/chunkserver/main.go --port=9002 --data-dir=./chunkserver_data_2 --master=localhost:9000 &
CHUNKSERVER2_PID=$!

go run ./cmd/chunkserver/main.go --port=9003 --data-dir=./chunkserver_data_3 --master=localhost:9000 &
CHUNKSERVER3_PID=$!

# Wait for chunkservers to register
sleep 5

# Start web interface
echo "🌐 Starting Web Interface (port 8080)..."
go run ./cmd/web/main.go &
WEB_PID=$!

echo ""
echo "✅ GoDFS is now running!"
echo ""
echo "📊 Services:"
echo "  • Master Server:    http://localhost:9000"
echo "  • Chunkserver 1:    localhost:9001"
echo "  • Chunkserver 2:    localhost:9002" 
echo "  • Chunkserver 3:    localhost:9003"
echo "  • Web Interface:    http://localhost:8080"
echo ""
echo "🎯 Open http://localhost:8080 in your browser to start uploading files!"
echo ""
echo "Press Ctrl+C to stop all services"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping GoDFS services..."
    kill $MASTER_PID $CHUNKSERVER1_PID $CHUNKSERVER2_PID $CHUNKSERVER3_PID $WEB_PID 2>/dev/null || true
    echo "✅ All services stopped"
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup SIGINT SIGTERM

# Wait for user to stop
wait
