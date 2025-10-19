package main

import (
	"log"
	"net"

	"github.com/sdudhani/godfs/internal/master"
	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
)

func main(){
	// Master listens on port 9000
	lis, err := net.Listen("tcp", ":9000") 	
	if err!= nil {
		log.Fatal("Failed to listen %v", err)
	}

	grpcServer := grpc.NewServer()

	masterServer := &master.Server{}

	gfs.RegisterMasterServer(grpcServer, masterServer)
	
	log.Println("Master server listening on Port 9000")

	if err:= grpcServer.Serve(lis); err  != nil {
		log.Fatal("Failed to serve: %v", err)
	} 
}