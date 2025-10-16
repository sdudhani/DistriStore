package master

import (
	"context"
	"log"

	"github.com/sdudhani/godfs/pkg/gfs"
)

// Server implements the gRPC Master sever.
type Server struct {
	gfs.UnimplementedMasterServer
}

func (s *Server) Heartbeat(ctx context.Context, req *gfs.HeartbeatRequest) (*gfs.HeartbeatResponse, error) {
	chunkserverID := req.GetChunkserverId()
	log.Printf("Received heartbeat from :%s", chunkserverID)

	return &gfs.HeartbeatResponse{Message: "Heartbeat received"}, nil
}
