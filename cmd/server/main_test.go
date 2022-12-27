package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"log"

	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/handlers"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, path string, metricsObj metrics.Metrics) (*http.Response, string) {

	body, _ := json.Marshal(metricsObj)

	req, err := http.NewRequest(http.MethodPost, ts.URL+path, bytes.NewBuffer(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {

	r := mux.NewRouter()

	handlersWithKey := handlers.NewWrapperJSONStruct()

	r.HandleFunc("/update/", handlersWithKey.UpdateJSONHandler)
	r.HandleFunc("/value/", handlersWithKey.ValueJSONHandler)
	r.HandleFunc("/update/{type}/{name}/{value}", handlersWithKey.UpdateStringHandler)
	r.HandleFunc("/value/{type}/{name}", handlersWithKey.ValueStringHandler)

	r.HandleFunc("/", handlersWithKey.GenericHandler)
	r.Use(middleware.GzipHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()
	floatValue := 2.0

	metricsObj := metrics.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &floatValue,
	}

	resp, body := testRequest(t, ts, "/update/", metricsObj)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `{"status":"ok"}`, body)

	wrongMetrics := metrics.Metrics{
		ID:    "Alloc",
		MType: "othertype",
		Value: &floatValue,
	}

	resp, body = testRequest(t, ts, "/update/", wrongMetrics)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Equal(t, `{"status":"invalid type"}`, body)

}

func TestInitializeRouter(t *testing.T) {

	tests := []struct {
		name string
	}{
		{
			name: "trial run",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitializeRouter()

		})
	}
}

func TestParseStoreInterval(t *testing.T) {

	tests := []struct {
		name           string
		storeParameter string
		want           int
	}{
		{
			name:           "trial run",
			storeParameter: "3m",
			want:           3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ParseStoreInterval(&tt.storeParameter)
			assert.Equal(t, tt.want, v)
		})
	}
}

func TestParseRestoreValue(t *testing.T) {

	tests := []struct {
		name    string
		restore string
		want    bool
	}{
		{
			name:    "trial run",
			restore: "true",
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ParseRestoreValue(&tt.restore)
			assert.Equal(t, tt.want, v)
		})
	}
}

func TestShutdownGracefully(t *testing.T) {

	tests := []struct {
		name      string
		srv       *http.Server
		storeFile string
		connStr   string
	}{
		{
			name:      "trial run",
			storeFile: "/tmp/devops-metrics-db.json",
			connStr:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.srv = &http.Server{}
			ShutdownGracefully(tt.srv, storeFile, connStr)
		})
	}
}

func ExampleInitializeRouter() {

	r := InitializeRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()
	floatValue := 2.0

	metricsObj := metrics.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &floatValue,
	}
	body, _ := json.Marshal(metricsObj)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/update/", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error happened in creating request. Err: %s", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error happened in posting request. Err: %s", err)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error happened in reading response body. Err: %s", err)
		return
	}
	defer resp.Body.Close()
	log.Print(respBody)

}