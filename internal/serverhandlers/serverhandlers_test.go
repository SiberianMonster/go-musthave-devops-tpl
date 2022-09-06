package serverhandlers_test

import  (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go-musthave-devops-tpl/internal/serverhandlers"
    "io/ioutil"
)

func TestStatusHandler(t *testing.T) {
    type want struct {
        code        int
        response    string
        contentType string
    }
    tests := []struct {
        name string
        request string
        want want
    }{
        {
            name: "simple test #1",
            request: "/update/gauge/RandomValue/0.318058",
            want: want{
                contentType: "application/json",
                code:  200,
                response:   `{"status":"ok"}`,
            },
        },
        {
            name: "unknown variable test #2",
            request: "/update/gauge/NotRandomValue/0.318058",
            want: want{
                contentType: "application/json",
                code:  404,
                response:   `wrong parameter`,
            },
        },
        {
            name: "wrong value format test #3",
            request: "/update/gauge/RandomValue/aaa",
            want: want{
                contentType: "application/json",
                code:  404,
                response:   `wrong value`,
            },
        },
    }


    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            request := httptest.NewRequest(http.MethodPost, tt.request, nil)
            w := httptest.NewRecorder()
            h := http.HandlerFunc(serverhandlers.StatusHandler)
            h.ServeHTTP(w, request)
            result := w.Result()


            assert.Equal(t, tt.want.code, result.StatusCode)
            assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

            statusResult, err := ioutil.ReadAll(result.Body)
            require.NoError(t, err)
            err = result.Body.Close()
            require.NoError(t, err)

            assert.Equal(t, tt.want.response, string(statusResult))
        })
    }
} 
 