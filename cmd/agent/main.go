package main

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"
	"encoding/json"
	"os"
	"bytes"
	"net/http/httptrace"
	"io/ioutil"
	"io"
)

type gauge float64
type counter int64


func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}


type metricsContainer struct {
	PollCount int64
	RandomValue,
	Alloc,
	BuckHashSys,
	Frees,
	GCSys,
	HeapAlloc,
	HeapIdle,
	HeapInuse,
	HeapObjects,
	HeapReleased,
	HeapSys,
	LastGC,
	Lookups,
	MCacheInuse,
	MCacheSys,
	MSpanInuse,
	MSpanSys,
	Mallocs,
	NextGC,
	OtherSys,
	PauseTotalNs,
	StackInuse,
	StackSys,
	Sys,
	TotalAlloc,
	GCCPUFraction,
	NumForcedGC,
	NumGC float64
}

func MetricsUpdate(m metricsContainer, rtm runtime.MemStats) metricsContainer {

	m.Alloc = float64(rtm.Alloc)
	m.BuckHashSys = float64(rtm.BuckHashSys)
	m.Frees = float64(rtm.Frees)
	m.GCCPUFraction = float64(rtm.GCCPUFraction)
	m.GCSys = float64(rtm.HeapAlloc)
	m.HeapIdle = float64(rtm.HeapIdle)
	m.HeapInuse = float64(rtm.HeapInuse)
	m.HeapObjects = float64(rtm.HeapObjects)
	m.HeapReleased = float64(rtm.HeapReleased)
	m.HeapSys = float64(rtm.HeapSys)
	m.LastGC = float64(rtm.LastGC)
	m.Lookups = float64(rtm.Lookups)
	m.MCacheInuse = float64(rtm.MCacheInuse)
	m.MCacheSys = float64(rtm.MCacheSys)
	m.MSpanInuse = float64(rtm.MSpanInuse)
	m.MSpanSys = float64(rtm.MSpanSys)
	m.Mallocs = float64(rtm.Mallocs)
	m.NextGC = float64(rtm.NextGC)
	m.NumForcedGC = float64(rtm.NumForcedGC)
	m.NumGC = float64(rtm.NumGC)
	m.OtherSys = float64(rtm.OtherSys)
	m.PauseTotalNs = float64(rtm.PauseTotalNs)
	m.StackInuse = float64(rtm.StackInuse)
	m.StackSys = float64(rtm.StackSys)
	m.Sys = float64(rtm.Sys)
	m.TotalAlloc = float64(rtm.TotalAlloc)
	m.PollCount += 1
	m.RandomValue = rand.Float64()
	return m

}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
} 

func ReportUpdate(p int, r int) error {

	var m metricsContainer
	var rtm runtime.MemStats
	var v reflect.Value
	var typeOfS reflect.Type
	var err error
	var metrics Metrics

	h := getEnv("ADDRESS", "localhost:8080")

	if p >= r {
		err = errors.New("reportduration needs to be larger than pollduration")
		return err

	}

	pollTicker := time.NewTicker(time.Second * time.Duration(p))
	reportTicker := time.NewTicker(time.Second * time.Duration(r))

	m.PollCount = 0

	clientTrace := &httptrace.ClientTrace{
        GotConn: func(info httptrace.GotConnInfo) { log.Printf("conn was reused: %t", info.Reused) },
    }
    traceCtx := httptrace.WithClientTrace(context.Background(), clientTrace)

	for {

		select {
		case <-pollTicker.C:
			// update stats
			runtime.ReadMemStats(&rtm)
			m = MetricsUpdate(m, rtm)
			v = reflect.ValueOf(m)
			typeOfS = v.Type()
		case <-reportTicker.C:
			// send stats to the server

			for i := 0; i < v.NumField(); i++ {

				time.Sleep(500 * time.Millisecond)

				url := url.URL{
					Scheme: "http",
					Host:   h,
				}
				url.Path += "update/"

				if v.Field(i).Kind() == reflect.Float64 {
					metrics.ID =  typeOfS.Field(i).Name   
					metrics.MType =  "gauge"      
					value := v.Field(i).Interface().(float64)
					metrics.Value = &value

				} else {
					metrics.ID =  typeOfS.Field(i).Name   
					metrics.MType =  "counter"      
					delta := v.Field(i).Interface().(int64)
					metrics.Delta = &delta
				}

				body, _ := json.Marshal(&metrics)
				log.Print(string(body))

				request, err := http.NewRequestWithContext(traceCtx, http.MethodPost, url.String(), bytes.NewBuffer(body))
				if err != nil {
					log.Printf("Error when request made")
					log.Fatal(err)
					return err
				}
				
				request.Header.Set("Content-Type", "application/json")	
				response, err := http.DefaultClient.Do(request)
				if err != nil {
					log.Printf("Error when response received")
					log.Printf("Error type %q\n", err)
					//log.Fatal(err)
					return err

				} 
				if _, err := io.Copy(ioutil.Discard, response.Body); err != nil {
					log.Fatal(err)
				}
				log.Printf("Status code %q\n", response.Status)
				response.Body.Close()
				
				// response status
				//log.Printf("Status code %q\n", response.Status)
			}

		}

	}
}

func main() {

	
	err := ReportUpdate(2, 10)
	if err != nil {
		log.Fatal(err)
	}

}
