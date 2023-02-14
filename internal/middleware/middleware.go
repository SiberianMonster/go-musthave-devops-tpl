// Middlerware package contains gzip wrapper for the server endpoints handlers.
//
// Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/internal/middlerware
package middleware

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"github.com/gorilla/mux"
	"github.com/klauspost/compress/gzip"
	"io"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/httpp"
)

// GzipHandler function retruns a gzip wrapper for the server endpoints handlers.
func GzipHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gz
			defer gz.Close()
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		h.ServeHTTP(httpp.GzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

// EncryptionHandler ensures incoming messages decryption.
func EncryptionHandler(privateKey *rsa.PrivateKey) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if privateKey != nil {
				defer r.Body.Close()
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				decryptedBytes, err := privateKey.Decrypt(nil, bodyBytes, &rsa.OAEPOptions{Hash: crypto.SHA256})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(decryptedBytes))
			}
			h.ServeHTTP(w, r)
		})
	}
}

// IPHandler ensures processing requests from trusted subnets only.
func IPHandler(trustedSubnet *net.IPNet) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if trustedSubnet != nil {
				reqIP := r.Header.Get("X-Real-IP")
				if !trustedSubnet.Contains(reqIP) {
					resp = make(map[string]string)
					w.WriteHeader(http.StatusForbidden)
					resp["status"] = "network not trusted"
					jsonResp, err := json.Marshal(resp)
					if err != nil {
						log.Printf("Error happened in JSON marshal. Err: %s", err)
						return
					}
					w.Write(jsonResp)
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}