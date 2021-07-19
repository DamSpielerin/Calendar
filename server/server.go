package server

import (
	"calendar/event"
	"calendar/storage"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const jsonContentType = "application/json"

type EventServer struct {
	Store storage.EventStore
	http.Handler
}

func NewEventServer(store storage.EventStore) *EventServer {
	es := new(EventServer)

	es.Store = store

	router := http.NewServeMux()
	router.Handle("/event/", http.HandlerFunc(es.ServeEvent))
	router.Handle("/events/", http.HandlerFunc(es.ServeEvents))

	es.Handler = router

	return es
}
func (es *EventServer) ServeEvents(w http.ResponseWriter, r *http.Request) {
	var filter event.EventFilter
	var decoder = schema.NewDecoder()
	fmt.Printf("%#v", r.URL.Query())
	err := decoder.Decode(&filter, r.URL.Query())
	if err != nil {
		log.Println("Error in GET parameters : ", err)
	} else {
		log.Println("GET parameters : ", filter)
	}
	evs := es.Store.GetEvents(filter)
	w.Header().Set("content-type", jsonContentType)
	err = json.NewEncoder(w).Encode(evs)
	if err != nil {
		w.WriteHeader(http.StatusInsufficientStorage)
	}
}

func (es *EventServer) ServeEvent(w http.ResponseWriter, r *http.Request) {
	eventId, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/event/"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut:
		es.SaveEvent(w, *r, eventId)
	case http.MethodGet:
		es.GetEvent(w, eventId)
	case http.MethodDelete:
		es.DeleteEvent(w, eventId)
	}

}

func (es *EventServer) SaveEvent(w http.ResponseWriter, r http.Request, id int) {
	exists := es.Store.IsExist(id)
	if (exists && r.Method == http.MethodPost) || (!exists && r.Method == http.MethodPut) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong Method Used!"))
		return
	}
	decoder := json.NewDecoder(r.Body)
	var ev event.Event
	err := decoder.Decode(&ev)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong Entity!"))
		return
	}
	es.Store.Save(ev)
	if evGet := es.Store.GetEventById(ev.Id); evGet.Id != 0 {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
	}

}

func (es *EventServer) GetEvent(w http.ResponseWriter, id int) {
	ev := es.Store.GetEventById(id)
	if ev.Id == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("content-type", jsonContentType)
	err := json.NewEncoder(w).Encode(ev)
	if err != nil {
		w.WriteHeader(http.StatusInsufficientStorage)
	}
}

func (es *EventServer) DeleteEvent(w http.ResponseWriter, id int) {
	if !es.Store.IsExist(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	es.Store.Delete(id)
}
