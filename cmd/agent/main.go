package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/config"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var host *string
var pollCounterEnv, reportCounterEnv string

func ReportUpdate(pollCounterVar int, reportCounterVar int) error {

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

				} else {
					metricsObj.ID = typeOfS.Field(i).Name
					metricsObj.MType = metrics.Counter
					delta := v.Field(i).Interface().(int64)
					metricsObj.Delta = &delta
				}

				body, err := json.Marshal(metricsObj)
				if err != nil {
					log.Printf("Error happened in JSON marshal. Err: %s", err)
					return err
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
	}
}

func init() {

	host = config.GetEnv("ADDRESS", flag.String("a", "127.0.0.1:8080", "ADDRESS"))
	pollCounterEnv = strings.Replace(*config.GetEnv("POLL_INTERVAL", flag.String("p", "2", "POLL_INTERVAL")), "s", "", -1)
	reportCounterEnv = strings.Replace(*config.GetEnv("REPORT_INTERVAL", flag.String("r", "10", "REPORT_INTERVAL")), "s", "", -1)

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

	err = ReportUpdate(pollCounterVar, reportCounterVar)
	if err != nil {
		log.Fatalf("Error happened in ReportUpdate. Err: %s", err)
	}

}
