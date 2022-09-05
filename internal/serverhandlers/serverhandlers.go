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

func StatusHandler(rw http.ResponseWriter, r *http.Request) {
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(http.StatusOK)
    params := strings.Split(r.URL.Path, "/")
    fv, err := strconv.ParseFloat(params[len(params)-1], 8)
    if err != nil {
        fmt.Println(err)
    }
    var struct_params = utils.UpdateMetrics{params[len(params)-2], fv}

    sharedMetrics = storage.RepositoryUpdate(sharedMetrics, struct_params)
    s, _ := json.Marshal(sharedMetrics)
    fmt.Println(string(s))
    rw.Write([]byte(`{"status":"ok"}`))
} 