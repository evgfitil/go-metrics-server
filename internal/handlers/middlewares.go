package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
)

const hashHeader = "HashSHA256"

type hashResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
	key string
}

func (w *hashResponseWriter) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}

func VerifyRequestHash(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if req.Method == "POST" && key != "" && req.Header.Get(hashHeader) != "" {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					http.Error(res, "error reading body", http.StatusBadRequest)
					return
				}
				req.Body = io.NopCloser(bytes.NewBuffer(body))
				receivedHash := req.Header.Get(hashHeader)
				if !verifyHash(key, body, receivedHash) {
					http.Error(res, "invalid hash", http.StatusBadRequest)
					return
				}
			}
			next.ServeHTTP(res, req)
		})
	}
}

func WithResponseHash(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if key != "" {
				buf := &bytes.Buffer{}
				hashWriter := &hashResponseWriter{ResponseWriter: res, buf: buf, key: key}

				next.ServeHTTP(hashWriter, req)

				hash := computeHash(key, buf.Bytes())
				hashWriter.Header().Set(hashHeader, hash)

				res.Write(buf.Bytes())
			} else {
				next.ServeHTTP(res, req)
			}
		})
	}
}

func computeHash(key string, data []byte) string {
	secretKey := []byte(key)
	h := hmac.New(sha256.New, secretKey)
	h.Write(data)
	dst := h.Sum(nil)
	return fmt.Sprintf("%x", dst)
}

func verifyHash(key string, data []byte, receivedHash string) bool {
	expectedHash := computeHash(key, data)
	return hmac.Equal([]byte(expectedHash), []byte(receivedHash))
}
