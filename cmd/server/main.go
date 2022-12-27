package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/config"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/handlers"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/middleware"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/storage"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var err error
var host, storeFile, restore, key, connStr, storeParameter *string
var storeInterval string
var db *sql.DB

func init() {

	metrics.Container = make(map[string]interface{})

	host = config.GetEnv("ADDRESS", flag.String("a", "127.0.0.1:8080", "ADDRESS"))
	key = config.GetEnv("KEY", flag.String("k", "", "KEY"))
	storeParameter = config.GetEnv("STORE_INTERVAL", flag.String("i", "300", "STORE_INTERVAL"))
	storeFile = config.GetEnv("STORE_FILE", flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE"))
	restore = config.GetEnv("RESTORE", flag.String("r", "true", "RESTORE"))
	connStr = config.GetEnv("DATABASE_DSN", flag.String("d", "", "DATABASE_DSN"))

}

func InitializeRouter() *mux.Router {

	r := mux.NewRouter()

	handlersWithKey := handlers.NewWrapperJSONStruct()
	r.HandleFunc("/update/", handlersWithKey.UpdateJSONHandler)
	r.HandleFunc("/value/", handlersWithKey.ValueJSONHandler)
	r.HandleFunc("/update/{type}/{name}/{value}", handlersWithKey.UpdateStringHandler)
	r.HandleFunc("/value/{type}/{name}", handlersWithKey.ValueStringHandler)
	r.HandleFunc("/ping", handlersWithKey.PostgresHandler)
	r.HandleFunc("/updates/", handlersWithKey.UpdateBatchJSONHandler)

	r.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	r.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	r.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	r.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	r.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	r.Handle("/debug/pprof/{cmd}", http.HandlerFunc(pprof.Index)) // special handling for Gorilla mux

	r.HandleFunc("/", handlersWithKey.GenericHandler)
	r.Use(middleware.GzipHandler)
	return r
}

func ParseStoreInterval(storeParameter *string) int {

	storeInterval = strings.Replace(strings.Replace(*storeParameter, "s", "", -1), "m", "", -1)
	storeInt, err := strconv.Atoi(storeInterval)
	if err != nil {
		log.Fatalf("Error happened in reading storeInt variable. Err: %s", err)
	}
	return storeInt
}

func ParseRestoreValue(restore *string) bool {

	restoreValue, err := strconv.ParseBool(*restore)
	if err != nil {
		log.Fatalf("Error happened in reading restoreValue variable. Err: %s", err)
	}
	return restoreValue
}

func ShutdownGracefully(srv *http.Server, storeFile *string, connStr *string) {

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), config.ContextSrvTimeout*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")

	if len(*storeFile) > 0 && len(*connStr) == 0 {
		storage.StaticFileSave(*storeFile)
	}
}

func main() {

	flag.Parse()

	restoreValue := ParseRestoreValue(restore)

	storeInt := ParseStoreInterval(storeParameter)

	config.Key = *key

	if len(*connStr) > 0 {
		log.Println("Start db connection.")
		ctx, cancel := context.WithTimeout(context.Background(), config.ContextDBTimeout*time.Second)
		// не забываем освободить ресурс
		defer cancel()
		db, err = sql.Open("postgres", *connStr)
		if err != nil {
			log.Fatalf("Error happened when initiating connection to the db. Err: %s", err)
		}
		_, err := db.ExecContext(ctx,
			"CREATE TABLE IF NOT EXISTS metrics (metrics_id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY, name text NOT NULL, delta bigint, value double precision)")
		if err != nil {
			log.Fatalf("Error happened when creating sql table. Err: %s", err)

		}

		config.DB = db
		config.DBFlag = true
		defer db.Close()

	} else {
		if len(*storeFile) > 0 {
			if restoreValue {
				storage.StaticFileUpload(*storeFile)
			}
			go storage.ContainerUpdate(storeInt, *storeFile, db, *storeParameter)
		}
		config.DBFlag = false
	}

	r := InitializeRouter()

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

	ShutdownGracefully(srv, storeFile, connStr)

}
