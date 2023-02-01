// Metrics package contains pre-defined structs for system metrics storage.
//
// Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics
package metrics

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"

	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/httpp"
)

// Container object allows collection of dynamic system metrics.
var Container map[string]interface{}

// System metrics may belong to either counter or gauge type where "counter" is always an integer and "gauge" is a float value.
const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics struct is used for storing system metrics and also for exchanging data between the data collecting agent and the server.
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

// MetricsContainer struct has all the system metrics available from the runtime.ReadMemStats and the update counter.
type MetricsContainer struct {
	PollCount int64
	RandomValue,
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
	NumForcedGC,
	TotalMemory,
	FreeMemory,
	CPUutilization1,
	NumGC float64
}

// MetricsUpdate function is used to populate MetricsContainer with the system metrics from runtime.ReadMemStats.
func MetricsUpdate(m MetricsContainer, rtm runtime.MemStats) MetricsContainer {

	m.Alloc = float64(rtm.Alloc)
	m.BuckHashSys = float64(rtm.BuckHashSys)
	m.Frees = float64(rtm.Frees)
	m.GCCPUFraction = float64(rtm.GCCPUFraction)
	m.GCSys = float64(rtm.GCSys)
	m.HeapAlloc = float64(rtm.HeapAlloc)
	m.HeapIdle = float64(rtm.HeapIdle)
	m.HeapInuse = float64(rtm.HeapInuse)
	m.HeapObjects = float64(rtm.HeapObjects)
	m.HeapReleased = float64(rtm.HeapReleased)
	m.HeapSys = float64(rtm.HeapSys)
	m.LastGC = float64(rtm.LastGC)
	m.Lookups = float64(rtm.Lookups)
	m.MCacheInuse = float64(rtm.MCacheInuse)
	m.MCacheSys = float64(rtm.MCacheSys)
	m.MSpanInuse = float64(rtm.MSpanInuse)
	m.MSpanSys = float64(rtm.MSpanSys)
	m.Mallocs = float64(rtm.Mallocs)
	m.NextGC = float64(rtm.NextGC)
	m.NumForcedGC = float64(rtm.NumForcedGC)
	m.NumGC = float64(rtm.NumGC)
	m.OtherSys = float64(rtm.OtherSys)
	m.PauseTotalNs = float64(rtm.PauseTotalNs)
	m.StackInuse = float64(rtm.StackInuse)
	m.StackSys = float64(rtm.StackSys)
	m.Sys = float64(rtm.Sys)
	m.TotalAlloc = float64(rtm.TotalAlloc)
	m.PollCount += 1
	m.RandomValue = rand.Float64()
	return m

}

// MetricsHash function allows to hash the Metrics struct with system metrics using http.Hash algorythm.
func MetricsHash(m Metrics, key string) string {

	var strHash string
	var err error
	if m.MType == Counter {
		strHash, err = httpp.Hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), key)
		if err != nil {
			log.Printf("Error happened when hashing received value. Err: %s", err)
		}
	} else {
		strHash, err = httpp.Hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), key)
		if err != nil {
			log.Printf("Error happened when hashing received value. Err: %s", err)
		}
	}
	return strHash
}
