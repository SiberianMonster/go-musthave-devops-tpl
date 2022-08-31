package main

import (
	"time"
	"runtime"
	"fmt"
	"math/rand"
	"reflect"
	"net/url"
	"net/http"
)

type gauge float64
type counter int64


type MetricsContainer struct {
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

func MetricsUpdate(m MetricsContainer, rtm runtime.MemStats ) MetricsContainer {

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



func ReportUpdate(pollduration int, reportduration int) {

	var m MetricsContainer
	var pollInterval = time.Duration(pollduration) * time.Second
	var reportInterval = time.Duration(reportduration) * time.Second
	var rtm runtime.MemStats
	var valueType string
	var v reflect.Value
	var typeOfS reflect.Type

	m.PollCount = 0

	for {
		<-time.After(pollInterval)

		runtime.ReadMemStats(&rtm)
		m = MetricsUpdate(m, rtm)
		v = reflect.ValueOf(m)
		typeOfS = v.Type()
	
		

		for {
			<-time.After(reportInterval)

			
			for i := 0; i< v.NumField(); i++ {

					url := url.URL{
						Scheme: "http",
						Host:   "127.0.0.1:8080",
					}
					
					url.Path += "update/"

					if typeOfS.Field(i).Name == "PollCount" {
						valueType = "counter/"
					} else {
						valueType = "gauge/"
					}
					url.Path += valueType
					url.Path += typeOfS.Field(i).Name
					url.Path += string('/')
					url.Path += fmt.Sprint(v.Field(i).Interface()) 
					fmt.Printf("Encoded URL is %q\n", url.String())
					client := &http.Client{}
					request, err := http.NewRequest("POST", url.String(), nil)
					if err != nil {
						fmt.Println(err)
					}
					request.Header.Set("Content-Type", "text/plain")
					response, err := client.Do(request)
					if err != nil {
						fmt.Println(err)
					}
					response.Body.Close()
					// печатаем код ответа
					fmt.Println("Статус-код ", response.Status)
					
			}
			
		}	
		
	}
}


func main() {
	ReportUpdate(2,10)

}