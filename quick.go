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
	fmt.Println("🔍 Quick GoDFS Test")
	
	// Connect with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	conn, err := grpc.DialContext(ctx, "localhost:9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer conn.Close()

	masterClient := gfs.NewMasterClient(conn)
	fmt.Println("✅ Connected to master!")

	// Test upload with timeout
	fmt.Println("📤 Testing upload...")
	uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer uploadCancel()
	
	uploadReq := &gfs.UploadFileRequest{
		Filename: "quick_test.txt",
		Data:     []byte("Hello GoDFS!"),
	}

	uploadResp, err := masterClient.UploadFile(uploadCtx, uploadReq)
	if err != nil {
		log.Printf("❌ Upload failed: %v", err)
		return
	}
	
	fmt.Printf("✅ Upload result: %s\n", uploadResp.GetMessage())
}
