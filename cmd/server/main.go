package main

import (
    "net/http"
    "go-musthave-devops-tpl/internal/serverhandlers"
    "log"
)


func main() {

    // маршрутизация запросов обработчику
    http.HandleFunc("/update/", serverhandlers.StatusHandler)

    // запуск сервера с адресом localhost, порт 8080
    //server := &http.Server{
    //    Addr: "127.0.0.1:8080",
    //}
    log.Fatal(http.ListenAndServe(":8080", nil))
} 