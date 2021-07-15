package main

import (
	"calendar/storage"
	"log"
	"net/http"
	"calendar/server"
)

func main() {
	server := &server.EventServer{storage.NewEventStorage()}
	log.Fatal(http.ListenAndServe(":5000", server))
}
