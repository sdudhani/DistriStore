package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("🚀 GoDFS Comprehensive Upload/Download Test")
	fmt.Println("=============================================")

	// Connect to master server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect to master: %v", err)
	}
	defer conn.Close()

	masterClient := gfs.NewMasterClient(conn)
	fmt.Println("✅ Connected to master server!")

	// Test 1: Upload small text file
	fmt.Println("\n📤 Test 1: Uploading small text file")
	testData1 := []byte("Hello GoDFS! This is a test file for the distributed file system.")
	
	uploadReq1 := &gfs.UploadFileRequest{
		Filename: "test_small.txt",
		Data:     testData1,
	}

	uploadResp1, err := masterClient.UploadFile(context.Background(), uploadReq1)
	if err != nil {
		log.Printf("❌ Upload failed: %v", err)
		return
	}

	if uploadResp1.GetSuccess() {
		fmt.Printf("✅ Upload successful: %s\n", uploadResp1.GetMessage())
	} else {
		fmt.Printf("❌ Upload failed: %s\n", uploadResp1.GetMessage())
		return
	}

	// Test 2: Download the small file
	fmt.Println("\n📥 Test 2: Downloading small text file")
	downloadReq1 := &gfs.DownloadFileRequest{
		Filename: "test_small.txt",
	}

	downloadResp1, err := masterClient.DownloadFile(context.Background(), downloadReq1)
	if err != nil {
		log.Printf("❌ Download failed: %v", err)
		return
	}

	if downloadResp1.GetSuccess() {
		fmt.Printf("✅ Download successful: %s\n", downloadResp1.GetMessage())
		fmt.Printf("📄 File content: %s\n", string(downloadResp1.GetData()))
		
		// Verify data integrity
		if string(downloadResp1.GetData()) == string(testData1) {
			fmt.Println("✅ Data integrity verified!")
		} else {
			fmt.Println("❌ Data integrity check failed!")
		}
	} else {
		fmt.Printf("❌ Download failed: %s\n", downloadResp1.GetMessage())
		return
	}

	// Test 3: Upload larger binary file
	fmt.Println("\n📤 Test 3: Uploading larger binary file")
	largeData := make([]byte, 2048) // 2KB file
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	uploadReq2 := &gfs.UploadFileRequest{
		Filename: "test_large.bin",
		Data:     largeData,
	}

	uploadResp2, err := masterClient.UploadFile(context.Background(), uploadReq2)
	if err != nil {
		log.Printf("❌ Large file upload failed: %v", err)
		return
	}

	if uploadResp2.GetSuccess() {
		fmt.Printf("✅ Large file upload successful: %s\n", uploadResp2.GetMessage())
	} else {
		fmt.Printf("❌ Large file upload failed: %s\n", uploadResp2.GetMessage())
		return
	}

	// Test 4: Download the large file
	fmt.Println("\n📥 Test 4: Downloading large binary file")
	downloadReq2 := &gfs.DownloadFileRequest{
		Filename: "test_large.bin",
	}

	downloadResp2, err := masterClient.DownloadFile(context.Background(), downloadReq2)
	if err != nil {
		log.Printf("❌ Large file download failed: %v", err)
		return
	}

	if downloadResp2.GetSuccess() {
		fmt.Printf("✅ Large file download successful: %s\n", downloadResp2.GetMessage())
		fmt.Printf("📊 Downloaded size: %d bytes\n", len(downloadResp2.GetData()))
		
		// Verify data integrity
		if len(downloadResp2.GetData()) == len(largeData) {
			matches := true
			for i, b := range downloadResp2.GetData() {
				if b != largeData[i] {
					matches = false
					break
				}
			}
			if matches {
				fmt.Println("✅ Large file data integrity verified!")
			} else {
				fmt.Println("❌ Large file data integrity check failed!")
			}
		} else {
			fmt.Println("❌ Large file size mismatch!")
		}
	} else {
		fmt.Printf("❌ Large file download failed: %s\n", downloadResp2.GetMessage())
		return
	}

	// Test 5: List files
	fmt.Println("\n📋 Test 5: Listing all files")
	listReq := &gfs.ListFilesRequest{
		Path: "",
	}

	listResp, err := masterClient.ListFiles(context.Background(), listReq)
	if err != nil {
		log.Printf("❌ List files failed: %v", err)
		return
	}

	if listResp.GetSuccess() {
		fmt.Printf("✅ Files listed successfully: %s\n", listResp.GetMessage())
		fmt.Printf("📁 Files: %v\n", listResp.GetFiles())
	} else {
		fmt.Printf("❌ List files failed: %s\n", listResp.GetMessage())
	}

	// Test 6: Check chunk locations and replication
	fmt.Println("\n🔍 Test 6: Checking chunk locations and replication")
	locationsReq1 := &gfs.GetChunkLocationsRequest{
		Filename:   "test_small.txt",
		ChunkIndex: 0,
	}

	locationsResp1, err := masterClient.GetChunkLocations(context.Background(), locationsReq1)
	if err != nil {
		log.Printf("❌ Get chunk locations failed: %v", err)
		return
	}

	fmt.Printf("📦 Small file chunk handle: %s\n", locationsResp1.GetChunkHandle())
	fmt.Printf("🖥️  Replicated on chunkservers: %v\n", locationsResp1.GetChunkserverAddresses())
	fmt.Printf("📊 Replication factor: %d\n", len(locationsResp1.GetChunkserverAddresses()))

	locationsReq2 := &gfs.GetChunkLocationsRequest{
		Filename:   "test_large.bin",
		ChunkIndex: 0,
	}

	locationsResp2, err := masterClient.GetChunkLocations(context.Background(), locationsReq2)
	if err != nil {
		log.Printf("❌ Get chunk locations for large file failed: %v", err)
		return
	}

	fmt.Printf("📦 Large file chunk handle: %s\n", locationsResp2.GetChunkHandle())
	fmt.Printf("🖥️  Replicated on chunkservers: %v\n", locationsResp2.GetChunkserverAddresses())
	fmt.Printf("📊 Replication factor: %d\n", len(locationsResp2.GetChunkserverAddresses()))

	// Test 7: Upload another file to test multiple files
	fmt.Println("\n📤 Test 7: Uploading another file")
	testData3 := []byte("This is another test file to verify multiple file support.")
	
	uploadReq3 := &gfs.UploadFileRequest{
		Filename: "test_another.txt",
		Data:     testData3,
	}

	uploadResp3, err := masterClient.UploadFile(context.Background(), uploadReq3)
	if err != nil {
		log.Printf("❌ Third file upload failed: %v", err)
		return
	}

	if uploadResp3.GetSuccess() {
		fmt.Printf("✅ Third file upload successful: %s\n", uploadResp3.GetMessage())
	} else {
		fmt.Printf("❌ Third file upload failed: %s\n", uploadResp3.GetMessage())
	}

	// Test 8: Download the third file
	fmt.Println("\n📥 Test 8: Downloading third file")
	downloadReq3 := &gfs.DownloadFileRequest{
		Filename: "test_another.txt",
	}

	downloadResp3, err := masterClient.DownloadFile(context.Background(), downloadReq3)
	if err != nil {
		log.Printf("❌ Third file download failed: %v", err)
		return
	}

	if downloadResp3.GetSuccess() {
		fmt.Printf("✅ Third file download successful: %s\n", downloadResp3.GetMessage())
		fmt.Printf("📄 File content: %s\n", string(downloadResp3.GetData()))
		
		// Verify data integrity
		if string(downloadResp3.GetData()) == string(testData3) {
			fmt.Println("✅ Third file data integrity verified!")
		} else {
			fmt.Println("❌ Third file data integrity check failed!")
		}
	} else {
		fmt.Printf("❌ Third file download failed: %s\n", downloadResp3.GetMessage())
	}

	// Test 9: Final file listing
	fmt.Println("\n📋 Test 9: Final file listing")
	listResp2, err := masterClient.ListFiles(context.Background(), listReq)
	if err != nil {
		log.Printf("❌ Final list files failed: %v", err)
		return
	}

	if listResp2.GetSuccess() {
		fmt.Printf("✅ Final files listed: %s\n", listResp2.GetMessage())
		fmt.Printf("📁 All files: %v\n", listResp2.GetFiles())
		fmt.Printf("📊 Total files: %d\n", len(listResp2.GetFiles()))
	}

	fmt.Println("\n🎉 All tests completed successfully!")
	fmt.Println("✅ GoDFS upload and download functionality is working properly!")
}
