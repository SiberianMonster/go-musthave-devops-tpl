package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/generalutils"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/storage"
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
)

var err error
var resp map[string]string
var testHash string

type WrapperJSONStruct struct {
    Hashkey string
	DB *sql.DB
}

func (ws WrapperJSONStruct) UpdateJSONHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Connection", "close")
	var updateParams generalutils.Metrics

	err := json.NewDecoder(r.Body).Decode(&updateParams)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp["status"] = "missing json body"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	defer r.Body.Close()

	if updateParams.MType != generalutils.Counter && updateParams.MType != generalutils.Gauge {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "invalid type"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	if len(ws.Hashkey) > 0 {
		if updateParams.MType == generalutils.Counter {
			testHash, err = generalutils.Hash(fmt.Sprintf("%s:counter:%d", updateParams.ID, *updateParams.Delta), ws.Hashkey)
			if err != nil {
				log.Fatalf("Error happened when hashing received value. Err: %s", err)
				return
			}
		} else {
			testHash, err = generalutils.Hash(fmt.Sprintf("%s:gauge:%f", updateParams.ID, *updateParams.Value), ws.Hashkey)
			if err != nil {
				log.Fatalf("Error happened when hashing received value. Err: %s", err)
				return
			}
		}
		
		if testHash != updateParams.Hash {
			log.Printf("Hashing values do not match. Value produced: %s. Value received: %s", testHash, updateParams.Hash)
			rw.WriteHeader(http.StatusBadRequest)
			resp["status"] = "received hash does not match"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
				return
			}
			rw.Write(jsonResp)
			return
		}
	}

	err = storage.RepositoryUpdate(updateParams, ws.DB)
	if err != nil {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "update failed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	s, err := json.Marshal(generalutils.Container)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	log.Print(string(s))
	rw.WriteHeader(http.StatusOK)
	resp["status"] = "ok"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	rw.Write(jsonResp)

}

func (ws WrapperJSONStruct) UpdateStringHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")

	urlPart := mux.Vars(r)
	log.Printf(urlPart["name"])

	var structParams generalutils.Metrics
	fv, err := strconv.ParseFloat(urlPart["value"], 64)
	
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		resp["status"] = "wrong value"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	fieldType := urlPart["type"]
	if fieldType != generalutils.Counter && fieldType != generalutils.Gauge {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "missing type"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	if fieldType == generalutils.Counter {
		fvCounter := int64(fv)
		structParams = generalutils.Metrics{ID: urlPart["name"], MType: urlPart["type"], Delta: &fvCounter}
	} else {
		structParams = generalutils.Metrics{ID: urlPart["name"], MType: urlPart["type"], Value: &fv}
	}

	err = storage.RepositoryUpdate(structParams, ws.DB)
	if err != nil {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "update failed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	rw.WriteHeader(http.StatusOK)
	resp["status"] = "ok"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	rw.Write(jsonResp)

}

func (ws WrapperJSONStruct) ValueJSONHandler(rw http.ResponseWriter, r *http.Request) {

	resp = make(map[string]string)
	rw.Header().Set("Content-Type", "application/json")
	var receivedParams generalutils.Metrics

	err := json.NewDecoder(r.Body).Decode(&receivedParams)
	
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp["status"] = "missing json body"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	defer r.Body.Close()

	if _, ok := generalutils.Container[receivedParams.ID]; !ok {
		log.Printf("missing params value")
		receivedS, err := json.Marshal(receivedParams)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		log.Print(string(receivedS))
		rw.WriteHeader(http.StatusNotFound)
		resp["status"] = "missing parameter"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	if receivedParams.MType != generalutils.Counter && receivedParams.MType != generalutils.Gauge {
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "invalid type"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	retrievedMetrics, getErr := storage.RepositoryRetrieve(receivedParams)
	if len(ws.Hashkey) > 0 {
		if retrievedMetrics.MType == generalutils.Counter {
			log.Println("Retrieving hash value")
			retrievedMetrics.Hash, err = generalutils.Hash(fmt.Sprintf("%s:counter:%d", retrievedMetrics.ID, *retrievedMetrics.Delta), ws.Hashkey)
			if err != nil {
				log.Fatalf("Error happened when hashing received value. Err: %s", err)
				return
			}
		} else {
			retrievedMetrics.Hash, err = generalutils.Hash(fmt.Sprintf("%s:gauge:%f", retrievedMetrics.ID, *retrievedMetrics.Value), ws.Hashkey)
			log.Println("Retrieving hash value")
			if err != nil {
				log.Fatalf("Error happened when hashing received value. Err: %s", err)
				return
			}
		}
	}
	log.Println(retrievedMetrics)
	if getErr != nil {
		rw.WriteHeader(http.StatusNotFound)
		resp["status"] = "value retrieval failed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
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

	if _, ok := generalutils.Container[params]; !ok {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusNotFound)
		resp["status"] = "missing parameter"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}
	if fieldType != generalutils.Counter && fieldType != generalutils.Gauge {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusNotImplemented)
		resp["status"] = "invalid type"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
	}

	var structParams = generalutils.Metrics{ID: params, MType: urlPart["name"]}
	retrievedMetrics, getErr := storage.RepositoryRetrieveString(structParams)

	if getErr != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusNotFound)
		resp["status"] = "value retrieval failed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
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
	s, err := json.Marshal(generalutils.Container)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
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
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}
		rw.Write(jsonResp)
		return
    }

	rw.WriteHeader(http.StatusOK)
	resp["status"] = "ok"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	rw.Write(jsonResp)
}

