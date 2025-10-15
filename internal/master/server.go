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

func (s *Server) Hearbeat(ctx context.Context, req *gfs.HearbeatRequest) (*gfs.HeartbeatResponse, error) {
	chunkserverID := req.GetChunkserverID()
	log.Printf("Recieved heartbeat from :%s", chunkserverID)

	return &gfs.HeartbeatResponse{Message: "Hearbet recieved"}, nil
}
