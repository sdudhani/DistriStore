#!/bin/bash

echo "🚀 Starting GoDFS Distributed File System"
echo "=========================================="

# Kill any existing processes on our ports (graceful then force), and wait until free
echo "🧹 Cleaning up existing processes..."
free_port() {
  local port="$1"
  local pids
  pids=$(lsof -ti :"$port" 2>/dev/null || true)
  if [ -n "$pids" ]; then
    echo "  • Port $port in use by PIDs: $pids — sending SIGTERM"
    kill $pids 2>/dev/null || true
    # wait up to 3s for clean exit
    for i in 1 2 3; do
      sleep 1
      pids=$(lsof -ti :"$port" 2>/dev/null || true)
      [ -z "$pids" ] && break
    done
    # force kill if still present
    pids=$(lsof -ti :"$port" 2>/dev/null || true)
    if [ -n "$pids" ]; then
      echo "    - Forcing kill on PIDs: $pids"
      kill -9 $pids 2>/dev/null || true
    fi
  fi
  # final wait until port is free (up to 3s)
  for i in 1 2 3; do
    lsof -ti :"$port" >/dev/null 2>&1 || return 0
    sleep 1
  done
}

for p in 9000 9001 9002 9003 8080; do
  free_port "$p"
done

sleep 1

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
