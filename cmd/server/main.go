package main

import (
	"context"
	"errors"
	"flag"
	"github.com/gorilla/mux"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/generalutils"
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
	"database/sql"
	_ "github.com/lib/pq"
)

var err error
var host, storeFile, restore, key, connStr *string
var storeInterval string
var db *sql.DB

func init() {

	generalutils.Container = make(map[string]interface{})

	host = generalutils.GetEnv("ADDRESS", flag.String("a", "127.0.0.1:8080", "ADDRESS"))
	key = generalutils.GetEnv("KEY", flag.String("k","", "KEY"))
	storeInterval = strings.Replace(*generalutils.GetEnv("STORE_INTERVAL", flag.String("i", "300", "STORE_INTERVAL")), "s", "", -1)
	storeFile = generalutils.GetEnv("STORE_FILE", flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE"))
	restore = generalutils.GetEnv("RESTORE", flag.String("r", "true", "RESTORE"))
	connStr = generalutils.GetEnv("DATABASE_DSN", flag.String("d", "", "DATABASE_DSN"))
	
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

	storage.StaticFileUpload(*storeFile, restoreValue)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	r := mux.NewRouter()
	handlersWithKey := handlers.WrapperJSONStruct{Hashkey: *key}

	if len(*connStr) > 0 {
		db, err = sql.Open("postgres", *connStr)
		if err != nil {
			log.Fatalf("Error happened when initiating connection to the db. Err: %s", err)
		}
		handlersWithKey.DB = db
	}

	r.HandleFunc("/update/", handlersWithKey.UpdateJSONHandler)
	r.HandleFunc("/value/", handlersWithKey.ValueJSONHandler)
	r.HandleFunc("/update/{type}/{name}/{value}", handlersWithKey.UpdateStringHandler)
	r.HandleFunc("/value/{type}/{name}", handlersWithKey.ValueStringHandler)
	r.HandleFunc("/ping", handlersWithKey.PostgresHandler)

	r.HandleFunc("/", handlersWithKey.GenericHandler)
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

	go storage.ContainerUpdate(storeInt, *storeFile, db, ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 40*time.Second)
	defer shutdownRelease()
	defer db.Close()
	// не забываем освободить ресурс
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")
	storage.StaticFileSave(*storeFile)

}
