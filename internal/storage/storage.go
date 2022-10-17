package storage

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/config"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	_ "github.com/lib/pq"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

var err error

func RepositoryUpdate(mp metrics.Metrics, storeDB *sql.DB, dbFlag bool) error {

	smp, err := json.Marshal(mp)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return err
	}
	log.Print(string(smp))

	if dbFlag {
		DBSave(storeDB, mp)
		return nil
	}
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
	}
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

func RepositoryRetrieve(mp metrics.Metrics, storeDB *sql.DB, dbFlag bool) (metrics.Metrics, error) {

	if dbFlag {
		return DBUpload(storeDB, mp)

	} else {
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

}

func RepositoryRetrieveString(mp metrics.Metrics, storeDB *sql.DB, dbFlag bool) (string, error) {

	var requestedValue string
	if dbFlag {
		mp, err = DBUpload(storeDB, mp)
		if err != nil {
			log.Printf("Error happened in retrieveing metrics. Err: %s", err)
			return requestedValue, err
		}

		if mp.Delta != nil {
			requestedValue = fmt.Sprintf("%v", mp.Delta)
			return requestedValue, err
		} else {
			return fmt.Sprintf("%v", mp.Value), err
		}

	} else {
		v := reflect.ValueOf(mp)
		fieldName, ok := v.Field(0).Interface().(string)
		if !ok {
			err = errors.New("failed metrics retrieval")
			log.Printf("Error happened in validating metrics name. Err: %s", err)
			return requestedValue, err
		}
		requestedValue = fmt.Sprintf("%v", metrics.Container[fieldName])

		return requestedValue, nil
	}
}

func StaticFileSave(storeFile string) {

	file, err := os.OpenFile(storeFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatalf("Error happened in JSON file opening. Err: %s", err)
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
			log.Fatalf("Error happened when writing data to storage file. Err: %s", err)
		}

		if err := writer.WriteByte('\n'); err != nil {
			log.Fatalf("Error happened when writing data to storage file. Err: %s", err)
		}
		writer.Flush()
	}
	file.Close()
	log.Printf("saved JSON to file")

}

func StaticFileUpload(storeFile string) {

	file, err := os.OpenFile(storeFile, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatalf("Error happened in JSON file opening. Err: %s", err)

	} else {
		log.Printf("Uploading data from JSON")
		reader := bufio.NewReader(file)
		data, err := reader.ReadBytes('\n')
		if err != nil {
			log.Printf("Error happened in reading JSON file bytes. Err: %s", err)

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

func DBSave(storeDB *sql.DB, metricsObj metrics.Metrics) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// не забываем освободить ресурс
	defer cancel()

	_, err := storeDB.ExecContext(ctx, "INSERT INTO metrics (name, value, delta) VALUES ($1, $2, $3);",
		metricsObj.ID,
		metricsObj.Value,
		metricsObj.Delta,
	)
	if err != nil {
		log.Printf("Error happened when inserting a new entry into sql table. Err: %s", err)
		return
	}
	log.Printf("saved metrics data to DB")
}

func DBSaveBatch(storeDB *sql.DB, metricsObj []metrics.Metrics) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// не забываем освободить ресурс
	defer cancel()

	// шаг 1 — объявляем транзакцию
	tx, err := storeDB.Begin()
	if err != nil {
		log.Printf("Error happened when initiating sql transaction. Err: %s", err)
		return err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback()

	// шаг 2 — готовим инструкцию
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO metrics (name, value, delta) VALUES ($1, $2, $3)")
	if err != nil {
		log.Printf("Error happened when preparing sql transaction context. Err: %s", err)
		return err
	}
	// шаг 2.1 — не забываем закрыть инструкцию, когда она больше не нужна
	defer stmt.Close()

	for _, v := range metricsObj {
		// шаг 3 — указываем, что каждое будет добавлено в транзакцию
		if _, err = stmt.ExecContext(ctx, v.ID, v.Value, v.Delta); err != nil {
			log.Printf("Error happened when declaring transaction. Err: %s", err)
			return err
		}
	}
	// шаг 4 — сохраняем изменения

	return tx.Commit()
}

func DBUpload(storeDB *sql.DB, metricsObj metrics.Metrics) (metrics.Metrics, error) {

	ctx, cancel := context.WithTimeout(context.Background(), config.ContextDBTimeout*time.Second)
	// не забываем освободить ресурс
	defer cancel()
	var uploadedValue *float64
	var uploadedDelta *int64

	err := storeDB.QueryRowContext(ctx, "WITH ranked_metrics AS (SELECT m.*, ROW_NUMBER() OVER (PARTITION BY name ORDER BY metrics_id DESC) AS rn FROM metrics AS m) SELECT value FROM ranked_metrics WHERE name = ($1) AND rn = 1;", metricsObj.ID).Scan(&uploadedValue)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error at first query")
		log.Printf("Error happened when extracting delta entry from sql table. Err: %s", err)
		return metricsObj, err
	}
	if uploadedValue == nil {
		err := storeDB.QueryRowContext(ctx, "SELECT  SUM(delta) FROM metrics WHERE name=($1);", metricsObj.ID).Scan(&uploadedDelta)
		log.Println("Error at second query")
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error happened when extracting delta entry from sql table. Err: %s", err)
			return metricsObj, err
		}
		metricsObj.Delta = uploadedDelta
	} else {
		metricsObj.Value = uploadedValue
	}
	log.Printf("uploaded data from DB")
	s, err := json.Marshal(metricsObj)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		return metricsObj, err
	}
	log.Print(string(s))
	return metricsObj, nil
}

func ContainerUpdate(storeInt int, storeFile string, storeDB *sql.DB, storeParameter string) {

	var ticker *time.Ticker
	if strings.Contains(storeParameter, "m") {
		ticker = time.NewTicker(time.Duration(storeInt) * 100 * time.Millisecond)
	} else {
		ticker = time.NewTicker(time.Duration(storeInt) * time.Second)
	}

	for range ticker.C {

		StaticFileSave(storeFile)

	}

}
