package metricsfuncs

import (
	"time"
	"runtime"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"net/url"
	"net/http"
	"go-musthave-devops-tpl/internal/utils"
)


func MetricsUpdate(m utils.MetricsContainer, rtm runtime.MemStats ) utils.MetricsContainer {

	m.Alloc = utils.Gauge(rtm.Alloc)
	m.BuckHashSys = utils.Gauge(rtm.BuckHashSys)
	m.Frees = utils.Gauge(rtm.Frees)
	m.GCCPUFraction = utils.Gauge(rtm.GCCPUFraction)
	m.GCSys = utils.Gauge(rtm.GCSys)
	m.HeapAlloc = utils.Gauge(rtm.HeapAlloc)
	m.HeapIdle = utils.Gauge(rtm.HeapIdle)
	m.HeapInuse = utils.Gauge(rtm.HeapInuse)
	m.HeapObjects = utils.Gauge(rtm.HeapObjects)
	m.HeapReleased = utils.Gauge(rtm.HeapReleased)
	m.HeapSys = utils.Gauge(rtm.HeapSys)
	m.LastGC = utils.Gauge(rtm.LastGC)
	m.Lookups = utils.Gauge(rtm.Lookups)
	m.MCacheInuse = utils.Gauge(rtm.MCacheInuse)
	m.MCacheSys = utils.Gauge(rtm.MCacheSys)
	m.MSpanInuse = utils.Gauge(rtm.MSpanInuse)
	m.MSpanSys =utils.Gauge(rtm.MSpanSys)
	m.Mallocs = utils.Gauge(rtm.Mallocs)
	m.NextGC = utils.Gauge(rtm.NextGC)
	m.NumForcedGC = utils.Gauge(rtm.NumForcedGC)
	m.NumGC = utils.Gauge(rtm.NumGC)
	m.OtherSys = utils.Gauge(rtm.OtherSys)
	m.PauseTotalNs = utils.Gauge(rtm.PauseTotalNs)
	m.StackInuse = utils.Gauge(rtm.StackInuse)
	m.StackSys = utils.Gauge(rtm.StackSys)
	m.Sys = utils.Gauge(rtm.Sys)
	m.TotalAlloc = utils.Gauge(rtm.TotalAlloc)
	m.PollCount += 1
	m.RandomValue = utils.Gauge(rand.Float64())
	return m

}



func ReportUpdate(pollduration int, reportduration int) error {

	var m utils.MetricsContainer
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
						}
						request.Header.Set("Content-Type", "text/plain")
						response, err := client.Do(request)
						if err != nil {
							fmt.Println(err)
						}
						response.Body.Close()
						// response status
						fmt.Println("Status code ", response.Status)
						
				}
				
			}	
			
		}
		return err
	}
}

