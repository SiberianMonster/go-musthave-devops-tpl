package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"
)

type gauge float64
type counter int64

const (
	pollduration   int = 2
	reportduration int = 10
)

type metricsContainer struct {
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
	RandomValue,
	NumForcedGC,
	NumGC gauge
	PollCount counter
}

func MetricsUpdate(m metricsContainer, rtm runtime.MemStats) metricsContainer {

	m.Alloc = gauge(rtm.Alloc)
	m.BuckHashSys = gauge(rtm.BuckHashSys)
	m.Frees = gauge(rtm.Frees)
	m.GCCPUFraction = gauge(rtm.GCCPUFraction)
	m.GCSys = gauge(rtm.GCSys)
	m.HeapAlloc = gauge(rtm.HeapAlloc)
	m.HeapIdle = gauge(rtm.HeapIdle)
	m.HeapInuse = gauge(rtm.HeapInuse)
	m.HeapObjects = gauge(rtm.HeapObjects)
	m.HeapReleased = gauge(rtm.HeapReleased)
	m.HeapSys = gauge(rtm.HeapSys)
	m.LastGC = gauge(rtm.LastGC)
	m.Lookups = gauge(rtm.Lookups)
	m.MCacheInuse = gauge(rtm.MCacheInuse)
	m.MCacheSys = gauge(rtm.MCacheSys)
	m.MSpanInuse = gauge(rtm.MSpanInuse)
	m.MSpanSys = gauge(rtm.MSpanSys)
	m.Mallocs = gauge(rtm.Mallocs)
	m.NextGC = gauge(rtm.NextGC)
	m.NumForcedGC = gauge(rtm.NumForcedGC)
	m.NumGC = gauge(rtm.NumGC)
	m.OtherSys = gauge(rtm.OtherSys)
	m.PauseTotalNs = gauge(rtm.PauseTotalNs)
	m.StackInuse = gauge(rtm.StackInuse)
	m.StackSys = gauge(rtm.StackSys)
	m.Sys = gauge(rtm.Sys)
	m.TotalAlloc = gauge(rtm.TotalAlloc)
	m.PollCount += 1
	m.RandomValue = gauge(rand.Float64())
	return m

}

func ReportUpdate(p int, r int) error {

	var m metricsContainer
	var rtm runtime.MemStats
	var v reflect.Value
	var typeOfS reflect.Type
	var err error

	if p >= r {
		err = errors.New("reportduration needs to be larger than pollduration")
		return err

	}

	pollTicker := time.NewTicker(time.Second * time.Duration(p))
	reportTicker := time.NewTicker(time.Second * time.Duration(r))

	m.PollCount = 0

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

				url := url.URL{
					Scheme: "http",
					Host:   "127.0.0.1:8080",
				}

				if v.Field(i).Kind() == reflect.Float64 {
					url.Path += fmt.Sprintf("update/gauge/%v/%f", typeOfS.Field(i).Name, v.Field(i).Interface())
				} else {
					url.Path += fmt.Sprintf("update/counter/%v/%v", typeOfS.Field(i).Name, v.Field(i).Interface())
				}

				log.Println("Encoded URL is %q\n", url.String())
				client := &http.Client{}
				request, err := http.NewRequest("POST", url.String(), nil)
				if err != nil {
					log.Fatal(err)
					return err

				}
				request.Header.Set("Content-Type", "text/plain")
				response, err := client.Do(request)
				if err != nil {
					log.Fatal(err)
					return err

				}
				response.Body.Close()
				// response status
				log.Println("Status code ", response.Status)

			}

		}

	}
}

func main() {
	err := ReportUpdate(pollduration, reportduration)
	if err != nil {
		log.Fatal(err)
	}

}
