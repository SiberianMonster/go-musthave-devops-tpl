// Server module launches the web server for collecting system metrics and their storage in an SQL Database / json-file.
//
// Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/cmd/server
package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "github.com/SiberianMonster/go-musthave-devops-tpl/cmd/server/proto"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/config"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/handlers"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/middleware"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/storage"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var host, storeFile, restore, key, connStr, storeParameter, buildVersion, buildDate, buildCommit, jsonFile, cryptoKey, trustedSub, grpcPort *string
var privateKey *rsa.PrivateKey
var storeInterval string
var db *sql.DB

// GrpcServer supports all proto methods.
type GrpcServer struct {
	pb.UnimplementedGrpcServer

	metrics sync.Map
	key     string
	dB      *sql.DB
	dBFlag  bool
}

// Update receives new data from the client
func (s *GrpcServer) Update(ctx context.Context, in *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	var response pb.UpdateResponse
	var err error
	var updateParams metrics.Metrics

	if in.Metrics.Id == "" {
		log.Println("Empty metrics ID")
		response.Error = "empty metrics ID"
		err = errors.New("empty metrics ID")
		return &response, err
	}

	updateParams.ID = in.Metrics.Id
	updateParams.Hash = in.Metrics.Hash

	if in.Metrics.Mtype == pb.Metrics_GAUGE {
		updateParams.MType = "gauge"
		updateParams.Value = &in.Metrics.Value

	} else {
		updateParams.MType = "counter"
		updateParams.Delta = &in.Metrics.Delta
	}

	if s.key != "" {

		log.Println("checking hash")
		testHash := metrics.MetricsHash(updateParams, s.key)

		if testHash != updateParams.Hash {
			response.Error = "hashing values do not match"
			err = errors.New("hashing values do not match")
			return &response, err

		}
	}

	err = storage.RepositoryUpdate(updateParams, s.dB, s.dBFlag, ctx)
	if err != nil {
		response.Error = "data update failed"
		err = errors.New("data update failed")
	}

	return &response, err
}

// Value returns stored data to the client
func (s *GrpcServer) Value(ctx context.Context, in *pb.ValueRequest) (*pb.ValueResponse, error) {
	var response pb.ValueResponse
	var err error
	var receivedParams metrics.Metrics

	var ok bool
	if s.dBFlag {
		ok = storage.DBCheck(s.dB, in.Metricsname, ctx)

	} else {
		_, ok = metrics.Container[in.Metricsname]
	}

	if !ok {
		log.Printf("missing params value")
		err = errors.New("missing params value")
		return &response, err
	}

	receivedParams.ID = in.Metricsname
	if in.Metricsname == "Counter" {
		receivedParams.MType = "counter"
	}

	retrievedMetrics, getErr := storage.RepositoryRetrieve(receivedParams, s.dB, s.dBFlag, ctx)
	if getErr != nil {
		err = errors.New("value retrieval failed")
		return &response, err
	}

	if s.key != "" {
		retrievedMetrics.Hash = metrics.MetricsHash(retrievedMetrics, s.key)
	}

	result := &pb.Metrics{
		Id: in.Metricsname,
	}
	if retrievedMetrics.MType == "counter" {
		result.Mtype = pb.Metrics_COUNTER
		result.Delta = *retrievedMetrics.Delta
	} else {

		result.Mtype = pb.Metrics_GAUGE
		result.Value = *retrievedMetrics.Value
	}
	response.Metrics = result

	return &response, err
}

// ParseRsaPrivateKey function reads private rsa key from string.
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

// SetUpConfiguration function reads server configuration parameters from .json file.
func SetUpConfiguration(jsonFile *string, serverConfig config.ServerConfig) (config.ServerConfig, error) {

	var err error
	if *jsonFile == "" {
		return serverConfig, nil
	}
	if !strings.Contains(*jsonFile, ".json") {
		err = errors.New("config file should have .json extension")
		log.Printf("Error happened in setting configuration. Err: %s", err)
		return serverConfig, err
	}

	serverConfig, err = config.LoadServerConfiguration(jsonFile, serverConfig)
	if err != nil {
		log.Printf("Error happened when reading configs from json file. Err: %s", err)
		return serverConfig, err
	}
	return serverConfig, nil
}

// SetUpCryptoKey returns private rsa key from a .pem file.
func SetUpCryptoKey(cryptoKey *string, serverConfig config.ServerConfig) (*rsa.PrivateKey, error) {

	var privateKey *rsa.PrivateKey
	var err error
	if *cryptoKey != "" {
		if !strings.Contains(*cryptoKey, ".pem") {
			err = errors.New("private key file should have .pem extension")
			log.Printf("Error happened in opening rsa key. Err: %s", err)
			return nil, err
		}

		privPEM, err := os.ReadFile(*cryptoKey)
		if err != nil {
			log.Printf("Error happened when reading rsa key. Err: %s", err)
			return nil, err
		}

		privateKey, err = ParseRsaPrivateKey(privPEM)
		if err != nil {
			log.Printf("Error happened when decoding rsa key. Err: %s", err)
			return nil, err
		}
	}
	return privateKey, nil
}

// SetUpDataStorage initializes database connection / opens a .json file for storing received values.
func SetUpDataStorage(ctx context.Context, connStr *string, storeFile *string, restoreValue bool, storeInt int, storeParameter *string) (*sql.DB, bool) {

	if len(*connStr) > 0 {
		db, ok := storage.SetUpDBConnection(ctx, connStr)
		return db, ok

	} else {
		if len(*storeFile) > 0 {
			if restoreValue {
				storage.StaticFileUpload(*storeFile)
			}
			go storage.ContainerUpdate(storeInt, *storeFile, db, *storeParameter)
		}
		log.Println("Initialised json storage.")
	}
	return nil, false
}

