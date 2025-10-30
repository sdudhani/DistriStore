# DistriStore - Distributed File System

A distributed file system implementation in Go, inspired by the Google File System (GFS). Features automatic replication, fault tolerance, and a web interface for easy file management.


## Features

- ** Web Interface**: Beautiful, modern web UI for file upload/download
- ** Automatic Replication**: Files replicated across 3 chunkservers for fault tolerance
- ** Distributed Storage**: Files stored across multiple chunkservers
- ** Health Monitoring**: Real-time chunkserver health monitoring
- ** gRPC Communication**: High-performance RPC communication
- ** One-Command Setup**: Start everything with a single script
- ** Replication Visualization**: See replication status in real-time

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

## Quick Start

### One-Command Start 

```bash
# Clone the repository
git clone https://github.com/yourusername/godfs.git
cd godfs

# Start everything with one command
./start.sh
```

Then open **http://localhost:8080** in your browser! 

What you can do:
- Upload a file from the web UI
- See each file's replication count
- Download any file to verify contents

### Manual Start (Alternative)

If you prefer to start services individually:

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

Then open **http://localhost:8080** in your browser! 

##  Web Interface Features

The GoDFS web interface provides a all the essential features as: 

### File Upload
- Upload any file from your machine
- Automatic replication across up to 3 chunkservers
- Status and replica count shown in the list

### File Download
- One-click download from the list
- Automatic failover if a chunkserver is down

### System Dashboard
- Master address display
- Chunkserver health (basic view)

### Replication Visualization
- Replication factor shown per file

## Verify Replication

From the Web UI:
- Upload a file
- Refresh the page; you should see "3 replicas" if three chunkservers are running

From CLI (optional):
```bash
go run ./test_client.go
# or use the master RPC directly in code via GetChunkLocations
```

##  Where Files Are Stored

Each chunkserver writes file chunks to a data directory under your home directory:
- macOS/Linux: `~/.godfs/<data-dir>/<filename>-0`
- Example if you started with `--data-dir=./chunkserver_data_1`:
  - `/Users/<you>/.godfs/chunkserver_data_1/myfile.txt-0`

You should see identical chunk files across multiple chunkserver data dirs when replication succeeds.

## Example Usage

```
 GoDFS Client
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

### Web (`cmd/web/main.go`)
- Simple HTTP UI for upload/download
- Shows basic health info and replication counts

### Client (`cmd/client/main.go`) (optional)
- Interactive CLI to test RPCs

## Configuration

### Master Server
- **Port**: 9000 (default)
- **Heartbeat timeout**: 30 seconds
- **Health check interval**: 30 seconds

### Chunkserver
- **Port**: 9001, 9002, 9003 (configurable)
- **Data directory**: `./chunkserver_data_N`
- **Heartbeat interval**: 10 seconds

## Troubleshooting

Ports already in use (8080, 9000–9003):
```bash
kill -9 $(lsof -ti :8080 :9000 :9001 :9002 :9003) 2>/dev/null || true
```

Master can't find chunkservers:
- Wait ~10s for heartbeats to register
- Ensure you started three chunkservers (or reduce replication expectations)

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
│   ├── web/main.go             # Web interface
│   └── client/main.go          # Interactive client (optional)
├── internal/
│   ├── master/server.go        # Master server logic
│   └── chunkserver/server.go   # Chunkserver logic
├── pkg/gfs/
│   ├── gfs.proto              # Protocol definitions
│   ├── gfs.pb.go              # Generated protobuf
│   └── gfs_grpc.pb.go         # Generated gRPC
├── start.sh                   # One-command startup for all services
├── scripts/
│   └── start_chunkservers.sh  # Start chunkservers manually
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

**GoDFS** - A distributed file system built with Go and gRPC 
