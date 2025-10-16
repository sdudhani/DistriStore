package chunkserver

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sdudhani/godfs/pkg/gfs"
)

// Server implements the gRPC Chunkserver server
type Server struct {
	gfs.UnimplementedChunkserverServer
	DataDir string // Directory to store chunks
}

// NewServer creates a new chunkserver instance
func NewServer(dataDir string) *Server {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	return &Server{
		DataDir: dataDir,
	}
}

// StoreChunk stores a chunk of data
func (s *Server) StoreChunk(ctx context.Context, req *gfs.StoreChunkRequest) (*gfs.StoreChunkResponse, error) {
	chunkHandle := req.GetChunkHandle()
	data := req.GetData()

	// Create file path for this chunk
	chunkPath := filepath.Join(s.DataDir, chunkHandle)

	// Write data to file
	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		log.Printf("Failed to store chunk %s: %v", chunkHandle, err)
		return &gfs.StoreChunkResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to store chunk: %v", err),
		}, nil
	}

	log.Printf("Stored chunk %s (%d bytes)", chunkHandle, len(data))
	return &gfs.StoreChunkResponse{
		Success: true,
		Message: "Chunk stored successfully",
	}, nil
}

// RetrieveChunk retrieves a chunk of data
func (s *Server) RetrieveChunk(ctx context.Context, req *gfs.RetrieveChunkRequest) (*gfs.RetrieveChunkResponse, error) {
	chunkHandle := req.GetChunkHandle()

	// Create file path for this chunk
	chunkPath := filepath.Join(s.DataDir, chunkHandle)

	// Read data from file
	data, err := os.ReadFile(chunkPath)
	if err != nil {
		log.Printf("Failed to retrieve chunk %s: %v", chunkHandle, err)
		return &gfs.RetrieveChunkResponse{
			Success: false,
			Data:    nil,
			Message: fmt.Sprintf("Failed to retrieve chunk: %v", err),
		}, nil
	}

	log.Printf("Retrieved chunk %s (%d bytes)", chunkHandle, len(data))
	return &gfs.RetrieveChunkResponse{
		Success: true,
		Data:    data,
		Message: "Chunk retrieved successfully",
	}, nil
}

// DeleteChunk deletes a chunk
func (s *Server) DeleteChunk(ctx context.Context, req *gfs.DeleteChunkRequest) (*gfs.DeleteChunkResponse, error) {
	chunkHandle := req.GetChunkHandle()

	// Create file path for this chunk
	chunkPath := filepath.Join(s.DataDir, chunkHandle)

	// Delete file
	if err := os.Remove(chunkPath); err != nil {
		log.Printf("Failed to delete chunk %s: %v", chunkHandle, err)
		return &gfs.DeleteChunkResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete chunk: %v", err),
		}, nil
	}

	log.Printf("Deleted chunk %s", chunkHandle)
	return &gfs.DeleteChunkResponse{
		Success: true,
		Message: "Chunk deleted successfully",
	}, nil
}
