// Agent module enables repetitive collection of system metrics and posts the collected stats to a server.
//
// Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/cmd/agent
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof" 
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/config"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/shirou/gopsutil/v3/mem"
)

var host, key, buildVersion, buildDate, buildCommit *string
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

// CounterCheck function is used to collect system metrics from runtime.ReadMemStats.
func CollectStats() {

	log.Println("Collecting stats")
	Lm.mu.Lock()
	defer Lm.mu.Unlock()
	runtime.ReadMemStats(&rtm)
	Lm.m = metrics.MetricsUpdate(Lm.m, rtm)
}

// CounterCheck function collects additional system metrics with mem.VirtualMemory.
func CollectMemStats() {

	log.Println("Collecting mem stats")
	Lm.mu.Lock()
	defer Lm.mu.Unlock()
	v, _ := mem.VirtualMemory()
	Lm.m.TotalMemory = float64(v.Total)
	Lm.m.FreeMemory = float64(v.Free)
	Lm.m.CPUutilization1 = v.UsedPercent
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

		body, err := json.Marshal(metricsObj)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
		}
		log.Print(string(body))

		response, err := http.Post(url.String(), "application/json", bytes.NewBuffer(body))
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
	client := &http.Client{}

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

func init() {

	host = config.GetEnv("ADDRESS", flag.String("a", "127.0.0.1:8080", "ADDRESS"))
	pollCounterEnv = strings.Replace(*config.GetEnv("POLL_INTERVAL", flag.String("p", "2", "POLL_INTERVAL")), "s", "", -1)
	reportCounterEnv = strings.Replace(*config.GetEnv("REPORT_INTERVAL", flag.String("r", "10", "REPORT_INTERVAL")), "s", "", -1)
	key = config.GetEnv("KEY", flag.String("k", "", "KEY"))
	buildVersion = config.GetEnv("BUILD_VERSION", flag.String("bv", "N/A", "BUILD_VERSION"))
	buildDate = config.GetEnv("BUILD_DATE", flag.String("bd", "N/A", "BUILD_DATE"))
	buildCommit = config.GetEnv("BUILD_COMMIT", flag.String("bc", "N/A", "BUILD_COMMIT"))
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

	for {

		select {
		case <-pollTicker.C:
			// update stats
			go CollectStats()
			go CollectMemStats()

		case <-reportTicker.C:
			// send stats to the server
			go ReportStats()
		}
	}

}
