package middleware

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
)

type bufferedResponseWriter struct {
	http.ResponseWriter
	buf        bytes.Buffer
	statusCode int
}

func (w *bufferedResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// ETagMiddleware calculates the MD5 hash of the response body and sets the ETag header.
// If the client sends a matching If-None-Match header, it returns 304 Not Modified.
// Note: This middleware buffers the entire response in memory, so it's best for
// small to medium-sized responses (like HTML pages), not large file downloads.
func ETagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip for non-GET/HEAD methods
		if r.Method != "GET" && r.Method != "HEAD" {
			next.ServeHTTP(w, r)
			return
		}

		bw := &bufferedResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(bw, r)

		// Calculate ETag
		hash := md5.Sum(bw.buf.Bytes())
		etag := fmt.Sprintf(`"%x"`, hash)

		// Check If-None-Match
		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		// Set ETag and write response
		w.Header().Set("ETag", etag)
		w.WriteHeader(bw.statusCode)
		w.Write(bw.buf.Bytes())
	})
}
