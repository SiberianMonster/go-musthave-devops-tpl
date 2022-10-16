package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
)

var err error

func RepositoryUpdate(mp metrics.Metrics) error {

	smp, err := json.Marshal(mp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return err
	}
	log.Print(string(smp))
	v := reflect.ValueOf(mp)
	var newValue float64
	var newDelta int64
	var oldDelta int64
	fieldName := v.Field(0).Interface().(string)
	fieldType := v.Field(1).Interface().(string)

	if fieldType != metrics.Counter {
		newValue = *mp.Value
		log.Printf("New gauge %f\n", newValue)
		metrics.Container[fieldName] = newValue
		return nil

	} else {
		newDelta = *mp.Delta
		log.Printf("New counter %d\n", newDelta)
		if _, ok := metrics.Container[fieldName]; ok {
			if _, ok := metrics.Container[fieldName].(float64); ok {
				valOld, ok := metrics.Container[fieldName].(float64)
				if !ok {
					err = errors.New("failed metrics retrieval")
					log.Printf("Error happened in reading metrics from loaded storage. Metrics: %s Err: %s", fieldName, err)
					return err
				}
				oldDelta = int64(valOld)
			} else {
				oldDelta, ok = metrics.Container[fieldName].(int64)
				if !ok {
					err = errors.New("failed metrics retrieval")
					log.Printf("Error happened in reading container metrics. Metrics: %s Err: %s", fieldName, err)
					return err
				}
			}
			newDelta = oldDelta + newDelta
		}
		metrics.Container[fieldName] = newDelta
		return nil
	}

}

func RepositoryRetrieve(mp metrics.Metrics) (metrics.Metrics, error) {

	v := reflect.ValueOf(mp)
	var delta int64

	fieldName, ok := v.Field(0).Interface().(string)
	if !ok {
		err = errors.New("failed metrics retrieval")
		log.Printf("Error happened in validating metrics name. Err: %s", err)
		return mp, err
	}
	fieldType, ok := v.Field(1).Interface().(string)
	if !ok {
		err = errors.New("failed metrics retrieval")
		log.Printf("Error happened in validating metrics type. Err: %s", err)
		return mp, err
	}

	if fieldType == metrics.Counter {
		if _, ok := metrics.Container[fieldName].(float64); ok {
			valOld := metrics.Container[fieldName].(float64)
			delta = int64(valOld)
		} else {
			delta = metrics.Container[fieldName].(int64)
		}
		mp.Delta = &delta
	} else {
		value := metrics.Container[fieldName].(float64)
		mp.Value = &value
	}

	return mp, nil

}

func RepositoryRetrieveString(mp metrics.Metrics) (string, error) {

	v := reflect.ValueOf(mp)
	var requestedValue string
	fieldName, ok := v.Field(0).Interface().(string)
	if !ok {
		err = errors.New("failed metrics retrieval")
		log.Printf("Error happened in validating metrics type. Err: %s", err)
		return requestedValue, err
	}
	requestedValue = fmt.Sprintf("%v", metrics.Container[fieldName])

	return requestedValue, nil

}

func StaticFileSave(storeFile string) {

	file, err := os.OpenFile(storeFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Printf("Error happened in JSON file opening. Err: %s", err)
		return
	}
	writer := bufio.NewWriter(file)

	data, err := json.Marshal(&metrics.Container)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	if len(data) > 3 {
		log.Print(string(data))
		if _, err := writer.Write(data); err != nil {
			log.Printf("Error happened when writing data to storage file. Err: %s", err)
			return
		}

		if err := writer.WriteByte('\n'); err != nil {
			log.Printf("Error happened when writing data to storage file. Err: %s", err)
			return
		}
		writer.Flush()
	}
	file.Close()
	log.Printf("saved JSON to file")

}

func StaticFileUpload(storeFile string) {

		file, err := os.OpenFile(storeFile, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			log.Printf("Error happened in JSON file opening. Err: %s", err)
			return

		} else {
			log.Printf("Uploading data from JSON")
			reader := bufio.NewReader(file)
			data, err := reader.ReadBytes('\n')
			if err != nil {
				log.Printf("Error happened in reading JSON file bytes. Err: %s", err)
				return
			} else {
				err = json.Unmarshal([]byte(data), &metrics.Container)
				log.Print(string(data))
				if err != nil {
					log.Printf("No JSON data to decode")
				}
			}
			file.Close()
		}
}

func StaticFileUpdate(storeInt int, storeFile string) {

	ticker := time.NewTicker(time.Duration(storeInt) * time.Second)

	for range ticker.C {
		if len(storeFile) > 0 {
			StaticFileSave(storeFile)
		}
	}
}

