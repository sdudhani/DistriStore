# 🚀 GoDFS - Distributed File System

A distributed file system implementation in Go, inspired by Google File System (GFS). Features automatic replication, fault tolerance, and a beautiful web interface for easy file management.

## ✨ Features

- **🌐 Web Interface**: Beautiful, modern web UI for file upload/download
- **🔄 Automatic Replication**: Files replicated across 3 chunkservers for fault tolerance
- **💾 Distributed Storage**: Files stored across multiple chunkservers
- **❤️ Health Monitoring**: Real-time chunkserver health monitoring
- **⚡ gRPC Communication**: High-performance RPC communication
- **🐳 Docker Support**: Easy deployment with Docker Compose
- **📊 Replication Visualization**: See replication status in real-time

## Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │    │   Master    │    │ Chunkserver │
│             │◄──►│   Server    │◄──►│     1       │
└─────────────┘    └─────────────┘    └─────────────┘
                           │
                           ▼
                   ┌─────────────┐
                   │ Chunkserver │
                   │     2       │
                   └─────────────┘
```

## 🚀 Quick Start

### Option 1: One-Command Start (Recommended)

```bash
# Clone the repository
git clone https://github.com/yourusername/godfs.git
cd godfs

# Start everything with one command
./start.sh
```

Then open **http://localhost:8080** in your browser! 🎉

### Option 2: Docker Compose (Easiest)

```bash
# Clone the repository
git clone https://github.com/yourusername/godfs.git
cd godfs

# Start with Docker
./start-docker.sh
```

Then open **http://localhost:8080** in your browser! 🎉

### Option 3: Manual Start

```bash
# Terminal 1: Start Master Server
go run ./cmd/master/main.go

# Terminal 2: Start Chunkserver 1
go run ./cmd/chunkserver/main.go --port=9001 --data-dir=./chunkserver_data_1 --master=localhost:9000

# Terminal 3: Start Chunkserver 2
go run ./cmd/chunkserver/main.go --port=9002 --data-dir=./chunkserver_data_2 --master=localhost:9000

# Terminal 4: Start Chunkserver 3
go run ./cmd/chunkserver/main.go --port=9003 --data-dir=./chunkserver_data_3 --master=localhost:9000

# Terminal 5: Start Web Interface
go run ./cmd/web/main.go
```

Then open **http://localhost:8080** in your browser! 🎉

## 🌐 Web Interface Features

The GoDFS web interface provides a modern, user-friendly experience:

### 📤 File Upload
- **Drag & Drop**: Easy file upload with modern UI
- **Automatic Replication**: Files automatically replicated across 3 chunkservers
- **Real-time Status**: See upload progress and replication status
- **File Validation**: Built-in file type and size validation

### 📥 File Download
- **One-Click Download**: Download files with a single click
- **Automatic Failover**: Downloads from healthy chunkservers automatically
- **Progress Indicators**: Visual feedback during download

### 📊 System Dashboard
- **Live Status**: Real-time chunkserver health monitoring
- **Replication View**: See which files are replicated where
- **File Management**: List, view, and manage all uploaded files
- **Health Metrics**: Monitor system performance and availability

### 🔄 Replication Visualization
- **Replication Factor**: See how many copies of each file exist
- **Chunkserver Status**: Monitor health of all chunkservers
- **Fault Tolerance**: Visual indicators of system resilience

## Example Usage

```
🚀 GoDFS Client
===============
A Distributed File System Client

✅ Connected to GoDFS master server!

📋 GoDFS Client Menu:
1. 📤 Upload File
2. 📥 Download File
3. 📁 List Files
4. 🔍 System Status
5. ❓ Help
6. 🚪 Exit

Enter your choice: 1
Enter filename: my_document.txt
Enter file content: This is my important document!
📤 Uploading file...
✅ Upload successful: File uploaded successfully with 3 replicas
```

## System Components

### Master Server (`cmd/master/main.go`)
- Manages file metadata
- Tracks chunk locations
- Handles chunkserver registration
- Monitors system health

### Chunkserver (`cmd/chunkserver/main.go`)
- Stores actual file chunks
- Sends periodic heartbeats
- Handles chunk operations (store/retrieve/delete)

### Client (`cmd/client/main.go`)
- Interactive file operations
- User-friendly interface
- System status monitoring

## Configuration

### Master Server
- **Port**: 9000 (default)
- **Heartbeat timeout**: 30 seconds
- **Health check interval**: 30 seconds

### Chunkserver
- **Port**: 9001, 9002, 9003 (configurable)
- **Data directory**: `./chunkserver_data_N`
- **Heartbeat interval**: 10 seconds

## Testing

Run the existing test files:

```bash
# Simple test
go run simple_test.go

# Quick test
go run quick_test.go

# Comprehensive test
go run test_client.go
```

## Project Structure

```
godfs/
├── cmd/
│   ├── master/main.go          # Master server
│   ├── chunkserver/main.go     # Chunkserver
│   └── client/main.go          # Interactive client
├── internal/
│   ├── master/server.go        # Master server logic
│   └── chunkserver/server.go   # Chunkserver logic
├── pkg/gfs/
│   ├── gfs.proto              # Protocol definitions
│   ├── gfs.pb.go              # Generated protobuf
│   └── gfs_grpc.pb.go         # Generated gRPC
├── scripts/
│   └── start_chunkservers.sh  # Startup script
└── README.md                  # This file
```

## Features Demonstrated

- ✅ **File Upload**: Store files with automatic replication
- ✅ **File Download**: Retrieve files with failover
- ✅ **System Monitoring**: Health checks and status
- ✅ **Replication**: Fault tolerance across chunkservers
- ✅ **Interactive Interface**: User-friendly client
- ✅ **Distributed Architecture**: Master-chunkserver communication

## Next Steps

- Add authentication and authorization
- Implement file versioning
- Add compression and encryption
- Scale to multiple data centers
- Add web-based management interface

---

**GoDFS** - A distributed file system built with Go and gRPC 🚀