// ParseStoreInterval function does the procesing of storeinterval input variable.
func ParseStoreInterval(storeParameter *string) int {

	storeInterval = strings.Replace(strings.Replace(*storeParameter, "s", "", -1), "m", "", -1)
	storeInt, err := strconv.Atoi(storeInterval)
	if err != nil {
		log.Printf("Error happened in reading storeInt variable. Err: %s", err)
	}
	return storeInt
}

// ParseRestoreValue function does the procesing of restore input variable.
func ParseRestoreValue(restore *string) (bool, error) {

	var err error
	restoreValue := false
	restoreValue, err = strconv.ParseBool(*restore)
	if err != nil {
		err = errors.New("could not parse restore value")
		log.Printf("Error happened in retrieving env variable. Err: %s", err)
		return restoreValue, err
	}
	return restoreValue, nil
}

// ParseTrustedSub function does the procesing of trustedSubnet input variable.
func ParseTrustedSub(trustedSubnet *string) (*net.IPNet, error) {

	var err error
	var trustedSubNetwork *net.IPNet
	_, trustedSubNetwork, err = net.ParseCIDR(*trustedSubnet)
	if err != nil {
		err = errors.New("could not parse trustedSubnet value")
		log.Printf("Error happened in retrieving env variable. Err: %s", err)
		return nil, err
	}
	return trustedSubNetwork, nil
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

func init() {

	metrics.Container = make(map[string]interface{})
	var err error
	serverConfig := config.NewServerConfig()

	jsonFile = config.GetEnv("CONFIG", flag.String("c", "", "CONFIG"))
	serverConfig, err = SetUpConfiguration(jsonFile, serverConfig)
	if err != nil {
		log.Printf("Error happened in reading json config file. Err: %s", err)
	}

	host = config.GetEnv("ADDRESS", flag.String("a", serverConfig.Address, "ADDRESS"))
	key = config.GetEnv("KEY", flag.String("k", "", "KEY"))
	grpcPort = config.GetEnv("PORT", flag.String("p", ":3200", "PORT"))
	storeParameter = config.GetEnv("STORE_INTERVAL", flag.String("i", serverConfig.StoreInterval, "STORE_INTERVAL"))
	storeFile = config.GetEnv("STORE_FILE", flag.String("f", serverConfig.StoreFile, "STORE_FILE"))
	restore = config.GetEnv("RESTORE", flag.String("r", serverConfig.Restore, "RESTORE"))
	connStr = config.GetEnv("DATABASE_DSN", flag.String("d", serverConfig.DatabaseDsn, "DATABASE_DSN"))
	trustedSub = config.GetEnv("TRUSTED_SUBNET", flag.String("t", serverConfig.TrustedSub, "TRUSTED_SUBNET"))

	cryptoKey = config.GetEnv("CRYPTO_KEY", flag.String("crypto-key", serverConfig.CryptoKey, "CRYPTO_KEY"))
	privateKey, err = SetUpCryptoKey(cryptoKey, serverConfig)
	if err != nil {
		log.Printf("Error happened in uploading encryption key. Err: %s", err)
	}

	buildVersion = config.GetEnv("BUILD_VERSION", flag.String("bv", "N/A", "BUILD_VERSION"))
	buildDate = config.GetEnv("BUILD_DATE", flag.String("bd", "N/A", "BUILD_DATE"))
	buildCommit = config.GetEnv("BUILD_COMMIT", flag.String("bc", "N/A", "BUILD_COMMIT"))
	log.Printf("Build Version: %s", *buildVersion)
	log.Printf("Build Date: %s", *buildDate)
	log.Printf("Build Commit: %s", *buildCommit)

}

// InitializeRouter function returns Gorilla mux router with the endpoints that allow reception / retrieval of system metrics.
func InitializeRouter(privateKey *rsa.PrivateKey, trustedSubNetwork *net.IPNet) *mux.Router {

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
	r.Use(middleware.IPHandler(trustedSubNetwork))
	return r
}

// ParseEnvVariables function wraps processing of passed parameters.
func ParseEnvVariables(restore *string, storeParameter *string, trustedSub *string) (bool, int, *net.IPNet) {
	restoreValue, err := ParseRestoreValue(restore)
	if err != nil {
		log.Printf("Failed to obtain restore variable. Err: %s", err)
	}
	storeInt := ParseStoreInterval(storeParameter)

	trustedSubNetwork, err := ParseTrustedSub(trustedSub)
	if err != nil {
		log.Printf("Failed to obtain trustedSubNetwork variable. Err: %s", err)
	}

	config.Key = *key
	return restoreValue, storeInt, trustedSubNetwork
}

func main() {

	flag.Parse()

	restoreValue, storeInt, trustedSubNetwork := ParseEnvVariables(restore, storeParameter, trustedSub)

	ctx, cancel := context.WithTimeout(context.Background(), config.ContextDBTimeout*time.Second)
	defer cancel()

	config.DB, config.DBFlag = SetUpDataStorage(ctx, connStr, storeFile, restoreValue, storeInt, storeParameter)
	defer config.DB.Close()

	r := InitializeRouter(privateKey, trustedSubNetwork)

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

	listen, err := net.Listen("tcp", *grpcPort)
	if err != nil {
		log.Fatalf("GRPC server error: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGrpcServer(s, &GrpcServer{key: config.Key, dB: config.DB, dBFlag: config.DBFlag})
	if err := s.Serve(listen); err != nil {
		log.Fatalf("GRPC server error: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigChan

	ShutdownGracefully(srv, storeFile, connStr)

}
