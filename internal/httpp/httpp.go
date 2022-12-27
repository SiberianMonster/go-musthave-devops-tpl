// Httpp package contains support struct and functions for the Server module
//
//Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/internal/httpp
package httpp

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
)

// GzipWriter struct is a wrapper for the http.ResponseWriter that enables gzip-encoding
type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write function for the GzipWriter struct 
func (w GzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

// Hash function is used for hashing values with the sha256 algorythm
func Hash(value, key string) (string, error) {
	mac := hmac.New(sha256.New, []byte(key))
	_, err := mac.Write([]byte(value))
	return fmt.Sprintf("%x", mac.Sum(nil)), err
}
