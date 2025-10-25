package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
)

func main() {
	// Connect to master server
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to master: %v", err)
	}
	defer conn.Close()

	masterClient := gfs.NewMasterClient(conn)

	// Test 1: Upload a small text file
	fmt.Println("=== Test 1: Uploading small text file ===")
	testData := []byte("Hello, GoDFS! This is a test file for replication.")

	uploadReq := &gfs.UploadFileRequest{
		Filename: "test1.txt",
		Data:     testData,
	}

	uploadResp, err := masterClient.UploadFile(context.Background(), uploadReq)
	if err != nil {
		log.Printf("Upload failed: %v", err)
	} else {
		fmt.Printf("Upload result: Success=%v, Message=%s\n", uploadResp.GetSuccess(), uploadResp.GetMessage())
	}

	// Test 2: Download the file
	fmt.Println("\n=== Test 2: Downloading file ===")
	downloadReq := &gfs.DownloadFileRequest{
		Filename: "test1.txt",
	}

	downloadResp, err := masterClient.DownloadFile(context.Background(), downloadReq)
	if err != nil {
		log.Printf("Download failed: %v", err)
	} else {
		fmt.Printf("Download result: Success=%v, Message=%s\n", downloadResp.GetSuccess(), downloadResp.GetMessage())
		fmt.Printf("Downloaded data: %s\n", string(downloadResp.GetData()))
	}

	// Test 3: List files
	fmt.Println("\n=== Test 3: Listing files ===")
	listReq := &gfs.ListFilesRequest{
		Path: "",
	}

	listResp, err := masterClient.ListFiles(context.Background(), listReq)
	if err != nil {
		log.Printf("List files failed: %v", err)
	} else {
		fmt.Printf("List files result: Success=%v, Message=%s\n", listResp.GetSuccess(), listResp.GetMessage())
		fmt.Printf("Files: %v\n", listResp.GetFiles())
	}

	// Test 4: Get chunk locations
	fmt.Println("\n=== Test 4: Getting chunk locations ===")
	locationsReq := &gfs.GetChunkLocationsRequest{
		Filename:   "test1.txt",
		ChunkIndex: 0,
	}

	locationsResp, err := masterClient.GetChunkLocations(context.Background(), locationsReq)
	if err != nil {
		log.Printf("Get chunk locations failed: %v", err)
	} else {
		fmt.Printf("Chunk handle: %s\n", locationsResp.GetChunkHandle())
		fmt.Printf("Chunkserver addresses: %v\n", locationsResp.GetChunkserverAddresses())
	}

	// Test 5: Upload a larger file
	fmt.Println("\n=== Test 5: Uploading larger file ===")
	largeData := make([]byte, 1024) // 1KB file
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	largeUploadReq := &gfs.UploadFileRequest{
		Filename: "large_test.bin",
		Data:     largeData,
	}

	largeUploadResp, err := masterClient.UploadFile(context.Background(), largeUploadReq)
	if err != nil {
		log.Printf("Large file upload failed: %v", err)
	} else {
		fmt.Printf("Large file upload result: Success=%v, Message=%s\n", largeUploadResp.GetSuccess(), largeUploadResp.GetMessage())
	}

	// Test 6: List files again
	fmt.Println("\n=== Test 6: Listing files after second upload ===")
	listResp2, err := masterClient.ListFiles(context.Background(), listReq)
	if err != nil {
		log.Printf("List files failed: %v", err)
	} else {
		fmt.Printf("Files after second upload: %v\n", listResp2.GetFiles())
	}

	fmt.Println("\n=== All tests completed! ===")
}
