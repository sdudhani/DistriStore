#!/bin/bash

echo "🐳 Starting GoDFS with Docker Compose"
echo "====================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Build and start services
echo "🔨 Building and starting services..."
docker-compose up --build -d

echo ""
echo "✅ GoDFS is now running in Docker!"
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
echo "📋 Useful commands:"
echo "  • View logs:        docker-compose logs -f"
echo "  • Stop services:    docker-compose down"
echo "  • Restart:          docker-compose restart"
echo "  • Clean up:         docker-compose down -v"
