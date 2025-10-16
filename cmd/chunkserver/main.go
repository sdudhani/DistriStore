package main

import (
	"log"
	"net"

	"github.com/sdudhani/godfs/internal/chunkserver"
	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
)

func main() {
	// Chunkserver listens on port 9001
	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Create chunkserver with data directory
	chunkserverServer := chunkserver.NewServer("./chunkserver_data")

	// Register the chunkserver service
	gfs.RegisterChunkserverServer(grpcServer, chunkserverServer)

	log.Println("Chunkserver listening on port 9001")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
