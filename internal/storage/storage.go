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
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/generalutils"
	"database/sql"
	_ "github.com/lib/pq"
	"context"
)

var err error

func RepositoryUpdate(mp generalutils.Metrics) error {

	smp, err := json.Marshal(mp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		return err
	}
	log.Print(string(smp))
	v := reflect.ValueOf(mp)
	var newValue float64
	var newDelta int64
	var oldDelta int64
	fieldName := v.Field(0).Interface().(string)
	fieldType := v.Field(1).Interface().(string)

	if fieldType == generalutils.Counter {
		newDelta = *mp.Delta
		log.Printf("New counter %d\n", newDelta)
		if _, ok := generalutils.Container[fieldName]; ok {
			if _, ok := generalutils.Container[fieldName].(float64); ok {
				valOld, ok := generalutils.Container[fieldName].(float64)
				if !ok {
					err = errors.New("failed metrics retrieval")
					log.Printf("Error happened in reading metrics from loaded storage. Metrics: %s Err: %s", fieldName, err)
					return err
				}
				oldDelta = int64(valOld)
			} else {
				oldDelta, ok = generalutils.Container[fieldName].(int64)
				if !ok {
					err = errors.New("failed metrics retrieval")
					log.Printf("Error happened in reading container metrics. Metrics: %s Err: %s", fieldName, err)
					return err
				}
			}
			newDelta = oldDelta + newDelta
		}
		generalutils.Container[fieldName] = newDelta
	} else {
		newValue = *mp.Value
		log.Printf("New gauge %f\n", newValue)
		generalutils.Container[fieldName] = newValue
	}

	return nil

}

func RepositoryRetrieve(mp generalutils.Metrics) (generalutils.Metrics, error) {

	v := reflect.ValueOf(mp)
	var delta int64

	fieldName, ok := v.Field(0).Interface().(string)
	if !ok {
		err = errors.New("failed metrics retrieval")
		log.Fatalf("Error happened in validating metrics name. Err: %s", err)
		return mp, err
	}
	fieldType, ok := v.Field(1).Interface().(string)
	if !ok {
		err = errors.New("failed metrics retrieval")
		log.Fatalf("Error happened in validating metrics type. Err: %s", err)
		return mp, err
	}

	if fieldType == generalutils.Counter {
		if _, ok := generalutils.Container[fieldName].(float64); ok {
			valOld := generalutils.Container[fieldName].(float64)
			delta = int64(valOld)
		} else {
			delta = generalutils.Container[fieldName].(int64)
		}
		mp.Delta = &delta
	} else {
		value := generalutils.Container[fieldName].(float64)
		mp.Value = &value
	}

	return mp, nil

}

func RepositoryRetrieveString(mp generalutils.Metrics) (string, error) {

	v := reflect.ValueOf(mp)
	var requestedValue string
	fieldName, ok := v.Field(0).Interface().(string)
	if !ok {
		err = errors.New("failed metrics retrieval")
		log.Printf("Error happened in validating metrics type. Err: %s", err)
		return requestedValue, err
	}
	requestedValue = fmt.Sprintf("%v", generalutils.Container[fieldName])

	return requestedValue, nil

}

func StaticFileSave(storeFile string) {

	file, err := os.OpenFile(storeFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatalf("Error happened in JSON file opening. Err: %s", err)
		return
	}
	writer := bufio.NewWriter(file)

	data, err := json.Marshal(&generalutils.Container)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		return
	}
	if len(data) > 3 {
		log.Print(string(data))
		if _, err := writer.Write(data); err != nil {
			log.Fatalf("Error happened when writing data to storage file. Err: %s", err)
			return
		}

		if err := writer.WriteByte('\n'); err != nil {
			log.Fatalf("Error happened when writing data to storage file. Err: %s", err)
			return
		}
		writer.Flush()
	}
	file.Close()
	log.Printf("saved JSON to file")

}

func StaticFileUpload(storeFile string, restore bool) {

	if restore {
		file, err := os.OpenFile(storeFile, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			log.Fatalf("Error happened in JSON file opening. Err: %s", err)
			return

		} else {
			log.Printf("Uploading data from JSON")
			reader := bufio.NewReader(file)
			data, err := reader.ReadBytes('\n')
			if err != nil {
				log.Printf("Error happened in reading JSON file bytes. Err: %s", err)
				return
			} else {
				err = json.Unmarshal([]byte(data), &generalutils.Container)
				log.Print(string(data))
				if err != nil {
					log.Printf("No JSON data to decode")
				}
			}
			file.Close()
		}

	}
}

func DBSave(storeDB *sql.DB) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// не забываем освободить ресурс
		defer cancel()

	for fieldName := range generalutils.Container { 

		if _, ok := generalutils.Container[fieldName].(float64); ok {
			value := generalutils.Container[fieldName].(float64)
			_, err := storeDB.ExecContext(ctx,
				"INSERT INTO metrics (name, value) VALUES ($1, $2)",
				fieldName,
				value,
			)
			if err != nil {
				log.Printf("Error happened when inserting a new entry into sql table. Err: %s", err)
				return
			}
		} else {
			delta := generalutils.Container[fieldName].(int64)
			_, err := storeDB.ExecContext(ctx,
				"INSERT INTO metrics (name, delta) VALUES ($1, $2)",
				fieldName,
				delta,
			)
			if err != nil {
				log.Printf("Error happened when inserting a new entry into sql table. Err: %s", err)
				return
			}

		}
	}
	log.Printf("saved container data to DB")
}

func DBUpload(storeDB *sql.DB, restore bool) {

	if restore {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// не забываем освободить ресурс
		defer cancel()

		latestMetrics, err := storeDB.QueryContext(ctx, "WITH ranked_metrics AS (SELECT m.*, ROW_NUMBER() OVER (PARTITION BY name ORDER BY metrics_id DESC) AS rn FROM metrics AS m) SELECT * FROM ranked_metrics WHERE rn = 1;")
		if err != nil {
			log.Fatalf("Error happened when extracting entries from sql table. Err: %s", err)
			return
		}

		defer func() {
			_ = latestMetrics.Close()
			_ = latestMetrics.Err() 
		}()
		for latestMetrics.Next() {
			var row generalutils.Metrics
			if err := latestMetrics.Scan(&row.ID, &row.Delta, &row.Value); err != nil {
				log.Fatalf("Error happened when iterating over entries in sql table. Err: %s", err)
				return
			} else {
				if row.Delta != nil {
					generalutils.Container[row.ID] = row.Delta
				} else {
					generalutils.Container[row.ID] = row.Value
				}
			}
		}
		log.Printf("uploaded data from DB")
	}
}

func ContainerUpdate(storeInt int, storeFile string, storeDB *sql.DB, connStr string) {

	ticker := time.NewTicker(time.Duration(storeInt) * time.Second)

	for range ticker.C {
		
		if len(connStr) > 0 { 
			DBSave(storeDB)
		} else {
			StaticFileSave(storeFile)
		}
	}
}

