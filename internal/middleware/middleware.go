package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/httpp"
)

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

