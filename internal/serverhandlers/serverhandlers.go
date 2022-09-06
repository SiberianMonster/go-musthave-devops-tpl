package serverhandlers

import (
    "net/http"
    "go-musthave-devops-tpl/internal/storage"
    "go-musthave-devops-tpl/internal/utils"
    "strconv"
    "fmt"
    "strings"
    "encoding/json"
)

var sharedMetrics utils.MetricsContainer
var err error

func StatusHandler(rw http.ResponseWriter, r *http.Request) {
    rw.Header().Set("Content-Type", "application/json")
    params := strings.Split(r.URL.Path, "/")
    fv, err := strconv.ParseFloat(params[len(params)-1], 8)
    if err != nil {
        rw.WriteHeader(http.StatusNotFound)
        rw.Write([]byte("wrong value"))
        return
    }
    var struct_params = utils.UpdateMetrics{params[len(params)-2], fv}

    sharedMetrics , err = storage.RepositoryUpdate(sharedMetrics, struct_params)
    if err != nil {
        rw.WriteHeader(http.StatusNotFound)
        rw.Write([]byte("wrong parameter"))
        return
    }
    s, _ := json.Marshal(sharedMetrics)
    fmt.Println(string(s))
    rw.WriteHeader(http.StatusOK)
    rw.Write([]byte(`{"status":"ok"}`))
} 