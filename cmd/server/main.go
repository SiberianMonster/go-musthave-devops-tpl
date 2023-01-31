// Server module launches the web server for collecting system metrics and their storage in an SQL Database / json-file.
//
// Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/cmd/server
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
	"crypto/x509"
    "encoding/pem"
	"crypto/rsa"

	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/config"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/handlers"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/middleware"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/storage"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	
)

var host, storeFile, restore, key, connStr, storeParameter, buildVersion, buildDate, buildCommit, cryptoKey, jsonFile *string
var privateKey *rsa.PrivateKey
var storeInterval string
var db *sql.DB

func ParseRsaPrivateKey(privPEM []byte) (*rsa.PrivateKey, error) {
    block, _ := pem.Decode(privPEM)
    if block == nil {
		log.Printf("Error happened when parsing PEM")
        return nil, errors.New("failed to parse PEM block containing the key")
    }

    priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
		log.Printf("Error happened when parsing PEM. Err: %s", err)
        return nil, err
    }

    return priv, nil
}

func init() {

	metrics.Container = make(map[string]interface{})

	serverConfig := config.NewServerConfig()
	jsonFile = config.GetEnv("CONFIG", flag.String("c", "", "CONFIG"))
	if *jsonFile != "" {
		serverConfig = config.LoadServerConfiguration(jsonFile, serverConfig)
	}

	host = config.GetEnv("ADDRESS", flag.String("a",serverConfig.Address, "ADDRESS"))
	key = config.GetEnv("KEY", flag.String("k", "", "KEY"))
	storeParameter = config.GetEnv("STORE_INTERVAL", flag.String("i", serverConfig.StoreInterval, "STORE_INTERVAL"))
	storeFile = config.GetEnv("STORE_FILE", flag.String("f", serverConfig.StoreFile, "STORE_FILE"))
	restore = config.GetEnv("RESTORE", flag.String("r", serverConfig.Restore, "RESTORE"))
	connStr = config.GetEnv("DATABASE_DSN", flag.String("d", serverConfig.DatabaseDsn, "DATABASE_DSN"))
	cryptoKey = config.GetEnv("CRYPTO_KEY", flag.String("crypto-key", serverConfig.CryptoKey, "CRYPTO_KEY"))
	if *cryptoKey != "" {
		privPEM, err := os.ReadFile(*cryptoKey)
		if err != nil {
			log.Printf("Error happened when reading rsa key. Err: %s", err)
		}

		privateKey, err = ParseRsaPrivateKey(privPEM)
		if err != nil {
			log.Printf("Error happened when decoding rsa key. Err: %s", err)
		}
	} 

	buildVersion = config.GetEnv("BUILD_VERSION", flag.String("bv", "N/A", "BUILD_VERSION"))
	buildDate = config.GetEnv("BUILD_DATE", flag.String("bd", "N/A", "BUILD_DATE"))
	buildCommit = config.GetEnv("BUILD_COMMIT", flag.String("bc", "N/A", "BUILD_COMMIT"))
	log.Printf("Build Version: %s", *buildVersion)
	log.Printf("Build Date: %s", *buildDate)
	log.Printf("Build Commit: %s", *buildCommit)

}

// InitializeRouter function returns Gorilla mux router with the endpoints that allow reception / retrieval of system metrics.
func InitializeRouter(privateKey *rsa.PrivateKey) *mux.Router {

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
	r.Use(middleware.EncryptionHandler(privateKey))
	return r
}

// ParseStoreInterval function does the procesing of storeinterval input variable. 
func ParseStoreInterval(storeParameter *string) int {

	storeInterval = strings.Replace(strings.Replace(*storeParameter, "s", "", -1), "m", "", -1)
	storeInt, err := strconv.Atoi(storeInterval)
	if err != nil {
		log.Fatalf("Error happened in reading storeInt variable. Err: %s", err)
	}
	return storeInt
}

// ParseRestoreValue function does the procesing of restore input variable. 
func ParseRestoreValue(restore *string) bool {

	restoreValue, err := strconv.ParseBool(*restore)
	if err != nil {
		log.Fatalf("Error happened in reading restoreValue variable. Err: %s", err)
	}
	return restoreValue
}

// ShutdownGracefully handles server shutdown and information saving.
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
	config.PrivateKey = privateKey

	if len(*connStr) > 0 {
		log.Println("Start db connection.")
		ctx, cancel := context.WithTimeout(context.Background(), config.ContextDBTimeout*time.Second)
		// не забываем освободить ресурс
		defer cancel()
		var err error
		db, err = sql.Open("postgres", *connStr)
		if err != nil {
			log.Fatalf("Error happened when initiating connection to the db. Err: %s", err)
		}
		_, err = db.ExecContext(ctx,
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

	r := InitializeRouter(config.PrivateKey)

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
