// Agent module enables repetitive collection of system metrics and posts the collected stats to a server.
//
// Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/cmd/agent
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "github.com/SiberianMonster/go-musthave-devops-tpl/cmd/server/proto"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/config"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/shirou/gopsutil/v3/mem"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var host, key, buildVersion, buildDate, buildCommit, cryptoKey, jsonFile, grpcPort *string
var publicKey *rsa.PublicKey
var pollCounterEnv, reportCounterEnv string
var rtm runtime.MemStats
var v reflect.Value
var typeOfS reflect.Type
var err error

// LockMetricsContainer is a struct that serves to store and transmit system metrics.
// It contains RMutex attribute to avoid data corruption.
type LockMetricsContainer struct {
	m  metrics.MetricsContainer
	mu sync.RWMutex
}

var Lm LockMetricsContainer

// CounterCheck function controls variables used as intervals for system metrics collection and posting.
// The function checks that the posting interval is always larger than the collection interval.
func CounterCheck(pollCounterVar int, reportCounterVar int) error {

	if pollCounterVar >= reportCounterVar {
		err = errors.New("reportduration needs to be larger than pollduration")
		log.Printf("Error happened in setting timer. Err: %s", err)
		return err
	}
	return nil
}

// CollectStats function is used to collect system metrics from runtime.ReadMemStats.
func CollectStats() {

	log.Println("Collecting stats")
	Lm.mu.Lock()
	defer Lm.mu.Unlock()
	runtime.ReadMemStats(&rtm)
	Lm.m = metrics.MetricsUpdate(Lm.m, rtm)
}

// CollectMemStats function collects additional system metrics with mem.VirtualMemory.
func CollectMemStats() {

	log.Println("Collecting mem stats")
	Lm.mu.Lock()
	defer Lm.mu.Unlock()
	v, _ := mem.VirtualMemory()
	Lm.m.TotalMemory = float64(v.Total)
	Lm.m.FreeMemory = float64(v.Free)
	Lm.m.CPUutilization1 = v.UsedPercent
}

// SendMemStats function posts collected statistical data to the server.
func SendMemStats(metricsObj metrics.Metrics, urlString string, publicKey *rsa.PublicKey) {

	body, err := json.Marshal(metricsObj)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
	}
	log.Print(string(body))

	if publicKey != nil {
		PostEncryptedStats(body, urlString, publicKey)
		return
	}

	realIP := GetOutboundIP()

	client := &http.Client{}
	client.Timeout = config.RequestTimeout * time.Second
	req, reqerr := http.NewRequest(http.MethodPost, urlString, bytes.NewBuffer(body))
	if reqerr != nil {
		log.Printf("Error happened when creating a post request. Err: %s", reqerr)
		return

	}
	req.Header.Set("Content-Encoding", "application/json")
	req.Header.Set("X-Real-IP", realIP.String())
	response, err := client.Do(req)

	if err != nil {
		log.Printf("Error happened when response received. Err: %s", err)
		return

	}
	err = response.Body.Close()
	if err != nil {
		log.Printf("Error happened when response body closed. Err: %s", err)
		return
	}
	// response status
	log.Printf("Status code %q\n", response.Status)
}

// PostEncryptedStats function encrypts and posts assymetricaly encrypted data to the server.
func PostEncryptedStats(body []byte, urlString string, publicKey *rsa.PublicKey) {

	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		body,
		nil)
	if err != nil {
		log.Printf("Error happened when encryptying message body. Err: %s", err)
	}
	response, err := http.Post(urlString, "application/json", bytes.NewBuffer(encryptedBytes))
	if err != nil {
		log.Printf("Error happened when response received. Err: %s", err)
		return

	}
	err = response.Body.Close()
	if err != nil {
		log.Printf("Error happened when response body closed. Err: %s", err)
		return
	}
	// response status
	log.Printf("Status code %q\n", response.Status)
}

// ParseRsaPublicKey function reads public rsa key from string.
func ParseRsaPublicKey(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		log.Printf("Error happened when parsing PEM. Err: %s", err)
		return nil, err
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Printf("Error happened when parsing PEM. Err: %s", err)
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("key type is not RSA")
	}
}

