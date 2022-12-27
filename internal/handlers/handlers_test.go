package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
)

func TestUpdateJSONHandler(t *testing.T) {
    // Create a request to pass to our handler. We don't have any query parameters for now, so we'll
    // pass 'nil' as the third parameter.

	floatValue := 2.0
	tests := []struct {
		name string
		metricsObj     metrics.Metrics
		statusCode   int
		expected string
	}{
		{
			name:  "trial run #1",
			metricsObj: metrics.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &floatValue,
				},
			statusCode: http.StatusOK,
			expected: `{"status":"ok"}`,
		},

		{
			name:  "wrong metrics data",
			metricsObj: metrics.Metrics{},
			statusCode: http.StatusNotImplemented,
			expected: `{"status":"invalid type"}`,
		},

		{
			name:  "missing body",
			statusCode: http.StatusInternalServerError,
			expected: `{"status":"missing json body"}`,
		},

		{
			name:  "wrong metrics format",
			statusCode: http.StatusInternalServerError,
			expected: `{"status":"wrong metrics format"}`,
		},

	}
	for _, tt := range tests {
		body, _ := json.Marshal(tt.metricsObj)
		req, err := http.NewRequest("POST", "/update/", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		if tt.name == "missing body" {
			req, err = http.NewRequest("POST", "/update/", nil)
			if err != nil {
				t.Fatal(err)
			}
		}
		if tt.name == "wrong metrics format" {
			body, _ = json.Marshal("wrong data")
			req, err = http.NewRequest("POST", "/update/", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
		} 

		rr := httptest.NewRecorder()
		handlersWithKey := NewWrapperJSONStruct()

		handler := http.HandlerFunc(handlersWithKey.UpdateJSONHandler)

		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != tt.statusCode {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, tt.statusCode)
		}

		// Check the response body is what we expect.
		if rr.Body.String() != tt.expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), tt.expected)
		}


	}

}

func TestUpdateStringHandler(t *testing.T) {
    // Create a request to pass to our handler. We don't have any query parameters for now, so we'll
    // pass 'nil' as the third parameter.
	
	r := mux.NewRouter()
	handlersWithKey := NewWrapperJSONStruct()
	r.HandleFunc("/update/{type}/{name}/{value}", handlersWithKey.UpdateStringHandler)
	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL+"/update/gauge/Alloc/2.01", nil)
    if err != nil {
        t.Fatal(err)
    }


	resp, err := http.DefaultClient.Do(req)
	respBody, err := ioutil.ReadAll(resp.Body)
	

    // Check the status code is what we expect.
    if status := resp.StatusCode; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }

    // Check the response body is what we expect.
    expected := `{"status":"ok"}`
    if string(respBody) != expected {
        t.Errorf("handler returned unexpected body: got %v want %v",
		string(respBody), expected)
    }
}

func TestUpdateBatchJSONHandler(t *testing.T) {
    // Create a request to pass to our handler. We don't have any query parameters for now, so we'll
    // pass 'nil' as the third parameter.
	floatValue := 2.0
	metricsObj := metrics.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &floatValue,
	}
	metricsBatch := []metrics.Metrics{}
	metricsBatch = append(metricsBatch, metricsObj)
	body, _ := json.Marshal(metricsBatch)
    req, err := http.NewRequest("POST", "/updates/", bytes.NewBuffer(body))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
	handlersWithKey := NewWrapperJSONStruct()

    handler := http.HandlerFunc(handlersWithKey.UpdateBatchJSONHandler)

    handler.ServeHTTP(rr, req)

    // Check the status code is what we expect.
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }

    // Check the response body is what we expect.
    expected := `{"status":"ok"}`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v",
            rr.Body.String(), expected)
    }
}

func TestValueJSONHandler(t *testing.T) {
    // Create a request to pass to our handler. We don't have any query parameters for now, so we'll
    // pass 'nil' as the third parameter.
	metricsObj := metrics.Metrics{
		ID:    "Alloc",
		MType: "gauge",
	}
	body, _ := json.Marshal(metricsObj)
    req, err := http.NewRequest("POST", "/value/", bytes.NewBuffer(body))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
	handlersWithKey := NewWrapperJSONStruct()

    handler := http.HandlerFunc(handlersWithKey.ValueJSONHandler)

    handler.ServeHTTP(rr, req)

    // Check the status code is what we expect.
    if status := rr.Code; status != http.StatusNotFound {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusNotFound)
    }

    // Check the response body is what we expect.
    expected := `{"status":"missing parameter"}`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v",
            rr.Body.String(), expected)
    }
}

func TestValueStringHandler(t *testing.T) {
    // Create a request to pass to our handler. We don't have any query parameters for now, so we'll
    // pass 'nil' as the third parameter.
	
	r := mux.NewRouter()
	handlersWithKey := NewWrapperJSONStruct()
	r.HandleFunc("/value/{type}/{name}", handlersWithKey.ValueStringHandler)
	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL+"/value/gauge/Alloc", nil)
    if err != nil {
        t.Fatal(err)
    }


	resp, err := http.DefaultClient.Do(req)
	respBody, err := ioutil.ReadAll(resp.Body)
	

    // Check the status code is what we expect.
    if status := resp.StatusCode; status != http.StatusNotFound {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusNotFound)
    }

    // Check the response body is what we expect.
    expected := `{"status":"missing parameter"}`
    if string(respBody) != expected {
        t.Errorf("handler returned unexpected body: got %v want %v",
		string(respBody), expected)
    }
}