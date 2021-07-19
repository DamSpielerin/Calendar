package server

import (
	"calendar/storage"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type EventServer struct {
	Store storage.EventStore
}

func (es *EventServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventId, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/event/"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	var jsonEvent = ""
	switch r.Method {
	case http.MethodPost:
		es.SaveEvent(w, jsonEvent)
	case http.MethodGet:
		es.GetEvent(w, eventId)
	}

}

func (es *EventServer) SaveEvent(w http.ResponseWriter, jsonEvent string) {
	es.Store.Save(jsonEvent)
	w.WriteHeader(http.StatusOK)
}

func (es *EventServer) GetEvent(w http.ResponseWriter, id int) {
	fmt.Println(es.Store.GetEventById(id))
	w.WriteHeader(http.StatusOK)

}
