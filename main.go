package main

import (
	"calendar/server"
	"calendar/storage"
	"log"
	"net/http"
)

func main() {
	eventServer := server.NewEventServer(storage.NewEventStorage())
	metricServer := server.NewMetricsServer(storage.NewEventStorage())
	go func() {
		log.Fatal(http.ListenAndServe(":5000", eventServer))
	}()
	log.Fatal(http.ListenAndServe(":5050", metricServer))
}
