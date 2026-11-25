package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	filesDir  = "./files"
	maxUpload = 100 << 20 // 100 MB
)

type FileInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// Generate a random 128-bit file ID.
func generateFileID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// Ensure storage directory exists.
func ensureFilesDir() error {
	return os.MkdirAll(filesDir, 0750)
}

// GET /files
// Return list of available files.
func (h *Handler) listFiles(w http.ResponseWriter, r *http.Request) {
	if err := ensureFilesDir(); err != nil {
		http.Error(w, "Storage error", http.StatusInternalServerError)
		return
	}

	entries, err := os.ReadDir(filesDir)
	if err != nil {
		http.Error(w, "Failed to read files", http.StatusInternalServerError)
		return
	}

	files := make([]FileInfo, 0)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		metaPath := filepath.Join(filesDir, entry.Name())

		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		var file FileInfo

		if err := json.Unmarshal(data, &file); err != nil {
			continue
		}

		// Only include entries with an existing data file.
		if _, err := os.Stat(filepath.Join(filesDir, file.ID+".bin")); err != nil {
			continue
		}

		files = append(files, file)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// POST /files/upload
// Upload a new file.
func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	if err := ensureFilesDir(); err != nil {
		http.Error(w, "Storage error", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid upload or file too large", http.StatusBadRequest)
		return
	}

	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	src, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file field", http.StatusBadRequest)
		return
	}
	defer src.Close()

	// Generate unique ID.
	id, err := generateFileID()
	if err != nil {
		http.Error(w, "Failed to generate file ID", http.StatusInternalServerError)
		return
	}

	// Save uploaded file.
	filePath := filepath.Join(filesDir, id+".bin")

	dst, err := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0600,
	)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}

	size, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()

	if copyErr != nil || closeErr != nil {
		os.Remove(filePath)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Preserve original filename as metadata.
	name := filepath.Base(strings.ReplaceAll(header.Filename, "\\", "/"))

	if name == "." || name == "/" || name == "" {
		os.Remove(filePath)
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	file := FileInfo{
		ID:   id,
		Name: name,
		Size: size,
	}

	metaData, err := json.Marshal(file)
	if err != nil {
		os.Remove(filePath)
		http.Error(w, "Metadata error", http.StatusInternalServerError)
		return
	}

	metaPath := filepath.Join(filesDir, id+".json")

	if err := os.WriteFile(metaPath, metaData, 0600); err != nil {
		os.Remove(filePath)
		http.Error(w, "Failed to save metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(file)
}

// GET /files/{id}
// Download a file by its ID.
func (h *Handler) downloadFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Validate ID: 32 hexadecimal characters.
	if len(id) != 32 {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	if _, err := hex.DecodeString(id); err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	metaPath := filepath.Join(filesDir, id+".json")
	filePath := filepath.Join(filesDir, id+".bin")

	// Read metadata.
	data, err := os.ReadFile(metaPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	var file FileInfo

	if err := json.Unmarshal(data, &file); err != nil || file.ID != id {
		http.Error(w, "Invalid file metadata", http.StatusInternalServerError)
		return
	}

	src, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	w.Header().Set(
		"Content-Disposition",
		mime.FormatMediaType("attachment", map[string]string{
			"filename": file.Name,
		}),
	)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	http.ServeContent(w, r, file.Name, info.ModTime(), src)
}
