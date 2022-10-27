package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go-musthave-devops-tpl/internal/config"
	"go-musthave-devops-tpl/internal/metrics"
	"go-musthave-devops-tpl/internal/storage"
	"log"
	"net/http"
	"strconv"
	"time"
)

var err error
var resp map[string]string
var testHash string

type WrapperJSONStruct struct {
	Key    string
	DB     *sql.DB
	DBFlag bool
}

func (ws WrapperJSONStruct) UpdateJSONHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Connection", "close")
	var updateParams metrics.Metrics

	err := json.NewDecoder(r.Body).Decode(&updateParams)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp["status"] = "missing json body"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	defer r.Body.Close()

	if updateParams.MType != metrics.Counter && updateParams.MType != metrics.Gauge {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "invalid type"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	if ws.Key != "" {

		testHash = metrics.MetricsHash(updateParams, ws.Key)

		if testHash != updateParams.Hash {
			log.Printf("Hashing values do not match. Value produced: %s. Value received: %s", testHash, updateParams.Hash)
			rw.WriteHeader(http.StatusBadRequest)
			resp["status"] = "received hash does not match"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Printf("Error happened in JSON marshal. Err: %s", err)
				return
			}
			rw.Write(jsonResp)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), config.ContextDBTimeout*time.Second)
	// не забываем освободить ресурс
	defer cancel()

	err = storage.RepositoryUpdate(updateParams, ws.DB, ws.DBFlag, ctx)
	if err != nil {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "update failed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	rw.WriteHeader(http.StatusOK)
	resp["status"] = "ok"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	rw.Write(jsonResp)

}

func (ws WrapperJSONStruct) UpdateStringHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")

	urlPart := mux.Vars(r)
	log.Printf(urlPart["name"])

	var structParams metrics.Metrics
	fv, err := strconv.ParseFloat(urlPart["value"], 64)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		resp["status"] = "wrong value"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	fieldType := urlPart["type"]
	if fieldType != metrics.Counter && fieldType != metrics.Gauge {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "missing type"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	if fieldType == metrics.Counter {
		fvCounter := int64(fv)
		structParams = metrics.Metrics{ID: urlPart["name"], MType: urlPart["type"], Delta: &fvCounter}
	} else {
		structParams = metrics.Metrics{ID: urlPart["name"], MType: urlPart["type"], Value: &fv}
	}

	ctx, cancel := context.WithTimeout(r.Context(), config.ContextDBTimeout*time.Second)
	// не забываем освободить ресурс
	defer cancel()

	err = storage.RepositoryUpdate(structParams, ws.DB, ws.DBFlag, ctx)
	if err != nil {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "update failed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	rw.WriteHeader(http.StatusOK)
	resp["status"] = "ok"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	rw.Write(jsonResp)

}

func (ws WrapperJSONStruct) UpdateBatchJSONHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Connection", "close")
	metricsBatch := []metrics.Metrics{}

	err := json.NewDecoder(r.Body).Decode(&metricsBatch)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp["status"] = "error when decoding batch"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), config.ContextDBTimeout*time.Second)
	// не забываем освободить ресурс
	defer cancel()

	err = storage.DBSaveBatch(ws.DB, metricsBatch, ctx)
	if err != nil {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "batch update failed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	s, err := json.Marshal(metrics.Container)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	log.Print(string(s))
	rw.WriteHeader(http.StatusOK)
	resp["status"] = "ok"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	rw.Write(jsonResp)

}

func (ws WrapperJSONStruct) ValueJSONHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")
	var receivedParams metrics.Metrics

	err := json.NewDecoder(r.Body).Decode(&receivedParams)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp["status"] = "missing json body"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), config.ContextDBTimeout*time.Second)
	// не забываем освободить ресурс
	defer cancel()

	var ok bool
	if ws.DBFlag {
		ok = storage.DBCheck(ws.DB, receivedParams.ID, ctx)

	} else {
		_, ok = metrics.Container[receivedParams.ID]
	}

	if !ok {
		log.Printf("missing params value")
		receivedS, err := json.Marshal(receivedParams)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		log.Print(string(receivedS))
		rw.WriteHeader(http.StatusNotFound)
		resp["status"] = "missing parameter"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	if receivedParams.MType != metrics.Counter && receivedParams.MType != metrics.Gauge {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "invalid type"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	retrievedMetrics, getErr := storage.RepositoryRetrieve(receivedParams, ws.DB, ws.DBFlag, ctx)

	if ws.Key != "" {

		retrievedMetrics.Hash = metrics.MetricsHash(retrievedMetrics, ws.Key)

	}
	log.Println(retrievedMetrics)
	if getErr != nil {
		rw.WriteHeader(http.StatusNotFound)
		resp["status"] = "value retrieval failed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(retrievedMetrics)
}

func (ws WrapperJSONStruct) ValueStringHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)

	urlPart := mux.Vars(r)
	log.Printf(urlPart["name"])

	params := urlPart["name"]
	fieldType := urlPart["type"]

	ctx, cancel := context.WithTimeout(r.Context(), config.ContextDBTimeout*time.Second)
	// не забываем освободить ресурс
	defer cancel()

	var ok bool
	if ws.DBFlag {

		ok = storage.DBCheck(ws.DB, params, ctx)

	} else {
		_, ok = metrics.Container[params]
	}

	if !ok {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusNotFound)
		resp["status"] = "missing parameter"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	if fieldType != metrics.Counter && fieldType != metrics.Gauge {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "invalid type"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	var structParams = metrics.Metrics{ID: params, MType: urlPart["name"]}

	retrievedMetrics, getErr := storage.RepositoryRetrieveString(structParams, ws.DB, ws.DBFlag, ctx)

	if getErr != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusNotFound)
		resp["status"] = "value retrieval failed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(retrievedMetrics))

}

func (ws WrapperJSONStruct) GenericHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	log.Printf("Got to generic endpoint")
	s, err := json.Marshal(metrics.Container)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	log.Print(string(s))
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(string(s)))
}

func (ws WrapperJSONStruct) PostgresHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")

	pingErr := ws.DB.Ping()
	if pingErr != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp["status"] = "failed connection to the database"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	rw.WriteHeader(http.StatusOK)
	resp["status"] = "ok"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	rw.Write(jsonResp)
}
