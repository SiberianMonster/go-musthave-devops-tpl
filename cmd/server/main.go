package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"reflect"
	"strconv"
)

type gauge float64
type counter int64

type UpdateMetrics struct {
	StructKey   string
	StructValue float64
	StructType  string
}

type GetMetrics struct {
	StructKey string
}

type metricsContainer struct {
	Container map[string]interface{}
}

var sharedMetrics metricsContainer
var err error

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func RepositoryUpdate(m metricsContainer, mp UpdateMetrics) (metricsContainer, error) {

	v := reflect.ValueOf(mp)
	var newValue float64
	var newCvalue counter
	var newGvalue gauge
	fieldName, _ := v.Field(0).Interface().(string)
	newValue, _ = v.Field(1).Interface().(float64)
	fieldType, _ := v.Field(2).Interface().(string)

	if fieldType == "counter" {
		if val, ok := m.Container[fieldName]; ok {
			newCvalue = val.(counter) + counter(newValue)
		} else {
			newCvalue = counter(newValue)
		}
		log.Println(newCvalue)
		log.Println(v.Field(1).Interface())
		m.Container[fieldName] = newCvalue
	} else {
		newGvalue = gauge(newValue)
		log.Println(newGvalue)
		log.Println(v.Field(1).Interface())
		m.Container[fieldName] = newGvalue
	}

	return m, nil

}

func RepositoryRetrieve(m metricsContainer, mp GetMetrics) (string, error) {

	v := reflect.ValueOf(mp)
	var requestedValue string
	fieldName, _ := v.Field(0).Interface().(string)
	requestedValue = fmt.Sprintf("%v", m.Container[fieldName])

	return requestedValue, nil

}

func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	sharedMetrics.Container = make(map[string]interface{})

	r.HandleFunc("/update/{type}/{name}/{value}", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		fv, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("wrong value"))
			return
		}
		fieldType := chi.URLParam(r, "type")
		if fieldType != "counter" && fieldType != "gauge" {
			rw.WriteHeader(http.StatusNotImplemented)
			rw.Write([]byte("wrong value"))
			return
		}
		var structParams = UpdateMetrics{StructKey: chi.URLParam(r, "name"), StructValue: fv, StructType: chi.URLParam(r, "type")}

		sharedMetrics, err = RepositoryUpdate(sharedMetrics, structParams)
		if err != nil {
			rw.WriteHeader(http.StatusNotImplemented)
			rw.Write([]byte("invalid type"))
			return
		}
		s, _ := json.Marshal(sharedMetrics)
		log.Print(string(s))
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/value/{type}/{name}", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=UTF-8")

		params := chi.URLParam(r, "name")

		var requestParams = GetMetrics{StructKey: params}
		if _, ok := sharedMetrics.Container[params]; !ok {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("missing parameter"))
			return
		}

		retrievedMetrics, getErr := RepositoryRetrieve(sharedMetrics, requestParams)
		log.Println(retrievedMetrics)
		if getErr != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("wrong parameter"))
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(retrievedMetrics))

	})

	r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=UTF-8")

		s, _ := json.Marshal(sharedMetrics)
		log.Print(string(s))
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(string(s)))
	})
	return r
}

func main() {

	r := NewRouter()

	// запуск сервера с адресом localhost, порт 8080
	//server := &http.Server{
	//    Addr: "127.0.0.1:8080",
	//}
	log.Fatal(http.ListenAndServe(":8080", r))
}
