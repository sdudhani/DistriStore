package master

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
)

// FileMetadata represents metadata for a file
type FileMetadata struct {
	ChunkHandles []string
	Size         int64
}

// ChunkserverInfo represents information about a chunkserver
type ChunkserverInfo struct {
	Address   string
	LastSeen  int64
	IsHealthy bool
}

// Server implements the gRPC Master server
type Server struct {
	gfs.UnimplementedMasterServer

	// Metadata storage
	mu             sync.RWMutex
	fileMetadata   map[string]*FileMetadata // filename -> metadata
	chunkLocations map[string][]string      // chunkHandle -> chunkserver addresses

	// Chunkserver management
	chunkservers map[string]*ChunkserverInfo // address -> info
}

// NewServer creates a new master server
func NewServer() *Server {
	server := &Server{
		fileMetadata:   make(map[string]*FileMetadata),
		chunkLocations: make(map[string][]string),
		chunkservers:   make(map[string]*ChunkserverInfo),
	}

	// Start health monitoring
	go server.monitorChunkserverHealth()

	return server
}

// getChunkserverClient creates a connection to the chunkserver
func (s *Server) getChunkserverClient(address string) (gfs.ChunkserverClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure()) // TODO: Replace with secure connection in production
	if err != nil {
		return nil, err
	}
	return gfs.NewChunkserverClient(conn), nil
}

// Heartbeat handles chunkserver heartbeats
func (s *Server) Heartbeat(ctx context.Context, req *gfs.HeartbeatRequest) (*gfs.HeartbeatResponse, error) {
	chunkserverID := req.GetChunkserverId()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Register or update chunkserver info
	// For now, we'll assume the chunkserverID is the address
	// In a real system, you'd have a mapping from ID to address
	s.chunkservers[chunkserverID] = &ChunkserverInfo{
		Address:   chunkserverID,
		LastSeen:  time.Now().Unix(),
		IsHealthy: true,
	}

	log.Printf("Received heartbeat from: %s", chunkserverID)

	return &gfs.HeartbeatResponse{Message: "Heartbeat received"}, nil
}

// getAvailableChunkservers returns a list of healthy chunkservers
func (s *Server) getAvailableChunkservers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var available []string
	now := time.Now().Unix()

	for address, info := range s.chunkservers {
		// Check if chunkserver is healthy and recently seen (within 30 seconds)
		if info.IsHealthy && (now-info.LastSeen) < 30 {
			available = append(available, address)
		}
	}

	return available
}

