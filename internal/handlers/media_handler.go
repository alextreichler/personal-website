package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func (app *App) AdminMediaManager(w http.ResponseWriter, r *http.Request) {
	// List files in uploads directory
	files, err := os.ReadDir("web/static/uploads")
	if err != nil {
		slog.Error("Error reading uploads dir", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var images []string
	for _, file := range files {
		if !file.IsDir() {
			images = append(images, file.Name())
		}
	}

	data := map[string]interface{}{
		"Images": images,
	}

	app.Render(w, r, "admin_media.html", data)
}

func (app *App) AdminUploadImage(w http.ResponseWriter, r *http.Request) {
	// Limit upload size to 10MB
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("image")
	if err != nil {
		slog.Error("Error retrieving file", "error", err)
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
		return
	}
	defer file.Close()

	// Validate content type (Magic Numbers)
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		slog.Error("Error reading file header", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fileType := http.DetectContentType(buff)
	if _, err := file.Seek(0, 0); err != nil {
		slog.Error("Error resetting file pointer", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	allowedTypes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}

	ext, allowed := allowedTypes[fileType]
	if !allowed {
		http.Error(w, "Invalid file type. Only JPG, PNG, GIF, WEBP allowed.", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	filename := uuid.New().String() + ext
	filePath := filepath.Join("web/static/uploads", filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		slog.Error("Error creating file", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy contents
	if _, err := io.Copy(dst, file); err != nil {
		slog.Error("Error copying file", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
}
