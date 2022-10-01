package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"reflect"
	"os"
	"bufio"
	"strconv"
	"fmt"
)

type gauge float64
type counter int64


type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
} 

var err error
var Container map[string]interface{}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func RepositoryUpdate(mp Metrics) (error) {

	smp, _ := json.Marshal(mp)
	log.Print(string(smp))
	v := reflect.ValueOf(mp)
	var newValue float64
	var newDelta int64
	var oldDelta int64
	fieldName, _ := v.Field(0).Interface().(string)
	fieldType, _ := v.Field(1).Interface().(string)
	
	if fieldType == "counter" {
		newDelta = *mp.Delta
		log.Printf("New counter %d\n", newDelta)
		if _, ok := Container[fieldName]; ok {
			if _, ok := Container[fieldName].(float64); ok {
				valOld, _ := Container[fieldName].(float64)
				oldDelta = int64(valOld) 
			} else {
				oldDelta, _ = Container[fieldName].(int64)
			}
			newDelta = oldDelta + newDelta
		}
		Container[fieldName] = newDelta
	} else {
		newValue = *mp.Value
		log.Printf("New gauge %f\n", newValue)
		Container[fieldName] = newValue
	}

	return nil

}

func RepositoryRetrieve(mp Metrics) (Metrics, error) {

	v := reflect.ValueOf(mp)

	fieldName, _ := v.Field(0).Interface().(string)
	fieldType, _ := v.Field(1).Interface().(string)

	if fieldType == "counter" {
		delta := Container[fieldName].(int64)
		mp.Delta = &delta
	} else {
		value := Container[fieldName].(float64)
		mp.Value = &value
	}

	return mp, nil

}

func RepositoryRetrieveString(mp Metrics) (string, error) {

	v := reflect.ValueOf(mp)
	var requestedValue string
	fieldName, _ := v.Field(0).Interface().(string)
	requestedValue = fmt.Sprintf("%v", Container[fieldName])

	return requestedValue, nil

}




func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	resp := make(map[string]string)

	if Container == nil {
		Container = make (map[string]interface{})
	}

	r.HandleFunc("/update/{type}/{name}/{value}", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		
		fv, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		var structParams Metrics
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
		if fieldType == "counter" {
			fvCounter := int64(fv)
			structParams = Metrics{ID: chi.URLParam(r, "name"), MType: chi.URLParam(r, "type"), Delta: &fvCounter}
		} else {
			structParams = Metrics{ID: chi.URLParam(r, "name"), MType: chi.URLParam(r, "type"), Value: &fv}
		}

		err = RepositoryUpdate(structParams)
		if err != nil {
			rw.WriteHeader(http.StatusNotImplemented)
			rw.Write([]byte("update failed"))
			return
		}
		s, _ := json.Marshal(Container)
		log.Print(string(s))
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/value/{type}/{name}", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=UTF-8")

		params := chi.URLParam(r, "name")
		fieldType := chi.URLParam(r, "type")

		if _, ok := Container[params]; !ok {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("missing parameter"))
			return
		}
		if fieldType != "counter" && fieldType != "gauge" {
			rw.WriteHeader(http.StatusNotImplemented)
			rw.Write([]byte("wrong value"))
			return
		}

		var structParams = Metrics{ID: params,  MType: chi.URLParam(r, "type")}
		retrievedMetrics, getErr := RepositoryRetrieveString(structParams)

		log.Println(retrievedMetrics)
		if getErr != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("value retrieval failed"))
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(retrievedMetrics))
		

	})
	
	r.HandleFunc("/update/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Connection", "close")
		var updateParams Metrics

		err := json.NewDecoder(r.Body).Decode(&updateParams)

		if err != nil {
			log.Printf("Wrong params")
			rw.WriteHeader(http.StatusBadRequest)
			resp["status"] = "wrong value"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			rw.Write(jsonResp)
			return
		}
		defer r.Body.Close()

		if updateParams.MType != "counter" && updateParams.MType != "gauge" {
			rw.WriteHeader(http.StatusNotImplemented)
			resp["status"] = "invalid type"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			rw.Write(jsonResp)
			return
		}
		
		err = RepositoryUpdate(updateParams)
		if err != nil {
			rw.WriteHeader(http.StatusNotImplemented)
			resp["status"] = "update failed"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			rw.Write(jsonResp)
			return
		}
		s, _ := json.Marshal(Container)
		log.Print(string(s))
		rw.WriteHeader(http.StatusOK)
		resp["status"] = "ok"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		rw.Write(jsonResp)
	})

	r.HandleFunc("/value/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Connection", "close")
		var receivedParams Metrics

		err := json.NewDecoder(r.Body).Decode(&receivedParams)

		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			resp["status"] = "wrong request"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			rw.Write(jsonResp)
			return
		}
		defer r.Body.Close()

		if _, ok := Container[receivedParams.ID]; !ok {
			log.Printf("missing params value")
			receivedS, _ := json.Marshal(receivedParams)
			log.Print(string(receivedS))
			rw.WriteHeader(http.StatusNotFound)
			resp["status"] = "missing parameter"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			rw.Write(jsonResp)
			return
		}

		if receivedParams.MType != "counter" && receivedParams.MType != "gauge" {
			rw.WriteHeader(http.StatusNotImplemented)
			resp["status"] = "invalid type"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			rw.Write(jsonResp)
			return
		}

		retrievedMetrics, getErr := RepositoryRetrieve(receivedParams)
		log.Println(retrievedMetrics)
		if getErr != nil {
			rw.WriteHeader(http.StatusNotFound)
			resp["status"] = "value retrieval failed"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			rw.Write(jsonResp)
			return
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(retrievedMetrics)
		
	})

	r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=UTF-8")

		s, _ := json.Marshal(Container)
		log.Print(string(s))
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(string(s)))
	})
	return r
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}


func main() {

	Container = make (map[string]interface{})

	host := "127.0.0.1:8080"
	

	r := NewRouter()

	// запуск сервера с адресом localhost, порт 8080
	//server := &http.Server{
	//    Addr: "127.0.0.1:8080",
	//}
	log.Fatal(http.ListenAndServe(host, r))
}
