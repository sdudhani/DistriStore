package master

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
)

// FileMetadata represents metadata for a file
type FileMetadata struct {
	ChunkHandles []string
	Size         int64
}

// Server implements the gRPC Master server
type Server struct {
	gfs.UnimplementedMasterServer

	// Metadata storage
	mu             sync.RWMutex
	fileMetadata   map[string]*FileMetadata // filename -> metadata
	chunkLocations map[string][]string      // chunkHandle -> chunkserver addresses

	// Chunkserver client for communication
	chunkserverClient gfs.ChunkserverClient
}

// NewServer creates a new master server
func NewServer() *Server {
	return &Server{
		fileMetadata:   make(map[string]*FileMetadata),
		chunkLocations: make(map[string][]string),
	}
}

// getChunkserverClient creates a connection to the chunkserver
func (s *Server) getChunkserverClient() (gfs.ChunkserverClient, error) {
	conn, err := grpc.NewClient("localhost:9001", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return gfs.NewChunkserverClient(conn), nil
}

// Heartbeat handles chunkserver heartbeats
func (s *Server) Heartbeat(ctx context.Context, req *gfs.HeartbeatRequest) (*gfs.HeartbeatResponse, error) {
	chunkserverID := req.GetChunkserverId()
	log.Printf("Received heartbeat from: %s", chunkserverID)

	// TODO: Update chunkserver status and locations
	// For now, just acknowledge the heartbeat

	return &gfs.HeartbeatResponse{Message: "Heartbeat received"}, nil
}

// UploadFile handles file uploads
func (s *Server) UploadFile(ctx context.Context, req *gfs.UploadFileRequest) (*gfs.UploadFileResponse, error) {
	filename := req.GetFilename()
	data := req.GetData()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a chunk handle for this file
	chunkHandle := fmt.Sprintf("%s-0", filename)

	// Get chunkserver client
	chunkserverClient, err := s.getChunkserverClient()
	if err != nil {
		log.Printf("Failed to connect to chunkserver: %v", err)
		return &gfs.UploadFileResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to connect to chunkserver: %v", err),
		}, nil
	}

	// Store the chunk on chunkserver
	storeReq := &gfs.StoreChunkRequest{
		ChunkHandle: chunkHandle,
		Data:        data,
	}

	_, err = chunkserverClient.StoreChunk(ctx, storeReq)
	if err != nil {
		log.Printf("Failed to store chunk %s: %v", chunkHandle, err)
		return &gfs.UploadFileResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to store chunk: %v", err),
		}, nil
	}

	// Update metadata
	s.fileMetadata[filename] = &FileMetadata{
		ChunkHandles: []string{chunkHandle},
		Size:         int64(len(data)),
	}

	// Record chunk location
	s.chunkLocations[chunkHandle] = []string{"localhost:9001"}

	log.Printf("Uploaded file %s (%d bytes) as chunk %s", filename, len(data), chunkHandle)

	return &gfs.UploadFileResponse{
		Success: true,
		Message: "File uploaded successfully",
	}, nil
}

// DownloadFile handles file downloads
func (s *Server) DownloadFile(ctx context.Context, req *gfs.DownloadFileRequest) (*gfs.DownloadFileResponse, error) {
	filename := req.GetFilename()

	s.mu.RLock()
	fileMeta, exists := s.fileMetadata[filename]
	s.mu.RUnlock()

	if !exists {
		return &gfs.DownloadFileResponse{
			Success: false,
			Data:    nil,
			Message: "File not found",
		}, nil
	}

	// For now, we only support single-chunk files
	if len(fileMeta.ChunkHandles) != 1 {
		return &gfs.DownloadFileResponse{
			Success: false,
			Data:    nil,
			Message: "Multi-chunk files not supported yet",
		}, nil
	}

	chunkHandle := fileMeta.ChunkHandles[0]

	// Get chunkserver client
	chunkserverClient, err := s.getChunkserverClient()
	if err != nil {
		log.Printf("Failed to connect to chunkserver: %v", err)
		return &gfs.DownloadFileResponse{
			Success: false,
			Data:    nil,
			Message: fmt.Sprintf("Failed to connect to chunkserver: %v", err),
		}, nil
	}

	// Retrieve chunk from chunkserver
	retrieveReq := &gfs.RetrieveChunkRequest{
		ChunkHandle: chunkHandle,
	}

	resp, err := chunkserverClient.RetrieveChunk(ctx, retrieveReq)
	if err != nil {
		log.Printf("Failed to retrieve chunk %s: %v", chunkHandle, err)
		return &gfs.DownloadFileResponse{
			Success: false,
			Data:    nil,
			Message: fmt.Sprintf("Failed to retrieve chunk: %v", err),
		}, nil
	}

	if !resp.GetSuccess() {
		return &gfs.DownloadFileResponse{
			Success: false,
			Data:    nil,
			Message: resp.GetMessage(),
		}, nil
	}

	log.Printf("Downloaded file %s (%d bytes)", filename, len(resp.GetData()))

	return &gfs.DownloadFileResponse{
		Success: true,
		Data:    resp.GetData(),
		Message: "File downloaded successfully",
	}, nil
}

