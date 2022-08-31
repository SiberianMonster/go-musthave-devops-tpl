package main

import (
    "net/http"
)

// MetricsRegistered — обработчик запроса.
func MetricsRegistered(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("<h1>Metrics received</h1>"))
}

func main() {
    // маршрутизация запросов обработчику
    http.HandleFunc("/", MetricsRegistered)
    // запуск сервера с адресом localhost, порт 8080
    server := &http.Server{
        Addr: "127.0.0.1:8080",
    }
    server.ListenAndServe()
} 