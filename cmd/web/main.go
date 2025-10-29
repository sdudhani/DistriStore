package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sdudhani/godfs/pkg/gfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	masterAddr string
}

func (s *server) masterClient(ctx context.Context) (gfs.MasterClient, *grpc.ClientConn, error) {
	dctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(dctx, s.masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return gfs.NewMasterClient(conn), conn, nil
}

var tmpl = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>GoDFS – Distributed File System</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif; margin: 2rem; background: #f9fafb; }
    .container { max-width: 1200px; margin: 0 auto; }
    h1 { margin-bottom: 0.25rem; color: #1f2937; }
    .card { background: white; border: 1px solid #e5e7eb; border-radius: 12px; padding: 1.5rem; margin: 1rem 0; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
    .files li { margin: 0.5rem 0; padding: 0.5rem; background: #f3f4f6; border-radius: 6px; display: flex; justify-content: space-between; align-items: center; }
    .btn { background: #2563eb; color: white; border: 0; padding: 0.5rem 1rem; border-radius: 6px; cursor: pointer; text-decoration: none; display: inline-block; }
    .btn:hover { background: #1d4ed8; }
    .btn:disabled { background: #9ca3af; }
    .btn-secondary { background: #6b7280; }
    .btn-secondary:hover { background: #4b5563; }
    input[type=file] { margin: 0.5rem 0; padding: 0.5rem; border: 1px solid #d1d5db; border-radius: 6px; width: 100%; }
    .muted { color: #6b7280; }
    .status-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin: 1rem 0; }
    .status-card { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 1rem; text-align: center; }
    .status-online { color: #059669; font-weight: bold; }
    .status-offline { color: #dc2626; font-weight: bold; }
    .replication-info { background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 6px; padding: 0.75rem; margin: 0.5rem 0; }
    .file-info { display: flex; justify-content: space-between; align-items: center; }
    .file-name { font-family: monospace; font-weight: bold; }
    .replication-badge { background: #dbeafe; color: #1e40af; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.875rem; }
    .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; }
    .refresh-btn { background: #10b981; }
    .refresh-btn:hover { background: #059669; }
  </style>
  </head>
<body>
  <div class="container">
    <div class="header">
      <div>
        <h1>🚀 GoDFS - Distributed File System</h1>
        <p class="muted">Upload files with automatic replication across multiple chunkservers</p>
      </div>
      <a href="/" class="btn refresh-btn">🔄 Refresh</a>
    </div>

    <div class="status-grid">
      <div class="status-card">
        <h3>Master Server</h3>
        <p class="status-online">✅ Online</p>
        <p class="muted">{{.MasterAddr}}</p>
      </div>
      {{range .Chunkservers}}
      <div class="status-card">
        <h3>Chunkserver {{.Port}}</h3>
        <p class="{{if .Healthy}}status-online{{else}}status-offline{{end}}">
          {{if .Healthy}}✅ Online{{else}}❌ Offline{{end}}
        </p>
        <p class="muted">{{.Address}}</p>
      </div>
      {{end}}
    </div>

    <div class="card">
      <h2>📤 Upload File</h2>
      <form action="/upload" method="post" enctype="multipart/form-data">
        <input type="file" name="file" required />
        <div>
          <button class="btn" type="submit">Upload with Replication</button>
        </div>
      </form>
      <div class="replication-info">
        <strong>🔄 Replication:</strong> Files are automatically replicated across {{.ReplicationFactor}} chunkservers for fault tolerance
      </div>
    </div>

    <div class="card">
      <h2>📁 Files ({{len .Files}})</h2>
      {{if .Files}}
        <ul class="files" style="list-style: none; padding: 0;">
        {{range .Files}}
          <li>
            <div class="file-info">
              <div>
                <div class="file-name">{{.Name}}</div>
                <div class="muted">{{.Size}} bytes • {{.Replicas}} replicas</div>
              </div>
              <div>
                <span class="replication-badge">{{.Replicas}}x replicated</span>
                <a class="btn" href="/download?filename={{.Name}}">Download</a>
              </div>
            </div>
          </li>
        {{end}}
        </ul>
      {{else}}
        <p class="muted">No files found. Upload a file to see replication in action!</p>
      {{end}}
    </div>

    {{if .Flash}}
      <div class="card" style="background: #f0f9ff; border-color: #0ea5e9;">
        <p style="margin: 0; color: #0c4a6e;">{{.Flash}}</p>
      </div>
    {{end}}
  </div>
</body>
</html>`))

type FileInfo struct {
	Name     string
	Size     int64
	Replicas int
}

type ChunkserverStatus struct {
	Address string
	Port    string
	Healthy bool
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	client, conn, err := s.masterClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to connect to master: %v", err), http.StatusBadGateway)
		return
	}
	defer conn.Close()

	// Get file list
	resp, err := client.ListFiles(ctx, &gfs.ListFilesRequest{Path: ""})
	if err != nil {
		http.Error(w, fmt.Sprintf("list files failed: %v", err), http.StatusBadGateway)
		return
	}

	// Get file details with replication info
	var files []FileInfo
	for _, filename := range resp.GetFiles() {
		// Get chunk locations to determine replication
		locationsResp, err := client.GetChunkLocations(ctx, &gfs.GetChunkLocationsRequest{
			Filename:   filename,
			ChunkIndex: 0,
		})

		replicas := 0
		if err == nil {
			replicas = len(locationsResp.GetChunkserverAddresses())
		}

		// Get file size (simplified - in real system, this would be in metadata)
		files = append(files, FileInfo{
			Name:     filename,
			Size:     1024, // Placeholder - would get from metadata
			Replicas: replicas,
		})
	}

	// Mock chunkserver status (in real system, this would come from master)
	chunkservers := []ChunkserverStatus{
		{Address: "localhost:9001", Port: "9001", Healthy: true},
		{Address: "localhost:9002", Port: "9002", Healthy: true},
		{Address: "localhost:9003", Port: "9003", Healthy: true},
	}

	data := struct {
		Files             []FileInfo
		Chunkservers      []ChunkserverStatus
		ReplicationFactor int
		Flash             string
		MasterAddr        string
	}{
		Files:             files,
		Chunkservers:      chunkservers,
		ReplicationFactor: 3,
		Flash:             r.URL.Query().Get("flash"),
		MasterAddr:        s.masterAddr,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("template execute error: %v", err)
	}
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("invalid form: %v", err), http.StatusBadRequest)
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("file missing: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file fully (simple demo; for large files, chunking would be better)
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("read failed: %v", err), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	client, conn, err := s.masterClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to connect to master: %v", err), http.StatusBadGateway)
		return
	}
	defer conn.Close()

	_, err = client.UploadFile(ctx, &gfs.UploadFileRequest{Filename: hdr.Filename, Data: data})
	if err != nil {
		http.Error(w, fmt.Sprintf("upload failed: %v", err), http.StatusBadGateway)
		return
	}

	http.Redirect(w, r, "/?flash="+template.URLQueryEscaper("Uploaded "+hdr.Filename), http.StatusSeeOther)
}

func (s *server) handleDownload(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "filename is required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	client, conn, err := s.masterClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to connect to master: %v", err), http.StatusBadGateway)
		return
	}
	defer conn.Close()

	resp, err := client.DownloadFile(ctx, &gfs.DownloadFileRequest{Filename: filename})
	if err != nil || !resp.GetSuccess() {
		http.Error(w, fmt.Sprintf("download failed: %v %s", err, resp.GetMessage()), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp.GetData())
}

func main() {
	masterAddr := os.Getenv("MASTER_ADDR")
	if masterAddr == "" {
		masterAddr = "localhost:9000"
	}
	s := &server{masterAddr: masterAddr}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/upload", s.handleUpload)
	mux.HandleFunc("/download", s.handleDownload)

	addr := ":8080"
	log.Printf("GoDFS Web listening on %s (MASTER_ADDR=%s)", addr, masterAddr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}