// UploadFile handles file uploads with replication
func (s *Server) UploadFile(ctx context.Context, req *gfs.UploadFileRequest) (*gfs.UploadFileResponse, error) {
	filename := req.GetFilename()
	data := req.GetData()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a chunk handle for this file
	chunkHandle := fmt.Sprintf("%s-0", filename)

	// Get available chunkservers
	availableChunkservers := s.getAvailableChunkservers()

	// Check if we have any chunkservers available
	if len(availableChunkservers) == 0 {
		return &gfs.UploadFileResponse{
			Success: false,
			Message: "No chunkservers available",
		}, nil
	}

	// Replicate chunk to multiple chunkservers (3 replicas)
	replicaCount := 3
	if len(availableChunkservers) < replicaCount {
		replicaCount = len(availableChunkservers)
	}

	var successfulReplicas []string

	for i := 0; i < replicaCount; i++ {
		chunkserverAddr := availableChunkservers[i]

		// Get chunkserver client
		chunkserverClient, err := s.getChunkserverClient(chunkserverAddr)
		if err != nil {
			log.Printf("Failed to connect to chunkserver %s: %v", chunkserverAddr, err)
			continue
		}

		// Store the chunk on chunkserver
		storeReq := &gfs.StoreChunkRequest{
			ChunkHandle: chunkHandle,
			Data:        data,
		}

		_, err = chunkserverClient.StoreChunk(ctx, storeReq)
		if err != nil {
			log.Printf("Failed to store chunk %s on %s: %v", chunkHandle, chunkserverAddr, err)
			continue
		}

		successfulReplicas = append(successfulReplicas, chunkserverAddr)
		log.Printf("Stored chunk %s on %s", chunkHandle, chunkserverAddr)
	}

	if len(successfulReplicas) == 0 {
		return &gfs.UploadFileResponse{
			Success: false,
			Message: "Failed to store chunk on any chunkserver",
		}, nil
	}

	// Update metadata
	s.fileMetadata[filename] = &FileMetadata{
		ChunkHandles: []string{chunkHandle},
		Size:         int64(len(data)),
	}

	// Record chunk locations
	s.chunkLocations[chunkHandle] = successfulReplicas

	log.Printf("Uploaded file %s (%d bytes) as chunk %s with %d replicas",
		filename, len(data), chunkHandle, len(successfulReplicas))

	return &gfs.UploadFileResponse{
		Success: true,
		Message: fmt.Sprintf("File uploaded successfully with %d replicas", len(successfulReplicas)),
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

	// Get chunk locations
	s.mu.RLock()
	locations := s.chunkLocations[chunkHandle]
	s.mu.RUnlock()

	// Try to retrieve from any available replica
	for _, chunkserverAddr := range locations {
		chunkserverClient, err := s.getChunkserverClient(chunkserverAddr)
		if err != nil {
			log.Printf("Failed to connect to chunkserver %s: %v", chunkserverAddr, err)
			continue
		}

		retrieveReq := &gfs.RetrieveChunkRequest{
			ChunkHandle: chunkHandle,
		}

		resp, err := chunkserverClient.RetrieveChunk(ctx, retrieveReq)
		if err != nil {
			log.Printf("Failed to retrieve chunk %s from %s: %v", chunkHandle, chunkserverAddr, err)
			continue
		}

		if !resp.GetSuccess() {
			log.Printf("Chunkserver %s returned error for chunk %s: %s", chunkserverAddr, chunkHandle, resp.GetMessage())
			continue
		}

		log.Printf("Downloaded file %s (%d bytes) from %s", filename, len(resp.GetData()), chunkserverAddr)

		return &gfs.DownloadFileResponse{
			Success: true,
			Data:    resp.GetData(),
			Message: "File downloaded successfully",
		}, nil
	}

	return &gfs.DownloadFileResponse{
		Success: false,
		Data:    nil,
		Message: "Failed to retrieve chunk from any chunkserver",
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
		if path == "" || len(filename) >= len(path) && filename[:len(path)] == path {
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

	// Delete chunks from all replicas
	for _, chunkHandle := range fileMeta.ChunkHandles {
		locations := s.chunkLocations[chunkHandle]

		for _, chunkserverAddr := range locations {
			chunkserverClient, err := s.getChunkserverClient(chunkserverAddr)
			if err != nil {
				log.Printf("Failed to connect to chunkserver %s: %v", chunkserverAddr, err)
				continue
			}

			deleteReq := &gfs.DeleteChunkRequest{
				ChunkHandle: chunkHandle,
			}

			_, err = chunkserverClient.DeleteChunk(ctx, deleteReq)
			if err != nil {
				log.Printf("Failed to delete chunk %s from %s: %v", chunkHandle, chunkserverAddr, err)
			}
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

// monitorChunkserverHealth monitors chunkserver health and handles re-replication
func (s *Server) monitorChunkserverHealth() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.checkAndHandleFailedChunkservers()
	}
}

// checkAndHandleFailedChunkservers checks for failed chunkservers and handles re-replication
func (s *Server) checkAndHandleFailedChunkservers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	var failedChunkservers []string

	// Identify failed chunkservers (no heartbeat for 60 seconds)
	for address, info := range s.chunkservers {
		if (now - info.LastSeen) > 60 {
			info.IsHealthy = false
			failedChunkservers = append(failedChunkservers, address)
			log.Printf("Chunkserver %s marked as failed (last seen: %d seconds ago)", address, now-info.LastSeen)
		}
	}

	// Handle re-replication for chunks on failed chunkservers
	for _, failedAddr := range failedChunkservers {
		s.handleChunkserverFailure(failedAddr)
	}
}

// handleChunkserverFailure handles re-replication when a chunkserver fails
func (s *Server) handleChunkserverFailure(failedAddr string) {
	log.Printf("Handling failure of chunkserver %s", failedAddr)

	// Find chunks that were stored on the failed chunkserver
	for chunkHandle, locations := range s.chunkLocations {
		// Check if this chunk was stored on the failed chunkserver
		hasFailedReplica := false
		for _, addr := range locations {
			if addr == failedAddr {
				hasFailedReplica = true
				break
			}
		}

		if hasFailedReplica {
			// Remove failed chunkserver from locations
			var newLocations []string
			for _, addr := range locations {
				if addr != failedAddr {
					newLocations = append(newLocations, addr)
				}
			}
			s.chunkLocations[chunkHandle] = newLocations

			// Try to re-replicate if we have at least one good replica
			if len(newLocations) > 0 {
				s.replicateChunk(chunkHandle, newLocations[0])
			}
		}
	}
}

// replicateChunk replicates a chunk from a source chunkserver to available chunkservers
func (s *Server) replicateChunk(chunkHandle, sourceAddr string) {
	// Get available chunkservers (excluding the source)
	availableChunkservers := s.getAvailableChunkservers()
	var targetChunkservers []string

	for _, addr := range availableChunkservers {
		if addr != sourceAddr {
			targetChunkservers = append(targetChunkservers, addr)
		}
	}

	if len(targetChunkservers) == 0 {
		log.Printf("No available chunkservers for re-replication of chunk %s", chunkHandle)
		return
	}

	// Get the chunk data from the source
	sourceClient, err := s.getChunkserverClient(sourceAddr)
	if err != nil {
		log.Printf("Failed to connect to source chunkserver %s: %v", sourceAddr, err)
		return
	}

	ctx := context.Background()
	retrieveReq := &gfs.RetrieveChunkRequest{ChunkHandle: chunkHandle}
	resp, err := sourceClient.RetrieveChunk(ctx, retrieveReq)
	if err != nil || !resp.GetSuccess() {
		log.Printf("Failed to retrieve chunk %s from source %s: %v", chunkHandle, sourceAddr, err)
		return
	}

	// Replicate to target chunkservers
	var successfulReplicas []string
	for _, targetAddr := range targetChunkservers {
		targetClient, err := s.getChunkserverClient(targetAddr)
		if err != nil {
			log.Printf("Failed to connect to target chunkserver %s: %v", targetAddr, err)
			continue
		}

		storeReq := &gfs.StoreChunkRequest{
			ChunkHandle: chunkHandle,
			Data:        resp.GetData(),
		}

		_, err = targetClient.StoreChunk(ctx, storeReq)
		if err != nil {
			log.Printf("Failed to store chunk %s on target %s: %v", chunkHandle, targetAddr, err)
			continue
		}

		successfulReplicas = append(successfulReplicas, targetAddr)
		log.Printf("Re-replicated chunk %s to %s", chunkHandle, targetAddr)
	}

	// Update chunk locations
	if len(successfulReplicas) > 0 {
		s.mu.Lock()
		s.chunkLocations[chunkHandle] = append(s.chunkLocations[chunkHandle], successfulReplicas...)
		s.mu.Unlock()
	}
}
