package handlers

import (
	"bytes" // Keep import, needed by imaging.Decode
	// Keep import, needed by imaging.Decode
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

func (app *App) AdminMediaManager(w http.ResponseWriter, r *http.Request) {
	// List files in uploads directory
	uploadsDir := app.Config.UploadPath
	optimizedDir := filepath.Join(app.Config.UploadPath, "optimized")

	// Ensure optimized directory exists
	if err := os.MkdirAll(optimizedDir, 0755); err != nil {
		slog.Error("Error creating optimized upload dir", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	files, err := os.ReadDir(uploadsDir)
	if err != nil {
		slog.Error("Error reading uploads dir", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var images []string
	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), "optimized_") { // Filter out optimized versions
			images = append(images, file.Name())
		}
	}

	// Also list optimized images
	optimizedFiles, err := os.ReadDir(optimizedDir)
	if err != nil {
		slog.Error("Error reading optimized upload dir", "error", err)
		// Don't fail if optimized dir is empty
	} else {
		for _, file := range optimizedFiles {
			if !file.IsDir() {
				images = append(images, filepath.Join("optimized", file.Name()))
			}
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

	file, _, err := r.FormFile("image") // Removed 'header'
	if err != nil {
		slog.Error("Error retrieving file", "error", err)
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
		return
	}
	defer file.Close()

	// Read file into a buffer to allow multiple reads
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Error reading file into buffer", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Validate content type (Magic Numbers)
	fileType := http.DetectContentType(fileBytes)

	originalExt := ""
	switch fileType {
	case "image/jpeg":
		originalExt = ".jpg"
	case "image/png":
		originalExt = ".png"
	case "image/gif":
		originalExt = ".gif"
	case "image/webp":
		originalExt = ".webp"
	default:
		http.Error(w, "Invalid file type. Only JPG, PNG, GIF, WEBP allowed.", http.StatusBadRequest)
		return
	}

	// Generate unique filename for original
	originalFilename := uuid.New().String() + originalExt
	originalFilePath := filepath.Join(app.Config.UploadPath, originalFilename)

	// Save original image
	dst, err := os.Create(originalFilePath)
	if err != nil {
		slog.Error("Error creating original file", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := dst.Write(fileBytes); err != nil {
		slog.Error("Error writing original file", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// --- Image Optimization ---
	img, err := imaging.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		slog.Error("Error decoding image for optimization", "error", err)
		// Continue without optimization if decode fails
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
		return
	}

	// Resize to max 1200px width, maintaining aspect ratio
	if img.Bounds().Dx() > 1200 {
		img = imaging.Resize(img, 1200, 0, imaging.Lanczos)
	}

	// Generate unique filename for optimized WebP
	optimizedFilename := "optimized_" + uuid.New().String() + ".webp"
	optimizedFilePath := filepath.Join(app.Config.UploadPath, "optimized", optimizedFilename)

	// Encode as WebP
	if err := imaging.Save(img, optimizedFilePath); err != nil {
		slog.Error("Error encoding optimized image to WebP", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// --- End Image Optimization ---

	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
}
