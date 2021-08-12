package main

import (
	"calendar/server"
	"calendar/storage"
	"log"
	"net/http"
	"sync"
)

var once sync.Once

func main() {
	onceBody := func() {
		eventServer := server.NewEventServer(storage.NewEventStorage())
		metricServer := server.NewMetricServer(storage.NewEventStorage())
		go func() {
			log.Fatal(http.ListenAndServe(":5000", eventServer))
		}()
		log.Fatal(http.ListenAndServe(":5050", metricServer))
	}
	once.Do(onceBody)
}
