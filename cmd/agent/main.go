package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/generalutils"
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

	var m generalutils.MetricsContainer
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
			m = generalutils.MetricsUpdate(m, rtm)
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

				var metrics generalutils.Metrics

				if v.Field(i).Kind() == reflect.Float64 {
					metrics.ID = typeOfS.Field(i).Name
					metrics.MType = "gauge"
					value := v.Field(i).Interface().(float64)
					metrics.Value = &value

				} else {
					metrics.ID = typeOfS.Field(i).Name
					metrics.MType = "counter"
					delta := v.Field(i).Interface().(int64)
					metrics.Delta = &delta
				}

				body, err := json.Marshal(metrics)
				if err != nil {
					log.Fatalf("Error happened in JSON marshal. Err: %s", err)
					return err
				}
				log.Print(string(body))

				request, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(body))
				if err != nil {
					log.Fatalf("Error happened when request made. Err: %s", err)
					return err
				}

				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Accept", "application/json")
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

	host = generalutils.GetEnv("ADDRESS", flag.String("a", "127.0.0.1:8080", "ADDRESS"))
	pollCounterEnv = strings.Replace(*generalutils.GetEnv("POLL_INTERVAL", flag.String("p", "2", "POLL_INTERVAL")), "s", "", -1)
	reportCounterEnv = strings.Replace(*generalutils.GetEnv("REPORT_INTERVAL", flag.String("r", "10", "REPORT_INTERVAL")), "s", "", -1)

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
