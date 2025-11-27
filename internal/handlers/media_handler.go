package handlers

import (
	"io"
	"log"
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
		log.Printf("Error reading uploads dir: %v", err)
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
		log.Printf("Error retrieving file: %v", err)
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
		return
	}
	defer file.Close()

	// Validate extension
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		http.Error(w, "Invalid file type. Only JPG, PNG, GIF, WEBP allowed.", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	filename := uuid.New().String() + ext
	filePath := filepath.Join("web/static/uploads", filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy contents
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error copying file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
}