// ReportStats writes collected each system metric as a request body and posts them to the server.
func ReportStats() {

	Lm.mu.RLock()
	defer Lm.mu.RUnlock()
	v = reflect.ValueOf(Lm.m)
	typeOfS = v.Type()

	log.Println("Reporting stats")

	for i := 0; i < v.NumField(); i++ {

		url := url.URL{
			Scheme: "http",
			Host:   *host,
		}
		url.Path += "update/"

		var metricsObj metrics.Metrics

		if v.Field(i).Kind() == reflect.Float64 {
			metricsObj.ID = typeOfS.Field(i).Name
			metricsObj.MType = metrics.Gauge
			value := v.Field(i).Interface().(float64)
			metricsObj.Value = &value
			if *key != "" {
				metricsObj.Hash = metrics.MetricsHash(metricsObj, *key)
			}

		} else {
			metricsObj.ID = typeOfS.Field(i).Name
			metricsObj.MType = metrics.Counter
			delta := v.Field(i).Interface().(int64)
			metricsObj.Delta = &delta
			if *key != "" {
				metricsObj.Hash = metrics.MetricsHash(metricsObj, *key)
			}
		}

		SendMemStats(metricsObj, url.String(), publicKey)

	}

}

// ReportUpdateBatch allows to send all collected metrics in a single http request.
// All the metrics are appended to a single slice of metrics objects.
func ReportUpdateBatch(pollCounterVar int, reportCounterVar int) error {

	var m metrics.MetricsContainer
	var rtm runtime.MemStats
	var v reflect.Value
	var typeOfS reflect.Type
	var err error

	if pollCounterVar >= reportCounterVar {
		err = errors.New("reportduration needs to be larger than pollduration")
		return err

	}

	pollTicker := time.NewTicker(time.Second * time.Duration(pollCounterVar))
	reportTicker := time.NewTicker(time.Second * time.Duration(reportCounterVar))

	m.PollCount = 0
	client := &http.Client{Timeout: config.RequestTimeout * time.Second}

	for {

		select {
		case <-pollTicker.C:
			// update stats
			runtime.ReadMemStats(&rtm)
			m = metrics.MetricsUpdate(m, rtm)
			v = reflect.ValueOf(m)
			typeOfS = v.Type()
		case <-reportTicker.C:
			// send stats to the server

			url := url.URL{
				Scheme: "http",
				Host:   *host,
			}
			url.Path += "updates/"

			metricsBatch := []metrics.Metrics{}
			for i := 0; i < v.NumField(); i++ {

				var metricsObj metrics.Metrics

				if v.Field(i).Kind() == reflect.Float64 {
					metricsObj.ID = typeOfS.Field(i).Name
					metricsObj.MType = metrics.Gauge
					value := v.Field(i).Interface().(float64)
					metricsObj.Value = &value
					if *key != "" {
						metricsObj.Hash = metrics.MetricsHash(metricsObj, *key)
					}

				} else {
					metricsObj.ID = typeOfS.Field(i).Name
					metricsObj.MType = metrics.Counter
					delta := v.Field(i).Interface().(int64)
					metricsObj.Delta = &delta
					if *key != "" {
						metricsObj.Hash = metrics.MetricsHash(metricsObj, *key)
					}
				}
				metricsBatch = append(metricsBatch, metricsObj)
			}
			if len(metricsBatch) > 0 {
				body, err := json.Marshal(metricsBatch)
				if err != nil {
					log.Printf("Error happened in JSON marshal. Err: %s", err)
					return err
				}
				log.Print(string(body))

				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				gz.Write(body)
				gz.Close()

				request, err := http.NewRequest(http.MethodPost, url.String(), &buf)
				if err != nil {
					log.Fatalf("Error happened when request made. Err: %s", err)
				}

				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Accept", "application/json")
				request.Header.Set("Content-Encoding", "gzip")
				response, err := client.Do(request)
				if err != nil {
					log.Printf("Error happened when response received. Err: %s", err)
					continue

				}
				err = response.Body.Close()
				if err != nil {
					log.Printf("Error happened when response body closed. Err: %s", err)
					continue
				}
				// response status
				log.Printf("Status code %q\n", response.Status)

			}

		}

	}
}

