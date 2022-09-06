package main

import (

	"time"
	"runtime"
	"errors"
	"math/rand"
	"reflect"
	"fmt"
	"net/url"
	"net/http"
)

type gauge float64
type counter int64

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

func MetricsUpdate(m metricsContainer, rtm runtime.MemStats ) metricsContainer {

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
	m.MSpanSys =gauge(rtm.MSpanSys)
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



func ReportUpdate(pollduration int, reportduration int) error {

	var m metricsContainer
	var rtm runtime.MemStats
	var v reflect.Value
	var typeOfS reflect.Type
	var err error

	if pollduration >= reportduration {
		err = errors.New("reportduration needs to be larger than pollduration")
		return err
	} else {
		pollTicker := time.NewTicker(time.Second*time.Duration(pollduration))
		reportTicker := time.NewTicker(time.Second*time.Duration(reportduration))

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

				for i := 0; i< v.NumField(); i++ {

						url := url.URL{
							Scheme: "http",
							Host:   "127.0.0.1:8080",
						}

						if v.Field(i).Kind() == reflect.Float64 {
							url.Path += fmt.Sprintf("update/gauge/%v/%f", typeOfS.Field(i).Name, v.Field(i).Interface()) 
						} else {
							url.Path += fmt.Sprintf("update/counter/%v/%v", typeOfS.Field(i).Name, v.Field(i).Interface()) 
						}

						fmt.Printf("Encoded URL is %q\n", url.String())
						client := &http.Client{}
						request, err := http.NewRequest("POST", url.String(), nil)
						if err != nil {
							fmt.Println(err)
							return err
							
						}
						request.Header.Set("Content-Type", "text/plain")
						response, err := client.Do(request)
						if err != nil {
							fmt.Println(err)
							return err
							
						}
						response.Body.Close()
						// response status
						fmt.Println("Status code ", response.Status)
						
				}
				
			}	
			
		}
		
	}
}


func main() {
	err := ReportUpdate(2,10)
	if err != nil {fmt.Println(err)
		}

}