package main

import (
	"flag"
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
	// Command line flags for the chunkservers
	port := flag.String("port", "9001", "Port to listen on")
	dataDir := flag.String("data-dir", "./chunkserver_data", "Data directory for chunks")
	flag.Parse()

	// Created data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	fullDataDir := filepath.Join(homeDir, ".godfs", *dataDir)

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Create chunkserver with data directory
	chunkserverServer := chunkserver.NewServer(fullDataDir)

	// Register chunkserver service
	gfs.RegisterChunkserverServer(grpcServer, chunkserverServer)

	// Enable grpc reflection
	reflection.Register(grpcServer)

	log.Println("Chunkserver listening on port %s, data director: %s", *port, fullDataDir)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
