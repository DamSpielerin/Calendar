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
		log.Fatal(http.ListenAndServe(":5000", eventServer))
	}
	once.Do(onceBody)
}
