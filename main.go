package main

import (
	"calendar/server"
	"calendar/storage"
	"log"
	"net/http"
)

func main() {

	store, err := storage.NewDbStorage("root:123456@tcp(127.0.0.1:3306)/calendar?charset=utf8mb4", 60, 10)
	if err != nil {
		log.Fatal("can't connect to db: ", err)
	}
	eventServer := server.NewEventServer(store)
	metricServer := server.NewMetricsServer(store)
	go func() {
		log.Fatal(http.ListenAndServe(":5000", eventServer))
	}()
	log.Fatal(http.ListenAndServe(":5050", metricServer))
}
