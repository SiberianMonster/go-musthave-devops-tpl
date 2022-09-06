package main

import (
    "net/http"
    "strconv"
    "log"
    "fmt"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "encoding/json"
    "reflect"
	"errors"
)

type gauge float64
type counter int64

type UpdateMetrics struct {
	StructKey string
	StructValue float64
}

type GetMetrics struct {
	StructKey string
}

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


func RepositoryUpdate(m metricsContainer, mp UpdateMetrics ) (metricsContainer, error) {

	var v reflect.Value
	v = reflect.ValueOf(mp)
	var new_value float64
	var new_cvalue counter
	var new_gvalue gauge
	fieldName, _ := v.Field(0).Interface().(string)
	new_value, _ = v.Field(1).Interface().(float64)

	t := reflect.TypeOf(m)

	names := make([]string, t.NumField())
	for i := range names {
		names[i] = t.Field(i).Name
	}
	if stringInSlice(fieldName, names) {
		if fieldName == "PollCount" {
			new_cvalue = m.PollCount + counter(new_value)
			fmt.Println(new_cvalue)
			fmt.Println(v.Field(1).Interface())
			reflect.ValueOf(&m).Elem().FieldByName(fieldName).Set(reflect.ValueOf(new_cvalue))
		} else {
			new_gvalue = gauge(new_value)
			fmt.Println(new_gvalue)
			fmt.Println(v.Field(1).Interface())
			reflect.ValueOf(&m).Elem().FieldByName(fieldName).Set(reflect.ValueOf(new_gvalue))
		}
	} else {
		return m, errors.New("missing field")
	}
	return m, nil

}

func RepositoryRetrieve(m metricsContainer, mp GetMetrics ) (string, error) {

	var v reflect.Value
	v = reflect.ValueOf(mp)
	var requested_value string
	fieldName, _ := v.Field(0).Interface().(string)

	t := reflect.TypeOf(m)

	names := make([]string, t.NumField())
	for i := range names {
		names[i] = t.Field(i).Name
	}
	if stringInSlice(fieldName, names) {
		requested_value = fmt.Sprintf("%v", reflect.ValueOf(&m).Elem().FieldByName(fieldName).Interface())
	} else {
		return requested_value, errors.New("missing field")
	}
	return requested_value, nil

}

func NewRouter() chi.Router {
    r := chi.NewRouter()
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    
    r.HandleFunc("/update/{type}/{name}/{value}", func(rw http.ResponseWriter, r *http.Request) {
        rw.Header().Set("Content-Type", "application/json")
        fv, err := strconv.ParseFloat(chi.URLParam(r, "value"), 8)
        if err != nil {
            rw.WriteHeader(http.StatusNotFound)
            rw.Write([]byte("wrong value"))
            return
        }
        var struct_params = UpdateMetrics{StructKey: chi.URLParam(r, "name"), StructValue: fv}

        sharedMetrics , err = RepositoryUpdate(sharedMetrics, struct_params)
        if err != nil {
            rw.WriteHeader(http.StatusNotFound)
            rw.Write([]byte("wrong parameter"))
            return
        }
        s, _ := json.Marshal(sharedMetrics)
        fmt.Println(string(s))
        rw.WriteHeader(http.StatusOK)
        rw.Write([]byte(`{"status":"ok"}`))
    })

    r.Get("/value/{type}/{name}", func(rw http.ResponseWriter, r *http.Request) {
        rw.Header().Set("Content-Type", "text/html; charset=UTF-8")

        params := chi.URLParam(r, "name")

        var request_params = GetMetrics{StructKey: params}

        retrievedMetrics , get_err := RepositoryRetrieve(sharedMetrics, request_params)
        fmt.Println(retrievedMetrics)
        if get_err != nil {
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
        fmt.Println(string(s))
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