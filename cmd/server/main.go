package main

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"os"
	"bufio"
	"strconv"
	"time"
	"fmt"
	"flag"
	"io"
	"compress/gzip"
	"strings"
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
var host, storeInterval, storeFile, restore *string
var resp map[string]string

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getEnv(key string, fallback *string) *string {
    if value, ok := os.LookupEnv(key); ok {
        return &value
    }
    return fallback
}

type gzipWriter struct {
    http.ResponseWriter
    Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
    // w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
    return w.Writer.Write(b)
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

func StaticFileSave(storeFile string) {

	file, err := os.OpenFile(storeFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
			log.Fatal(err)
	}
	writer := bufio.NewWriter(file)
		
	data, err := json.Marshal(&Container)
	if err != nil {
			log.Fatal(err)
	}

	if _, err := writer.Write(data); err != nil {
			log.Fatal(err)
	}

	if err := writer.WriteByte('\n'); err != nil {
			log.Fatal(err)
	}
	writer.Flush()
	file.Close()
	log.Printf("saved json to file")
}


func StaticFileUpdate(storeInt int, storeFile string, restore bool) {

	if restore {
		file, err := os.OpenFile(storeFile, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Uploading data from json")
			reader := bufio.NewReader(file)
			data, err := reader.ReadBytes('\n')
			if err != nil {
			} else {
				err = json.Unmarshal([]byte(data), &Container)
				if err != nil {
					log.Printf("no json data to decode")
				}
			}
		file.Close()
		}

	}

	ticker := time.NewTicker(time.Duration(storeInt) * time.Second)

	for range ticker.C {
		StaticFileSave(storeFile)
	}
}

func valueStringHandler(rw http.ResponseWriter, r *http.Request) {
		
	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")

	params := r.URL.Query().Get("name")
	fieldType := r.URL.Query().Get("type")

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

	var structParams = Metrics{ID: params,  MType: r.URL.Query().Get("type")}
	retrievedMetrics, getErr := RepositoryRetrieveString(structParams)

	log.Println(retrievedMetrics)
	if getErr != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("value retrieval failed"))
			return
	}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(retrievedMetrics))
		

}
	
func updateHandler(rw http.ResponseWriter, r *http.Request) {
	
	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Connection", "close")
	var updateParams Metrics

	var reader io.Reader

    if r.Header.Get(`Content-Encoding`) == `gzip` {
        gz, err := gzip.NewReader(r.Body)
        if err != nil {
            http.Error(rw, err.Error(), http.StatusInternalServerError)
            return
        }
        reader = gz
        defer gz.Close()
    } else {
        reader = r.Body
    }

	err := json.NewDecoder(reader).Decode(&updateParams)

	if err != nil {
			urlPart := strings.Split(r.URL.Path, "/")
			log.Printf(urlPart[3])
			fv, err := strconv.ParseFloat(urlPart[4], 64)
			var structParams Metrics
			if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					resp["status"] = "wrong value"
					jsonResp, err := json.Marshal(resp)
					if err != nil {
						log.Fatalf("Error happened in JSON marshal. Err: %s", err)
					}
					rw.Write(jsonResp)
					return
			}
			fieldType := urlPart[2]
			if fieldType != "counter" && fieldType != "gauge" {
					rw.WriteHeader(http.StatusNotImplemented)
					resp["status"] = "missing type"
					jsonResp, err := json.Marshal(resp)
					if err != nil {
						log.Fatalf("Error happened in JSON marshal. Err: %s", err)
					}
					rw.Write(jsonResp)
					return
			}
			if fieldType == "counter" {
					fvCounter := int64(fv)
					structParams = Metrics{ID: urlPart[3], MType: urlPart[2], Delta: &fvCounter}
			} else {
					structParams = Metrics{ID: urlPart[3], MType: urlPart[2], Value: &fv}
			}
		
			err = RepositoryUpdate(structParams)
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
			rw.WriteHeader(http.StatusOK)
			resp["status"] = "ok"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
					log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			rw.Write(jsonResp)
	} else {
	
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
	}
}

func valueJsonHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Connection", "close")
	var receivedParams Metrics
	var reader io.Reader

    if r.Header.Get(`Content-Encoding`) == `gzip` {
        gz, err := gzip.NewReader(r.Body)
        if err != nil {
            http.Error(rw, err.Error(), http.StatusInternalServerError)
            return
        }
        reader = gz
        defer gz.Close()
    } else {
        reader = r.Body
    }

	err := json.NewDecoder(reader).Decode(&receivedParams)

	if err != nil {
			urlPart := strings.Split(r.URL.Path, "/")
			log.Printf(urlPart[3])

			params := urlPart[3]
			fieldType := urlPart[2]
		
			if _, ok := Container[params]; !ok {
					rw.WriteHeader(http.StatusNotFound)
					resp["status"] = "missing parameter"
					jsonResp, err := json.Marshal(resp)
					if err != nil {
						log.Fatalf("Error happened in JSON marshal. Err: %s", err)
					}
					rw.Write(jsonResp)
					return
			}
			if fieldType != "counter" && fieldType != "gauge" {
					rw.WriteHeader(http.StatusNotImplemented)
					resp["status"] = "invalid type"
					jsonResp, err := json.Marshal(resp)
					if err != nil {
						log.Fatalf("Error happened in JSON marshal. Err: %s", err)
					}
					rw.Write(jsonResp)
					return
			}
		
			var structParams = Metrics{ID: params,  MType: r.URL.Query().Get("type")}
			retrievedMetrics, getErr := RepositoryRetrieveString(structParams)
		
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

	} else {
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
	}
		
}

func genericHandler (rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")

	log.Printf("Got to generic endpoint")
	s, _ := json.Marshal(Container)
	log.Print(string(s))
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(string(s)))
}
	

func gzipHandler(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            h.ServeHTTP(w, r)
            return
        }
        w.Header().Set("Content-Encoding", "gzip")
        gz := gzip.NewWriter(w)
        defer gz.Close()
        h.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
    })
}

func init() {

	Container = make (map[string]interface{})

	host = getEnv("ADDRESS", flag.String("a", "127.0.0.1:8080", "ADDRESS"))
	storeInterval = getEnv("STORE_INTERVAL", flag.String("i", "300", "STORE_INTERVAL"))
	storeFile = getEnv("STORE_FILE", flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE"))
	restore = getEnv("RESTORE", flag.String("r", "false", "RESTORE"))

}

func main() {

	flag.Parse()

	restoreValue, err := strconv.ParseBool(*restore)
	if err != nil {
		log.Fatal(err)
	}

	storeInt, err := strconv.Atoi(*storeInterval)
    if err != nil {
        log.Fatal(err)
    }

	go StaticFileUpdate(storeInt, *storeFile, restoreValue)

	mux := http.NewServeMux()

	mux.HandleFunc("/update/", updateHandler)
	mux.HandleFunc("/value/", valueJsonHandler)
	
	mux.HandleFunc("/value/{type}/{name}", valueStringHandler)
	mux.HandleFunc("/", genericHandler)

	httpMux := reflect.ValueOf(mux).Elem()
	finList := httpMux.FieldByIndex([]int{1})
	fmt.Println(finList)

	// запуск сервера с адресом localhost, порт 8080
	//server := &http.Server{
	//    Addr: "127.0.0.1:8080",
	//}
	if err := http.ListenAndServe(*host, gzipHandler(mux)); err != nil {
        log.Fatal(err)
    }
	StaticFileSave(*storeFile)
}
