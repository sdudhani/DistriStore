package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/sdudhani/godfs/internal/chunkserver"
	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Command line flags for the chunkservers
	port := flag.String("port", "9001", "Port to listen on")
	dataDir := flag.String("data-dir", "./chunkserver_data", "Data directory for chunks")
	masterAddr := flag.String("master", "localhost:9000", "Master server address")
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

	log.Printf("Chunkserver listening on port %s, data director: %s", *port, fullDataDir)

	// Register with master server
	go registerWithMaster(*masterAddr, "localhost:"+*port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// registerWithMaster registers this chunkserver with the master
func registerWithMaster(masterAddr, chunkserverAddr string) {
	// Wait a bit for the server to start
	time.Sleep(2 * time.Second)

	// Connect to master
	conn, err := grpc.Dial(masterAddr, grpc.WithInsecure()) // TODO: Replace with secure connection in production
	if err != nil {
		log.Printf("Failed to connect to master %s: %v", masterAddr, err)
		return
	}
	defer conn.Close()

	masterClient := gfs.NewMasterClient(conn)

	// Send periodic heartbeats
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := masterClient.Heartbeat(ctx, &gfs.HeartbeatRequest{
			ChunkserverId: chunkserverAddr,
		})
		cancel()

		if err != nil {
			log.Printf("Failed to send heartbeat to master: %v", err)
		} else {
			log.Printf("Sent heartbeat to master")
		}
	}
}
