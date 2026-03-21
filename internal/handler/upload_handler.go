package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const maxUploadSize = 5 << 20 // 5MB

var allowedMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

func UploadHandler(uploadDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		file, header, err := r.FormFile("file")
		if err != nil {
			jsonError(w, "file is required (max 5MB)", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Detect MIME type from first 512 bytes
		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			jsonError(w, "failed to read file", http.StatusBadRequest)
			return
		}
		mime := http.DetectContentType(buf[:n])
		if !allowedMIME[mime] {
			jsonError(w, fmt.Sprintf("unsupported file type: %s", mime), http.StatusBadRequest)
			return
		}
		// Seek back to start
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		ext := filepath.Ext(header.Filename)
		if ext == "" {
			// Fallback extension from MIME
			switch mime {
			case "image/jpeg":
				ext = ".jpg"
			case "image/png":
				ext = ".png"
			case "image/gif":
				ext = ".gif"
			case "image/webp":
				ext = ".webp"
			}
		}
		ext = strings.ToLower(ext)

		filename := uuid.New().String() + ext
		dst, err := os.Create(filepath.Join(uploadDir, filename))
		if err != nil {
			jsonError(w, "failed to save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			jsonError(w, "failed to save file", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, map[string]string{"url": "/static/" + filename}, http.StatusCreated)
	}
}