package main

import (
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/sdudhani/godfs/internal/chunkserver"
	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Chunkserver listens on port 9001
	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	//Creating Data directory in the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	//Created the Datadir filepath
	DataDir := filepath.Join(homeDir, ".godfs", "chunkserver_data")

	// Create chunkserver with data directory
	chunkserverServer := chunkserver.NewServer("./chunkserver_data")

	// Register the chunkserver service
	gfs.RegisterChunkserverServer(grpcServer, chunkserverServer)

	// Enable gRPC reflection for grpcurl
	reflection.Register(grpcServer)

	log.Println("Chunkserver listening on port 9001")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