// GetOutboundIP allows to retrieve preferred outbound ip of the machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("Error happened when dialing udp for IP address. Err: %s", err)
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func init() {

	agentConfig := config.NewAgentConfig()
	jsonFile = config.GetEnv("CONFIG", flag.String("c", "", "CONFIG"))
	if *jsonFile != "" {
		agentConfig = config.LoadAgentConfiguration(jsonFile, agentConfig)
	}
	host = config.GetEnv("ADDRESS", flag.String("a", agentConfig.Address, "ADDRESS"))
	grpcPort = config.GetEnv("PORT", flag.String("gp", ":3200", "PORT"))
	pollCounterEnv = strings.Replace(*config.GetEnv("POLL_INTERVAL", flag.String("p", agentConfig.PollInterval, "POLL_INTERVAL")), "s", "", -1)
	reportCounterEnv = strings.Replace(*config.GetEnv("REPORT_INTERVAL", flag.String("r", agentConfig.ReportInterval, "REPORT_INTERVAL")), "s", "", -1)
	key = config.GetEnv("KEY", flag.String("k", "", "KEY"))
	buildVersion = config.GetEnv("BUILD_VERSION", flag.String("bv", "N/A", "BUILD_VERSION"))
	buildDate = config.GetEnv("BUILD_DATE", flag.String("bd", "N/A", "BUILD_DATE"))
	buildCommit = config.GetEnv("BUILD_COMMIT", flag.String("bc", "N/A", "BUILD_COMMIT"))
	cryptoKey = config.GetEnv("CRYPTO_KEY", flag.String("crypto-key", agentConfig.CryptoKey, "CRYPTO_KEY"))
	if *cryptoKey != "" {
		pubPEM, err := os.ReadFile(*cryptoKey)
		if err != nil {
			log.Printf("Error happened when reading file with rsa key. Err: %s", err)
		}
		publicKey, err = ParseRsaPublicKey(pubPEM)
		if err != nil {
			log.Printf("Error happened when decoding rsa key. Err: %s", err)
		}
	}

	log.Printf("Build Version: %s", *buildVersion)
	log.Printf("Build Date: %s", *buildDate)
	log.Printf("Build Commit: %s", *buildCommit)
}

func main() {

	flag.Parse()

	pollCounterVar, err := strconv.Atoi(pollCounterEnv)
	if err != nil {
		log.Fatalf("Error happened in reading poll counter variable. Err: %s", err)
	}

	reportCounterVar, err := strconv.Atoi(reportCounterEnv)
	if err != nil {
		log.Fatalf("Error happened in reading report counter variable. Err: %s", err)
	}

	err = CounterCheck(pollCounterVar, reportCounterVar)
	if err != nil {
		log.Fatalf("Error happened in checking counter variables. Err: %s", err)
	}

	pollTicker := time.NewTicker(time.Second * time.Duration(pollCounterVar))
	reportTicker := time.NewTicker(time.Second * time.Duration(reportCounterVar))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	conn, err := grpc.Dial(*grpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	grpcClient := pb.NewGrpcClient(conn)

	testInputGauge := &pb.UpdateRequest{Metrics: &pb.Metrics{
		Id:    "TotalMemAlloc",
		Mtype: pb.Metrics_GAUGE,
		Delta: 0,
		Value: 1.11,
		Hash:  "",
	},
	}
	testInputCounter := &pb.UpdateRequest{Metrics: &pb.Metrics{
		Id:    "Counter",
		Mtype: pb.Metrics_COUNTER,
		Delta: 1,
		Value: 0.0,
		Hash:  "",
	},
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.ContextDBTimeout*time.Second)
	defer cancel()
	_, err = grpcClient.Update(ctx, testInputGauge)
	_, err = grpcClient.Update(ctx, testInputCounter)

loop:
	for {

		select {
		case <-pollTicker.C:
			// update stats
			go CollectStats()
			go CollectMemStats()

		case <-reportTicker.C:
			// send stats to the server
			go ReportStats()

		case <-sigChan:
			break loop
		}
	}
	// send latest stats to the server
	ReportStats()
	log.Println("Graceful shutdown")

}
