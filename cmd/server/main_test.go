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


func testRequest(t *testing.T, ts *httptest.Server, method string, path string, metrics Metrics) (*http.Response, string) {

	body, _ := json.Marshal(metrics)
	
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(body))
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

	metrics := Metrics {
		ID: "Alloc",
		MType: "gauge",
		Value: 2.0 ,
	}

	resp, body := testRequest(t, ts, "POST", "/update/", metrics)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `{"status":"ok"}`, body)

	wrong_metrics := Metrics {
		ID: "Alloc",
		MType: "othertype",
		Value: 2.0 ,
	}

	resp, body = testRequest(t, ts, "POST", "/update/", wrong_metrics)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Equal(t, `invalid type`, body)

}
