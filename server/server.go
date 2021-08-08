package server

import (
	"calendar/event"
	"calendar/storage"
	"calendar/user"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/schema"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	router.Handle("/user/login", http.HandlerFunc(es.Login))
	//router.Handle("/user/changeTimezone")

	es.Handler = router

	return es
}

func (es *EventServer) ServeEvents(w http.ResponseWriter, r *http.Request) {
	if !CheckToken(w, r) {
		return
	}
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
	if !CheckToken(w, r) {
		return
	}
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
	if !CheckToken(w, &r) {
		return
	}
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

func (es *EventServer) Login(w http.ResponseWriter, r *http.Request) {
	if CheckToken(w, r) {
		return
	}
	var creds user.Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userEntity, ok := storage.Users.GetUserByLogin(creds.Username)

	if !ok || userEntity.Password != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userEntity.Timezone == "" {
		userEntity.Timezone = "UTC"
	}
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &user.Claims{
		Username: creds.Username,
		Timezone: creds.Timezone,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(user.JwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func CheckToken(w http.ResponseWriter, r *http.Request) bool {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		w.WriteHeader(http.StatusBadRequest)
		return false
	}

	tknStr := c.Value
	claims := &user.Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return user.JwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		w.WriteHeader(http.StatusBadRequest)
		return false
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}
