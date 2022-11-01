package main

import (
	"bytes"
	"encoding/json"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/handlers"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
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
