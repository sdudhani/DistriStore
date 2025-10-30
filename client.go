package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Connecting to GoDFS master server...")

	// Connect to master server
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to master: %v", err)
	}
	defer conn.Close()

	masterClient := gfs.NewMasterClient(conn)
	fmt.Println("Connected to master server!")

	// Test upload
	fmt.Println("\n Uploading test file...")
	testData := []byte("Hello GoDFS! This is a replication test.")

	uploadReq := &gfs.UploadFileRequest{
		Filename: "test.txt",
		Data:     testData,
	}

	uploadResp, err := masterClient.UploadFile(context.Background(), uploadReq)
	if err != nil {
		log.Printf("Upload failed: %v", err)
		return
	}

	fmt.Printf("Upload result: %s\n", uploadResp.GetMessage())

	// Test download
	fmt.Println("\nDownloading test file...")
	downloadReq := &gfs.DownloadFileRequest{
		Filename: "test.txt",
	}

	downloadResp, err := masterClient.DownloadFile(context.Background(), downloadReq)
	if err != nil {
		log.Printf("Download failed: %v", err)
		return
	}

	fmt.Printf("✅ Download result: %s\n", downloadResp.GetMessage())
	fmt.Printf("📄 File content: %s\n", string(downloadResp.GetData()))

	// Check replication
	fmt.Println("\n Checking chunk locations...")
	locationsReq := &gfs.GetChunkLocationsRequest{
		Filename:   "test.txt",
		ChunkIndex: 0,
	}

	locationsResp, err := masterClient.GetChunkLocations(context.Background(), locationsReq)
	if err != nil {
		log.Printf("Get locations failed: %v", err)
		return
	}

	fmt.Printf("Chunk handle: %s\n", locationsResp.GetChunkHandle())
	fmt.Printf("Replicated on chunkservers: %v\n", locationsResp.GetChunkserverAddresses())
	fmt.Printf("Replication factor: %d\n", len(locationsResp.GetChunkserverAddresses()))

	fmt.Println("\nTest completed successfully!")
}
