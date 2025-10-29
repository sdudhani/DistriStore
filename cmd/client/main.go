package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("🚀 GoDFS Client")
	fmt.Println("===============")
	fmt.Println("A Distributed File System Client")
	fmt.Println("")

	// Connect to master server
	conn, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Failed to connect to master: %v", err)
	}
	defer conn.Close()

	client := gfs.NewMasterClient(conn)
	fmt.Println("✅ Connected to GoDFS master server!")
	fmt.Println("")

	// Interactive menu
	scanner := bufio.NewScanner(os.Stdin)

	for {
		showMenu()
		fmt.Print("Enter your choice: ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			uploadFile(client, scanner)
		case "2":
			downloadFile(client, scanner)
		case "3":
			listFiles(client)
		case "4":
			showSystemStatus(client)
		case "5":
			showHelp()
		case "6", "q", "quit", "exit":
			fmt.Println("👋 Goodbye!")
			return
		default:
			fmt.Println("❌ Invalid choice. Please try again.")
		}

		fmt.Println("")
	}
}

func showMenu() {
	fmt.Println("📋 GoDFS Client Menu:")
	fmt.Println("1. 📤 Upload File")
	fmt.Println("2. 📥 Download File")
	fmt.Println("3. 📁 List Files")
	fmt.Println("4. 🔍 System Status")
	fmt.Println("5. ❓ Help")
	fmt.Println("6. 🚪 Exit")
	fmt.Println("")
}

func uploadFile(client gfs.MasterClient, scanner *bufio.Scanner) {
	fmt.Print("Enter filename: ")
	if !scanner.Scan() {
		return
	}
	filename := strings.TrimSpace(scanner.Text())

	if filename == "" {
		fmt.Println("❌ Filename cannot be empty")
		return
	}

	fmt.Print("Enter file content: ")
	if !scanner.Scan() {
		return
	}
	content := scanner.Text()

	fmt.Println("📤 Uploading file...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &gfs.UploadFileRequest{
		Filename: filename,
		Data:     []byte(content),
	}

	resp, err := client.UploadFile(ctx, req)
	if err != nil {
		fmt.Printf("❌ Upload failed: %v\n", err)
		return
	}

	if resp.GetSuccess() {
		fmt.Printf("✅ Upload successful: %s\n", resp.GetMessage())
	} else {
		fmt.Printf("❌ Upload failed: %s\n", resp.GetMessage())
	}
}

func downloadFile(client gfs.MasterClient, scanner *bufio.Scanner) {
	fmt.Print("Enter filename to download: ")
	if !scanner.Scan() {
		return
	}
	filename := strings.TrimSpace(scanner.Text())

	if filename == "" {
		fmt.Println("❌ Filename cannot be empty")
		return
	}

	fmt.Println("📥 Downloading file...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &gfs.DownloadFileRequest{
		Filename: filename,
	}

	resp, err := client.DownloadFile(ctx, req)
	if err != nil {
		fmt.Printf("❌ Download failed: %v\n", err)
		return
	}

	if resp.GetSuccess() {
		fmt.Printf("✅ Download successful: %s\n", resp.GetMessage())
		fmt.Printf("📄 File content: %s\n", string(resp.GetData()))
	} else {
		fmt.Printf("❌ Download failed: %s\n", resp.GetMessage())
	}
}

func listFiles(client gfs.MasterClient) {
	fmt.Println("📁 Listing files...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &gfs.ListFilesRequest{
		Path: "",
	}

	resp, err := client.ListFiles(ctx, req)
	if err != nil {
		fmt.Printf("❌ List files failed: %v\n", err)
		return
	}

	if resp.GetSuccess() {
		files := resp.GetFiles()
		if len(files) == 0 {
			fmt.Println("📁 No files found")
		} else {
			fmt.Printf("📁 Found %d files:\n", len(files))
			for i, file := range files {
				fmt.Printf("  %d. %s\n", i+1, file)
			}
		}
	} else {
		fmt.Printf("❌ List files failed: %s\n", resp.GetMessage())
	}
}

func showSystemStatus(client gfs.MasterClient) {
	fmt.Println("🔍 Checking system status...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// List files to check system health
	req := &gfs.ListFilesRequest{
		Path: "",
	}

	resp, err := client.ListFiles(ctx, req)
	if err != nil {
		fmt.Printf("❌ System check failed: %v\n", err)
		return
	}

	if resp.GetSuccess() {
		files := resp.GetFiles()
		fmt.Println("✅ GoDFS System Status:")
		fmt.Printf("  📊 Total files: %d\n", len(files))
		fmt.Printf("  🖥️  Master server: Connected\n")
		fmt.Printf("  📁 Files: %v\n", files)
	} else {
		fmt.Printf("❌ System check failed: %s\n", resp.GetMessage())
	}
}

func showHelp() {
	fmt.Println("❓ GoDFS Client Help:")
	fmt.Println("")
	fmt.Println("📤 Upload File:")
	fmt.Println("  - Enter a filename and content")
	fmt.Println("  - File will be stored with replication")
	fmt.Println("")
	fmt.Println("📥 Download File:")
	fmt.Println("  - Enter a filename to download")
	fmt.Println("  - File content will be retrieved from chunkservers")
	fmt.Println("")
	fmt.Println("📁 List Files:")
	fmt.Println("  - Shows all files in the system")
	fmt.Println("")
	fmt.Println("🔍 System Status:")
	fmt.Println("  - Shows system health and file count")
	fmt.Println("")
	fmt.Println("💡 Tips:")
	fmt.Println("  - Make sure master server is running on port 9000")
	fmt.Println("  - Make sure at least one chunkserver is running")
	fmt.Println("  - Files are automatically replicated for fault tolerance")
}
