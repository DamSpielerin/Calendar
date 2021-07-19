package main

import (
	"calendar/server"
	"calendar/storage"
	"log"
	"net/http"
)

func main() {
	eventServer := server.NewEventServer(storage.NewEventStorage())
	log.Fatal(http.ListenAndServe(":5000", eventServer))
}
