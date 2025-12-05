package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog" // Added this import
	"os"
	"strings"
)

var SecretKey = []byte(os.Getenv("SESSION_SECRET"))

func init() {
	if len(SecretKey) == 0 {
		slog.Error("SESSION_SECRET environment variable not set. This is required for secure session management.")
		os.Exit(1)
	}
}

// Sign creates a signed token from the given data string.
// Format: base64(data).base64(hmac)
func Sign(data string) string {
	h := hmac.New(sha256.New, SecretKey)
	h.Write([]byte(data))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return base64.URLEncoding.EncodeToString([]byte(data)) + "." + signature
}

// Verify checks the signature of the token and returns the original data.
func Verify(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", errors.New("invalid token format")
	}

	dataBytes, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", err
	}
	data := string(dataBytes)

	h := hmac.New(sha256.New, SecretKey)
	h.Write([]byte(data))
	expectedSignature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
		return "", errors.New("invalid signature")
	}

	return data, nil
}