// ListFiles lists files in a directory
func (s *Server) ListFiles(ctx context.Context, req *gfs.ListFilesRequest) (*gfs.ListFilesResponse, error) {
	path := req.GetPath()

	s.mu.RLock()
	defer s.mu.RUnlock()

	var files []string
	for filename := range s.fileMetadata {
		// Simple prefix matching for now
		if path == "" || filename[:len(path)] == path {
			files = append(files, filename)
		}
	}

	return &gfs.ListFilesResponse{
		Success: true,
		Files:   files,
		Message: fmt.Sprintf("Found %d files", len(files)),
	}, nil
}

// DeleteFile handles file deletion
func (s *Server) DeleteFile(ctx context.Context, req *gfs.DeleteFileRequest) (*gfs.DeleteFileResponse, error) {
	filename := req.GetFilename()

	s.mu.Lock()
	defer s.mu.Unlock()

	fileMeta, exists := s.fileMetadata[filename]
	if !exists {
		return &gfs.DeleteFileResponse{
			Success: false,
			Message: "File not found",
		}, nil
	}

	// Get chunkserver client
	chunkserverClient, err := s.getChunkserverClient()
	if err != nil {
		log.Printf("Failed to connect to chunkserver: %v", err)
		return &gfs.DeleteFileResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to connect to chunkserver: %v", err),
		}, nil
	}

	// Delete chunks from chunkserver
	for _, chunkHandle := range fileMeta.ChunkHandles {
		deleteReq := &gfs.DeleteChunkRequest{
			ChunkHandle: chunkHandle,
		}

		_, err := chunkserverClient.DeleteChunk(ctx, deleteReq)
		if err != nil {
			log.Printf("Failed to delete chunk %s: %v", chunkHandle, err)
		}

		// Remove from chunk locations
		delete(s.chunkLocations, chunkHandle)
	}

	// Remove file metadata
	delete(s.fileMetadata, filename)

	log.Printf("Deleted file %s", filename)

	return &gfs.DeleteFileResponse{
		Success: true,
		Message: "File deleted successfully",
	}, nil
}

// GetChunkLocations returns chunk locations for a file
func (s *Server) GetChunkLocations(ctx context.Context, req *gfs.GetChunkLocationsRequest) (*gfs.GetChunkLocationsResponse, error) {
	filename := req.GetFilename()
	chunkIndex := req.GetChunkIndex()

	s.mu.RLock()
	defer s.mu.RUnlock()

	fileMeta, exists := s.fileMetadata[filename]
	if !exists {
		return &gfs.GetChunkLocationsResponse{
			ChunkserverAddresses: nil,
			ChunkHandle:          "",
		}, nil
	}

	if chunkIndex < 0 || chunkIndex >= int32(len(fileMeta.ChunkHandles)) {
		return &gfs.GetChunkLocationsResponse{
			ChunkserverAddresses: nil,
			ChunkHandle:          "",
		}, nil
	}

	chunkHandle := fileMeta.ChunkHandles[chunkIndex]
	locations := s.chunkLocations[chunkHandle]

	return &gfs.GetChunkLocationsResponse{
		ChunkserverAddresses: locations,
		ChunkHandle:          chunkHandle,
	}, nil
}
