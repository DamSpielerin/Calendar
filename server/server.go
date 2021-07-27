package server

import (
	"calendar/event"
	"calendar/storage"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"io/ioutil"
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
	if r.Method == http.MethodGet {
		var filter event.EventFilter
		var decoder = schema.NewDecoder()
		fmt.Printf("%#v", r.URL.Query())
		err := decoder.Decode(&filter, r.URL.Query())
		if err != nil {
			log.Fatal("Error in GET parameters : ", err)
		} else {
			log.Println("GET parameters : ", filter)
		}
		evs := es.Store.GetEvents(filter)
		w.Header().Set("content-type", jsonContentType)
		err = json.NewEncoder(w).Encode(evs)
		if err != nil {
			w.WriteHeader(http.StatusInsufficientStorage)
		}
	} else {
		http.Error(w, "Wrong Method Used!", http.StatusBadRequest)
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
		http.Error(w, "Wrong method type", http.StatusBadRequest)
		return
	}
	var ev event.Event
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Wrong body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &ev)
	if err != nil {
		http.Error(w, "Wrong entity", http.StatusBadRequest)
		return
	}
	es.Store.Save(ev)
	fmt.Printf("%+v\n", ev)
	if evGet := es.Store.GetEventById(ev.ID); evGet.ID != 0 {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func (es *EventServer) GetEvent(w http.ResponseWriter, id int) {
	ev := es.Store.GetEventById(id)
	if ev.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Printf(" get event by id :%#v\n", ev.DateTime)
	w.Header().Set("content-type", jsonContentType)
	jsonEvent, err := json.Marshal(&ev)
	if err != nil {
		w.WriteHeader(http.StatusInsufficientStorage)
	}
	_, err = w.Write(jsonEvent)
	if err != nil {
		http.Error(w, "Wrong json response", http.StatusInternalServerError)
	}
}

func (es *EventServer) DeleteEvent(w http.ResponseWriter, id int) {
	if !es.Store.IsExist(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	es.Store.Delete(id)
}
