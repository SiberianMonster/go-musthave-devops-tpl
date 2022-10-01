package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"encoding/json"
	"bytes"
)


func testRequest(t *testing.T, ts *httptest.Server, path string, metrics Metrics) (*http.Response, string) {

	body, _ := json.Marshal(metrics)
	
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
	r := NewRouter()

	ts := httptest.NewServer(r)
	defer ts.Close()
	floatValue := 2.0

	metrics := Metrics {
		ID: "Alloc",
		MType: "gauge",
		Value: &floatValue, 
	}

	resp, body := testRequest(t, ts, "/update", metrics)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `{"status":"ok"}`, body)

	wrongMetrics := Metrics {
		ID: "Alloc",
		MType: "othertype",
		Value: &floatValue, 
	}

	resp, body = testRequest(t, ts, "/update", wrongMetrics)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Equal(t, `invalid type`, body)

}
