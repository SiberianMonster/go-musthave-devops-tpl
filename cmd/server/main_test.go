package main

import  (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "io/ioutil"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {

    req, err := http.NewRequest(method, ts.URL+path, nil)
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

    resp, body := testRequest(t, ts, "GET", "/update/gauge/RandomValue/0.318058")
	defer resp.Body.Close()
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    assert.Equal(t, `{"status":"ok"}`, body)

    resp, body = testRequest(t, ts, "GET", "/update/gauge/RandomValue/aaa")
	defer resp.Body.Close()
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
    assert.Equal(t, `wrong value`, body)

} 

