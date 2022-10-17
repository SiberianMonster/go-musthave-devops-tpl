package httpp

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
)

type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w GzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

//func getBinaryBySHA256(s string) []byte {
//    r := sha256.Sum256([]byte(s))
//    return r[:]
//}

func Hash(value, key string) (string, error) {
	mac := hmac.New(sha256.New, []byte(key))
	_, err := mac.Write([]byte(value))
	return fmt.Sprintf("%x", mac.Sum(nil)), err
}
