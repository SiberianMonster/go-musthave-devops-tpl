package main

import (
	"context"
	"errors"
	"flag"
	"github.com/gorilla/mux"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/config"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/handlers"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/middleware"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var err error
var host, storeFile, restore *string
var storeInterval string

func init() {

	metrics.Container = make(map[string]interface{})

	host = config.GetEnv("ADDRESS", flag.String("a", "127.0.0.1:8080", "ADDRESS"))
	storeInterval = strings.Replace(*config.GetEnv("STORE_INTERVAL", flag.String("i", "300", "STORE_INTERVAL")), "s", "", -1)
	storeFile = config.GetEnv("STORE_FILE", flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE"))
	restore = config.GetEnv("RESTORE", flag.String("r", "true", "RESTORE"))

}

func main() {

	flag.Parse()

	restoreValue, err := strconv.ParseBool(*restore)
	if err != nil {
		log.Fatalf("Error happened in reading restoreValue variable. Err: %s", err)
	}

	storeInt, err := strconv.Atoi(storeInterval)
	if err != nil {
		log.Fatalf("Error happened in reading storeInt variable. Err: %s", err)
	}

	if restore {
		storage.StaticFileUpload(*storeFile)
	}

	go storage.StaticFileUpdate(storeInt, *storeFile)

	r := mux.NewRouter()

	r.HandleFunc("/update/", handlers.UpdateJSONHandler)
	r.HandleFunc("/value/", handlers.ValueJSONHandler)
	r.HandleFunc("/update/{type}/{name}/{value}", handlers.UpdateStringHandler)
	r.HandleFunc("/value/{type}/{name}", handlers.ValueStringHandler)

	r.HandleFunc("/", handlers.GenericHandler)
	r.Use(middleware.GzipHandler)

	srv := &http.Server{
		Handler: r,
		Addr:    *host,
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), config.ContextSrvTimeout*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")
	storage.StaticFileSave(*storeFile)

}
