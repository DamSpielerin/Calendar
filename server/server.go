package server

import (
	"calendar/event"
	"calendar/storage"
	"calendar/user"
	"context"
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
	Store     storage.EventStore
	UserStore *storage.UserStorage
	http.Handler
}

func NewEventServer(store storage.EventStore) *EventServer {
	es := new(EventServer)

	es.Store = store
	es.UserStore = &storage.Users

	privateRouter := http.NewServeMux()
	privateRouter.HandleFunc("/api/event/", es.ServeEvent)
	privateRouter.HandleFunc("/api/events", es.ServeEvents)
	privateRouter.HandleFunc("/api/user", es.ServeUser)

	privatHandler := AuthMiddleware(privateRouter)
	router := http.NewServeMux()
	router.Handle("/api/", privatHandler)
	router.HandleFunc("/login", es.Login)
	router.HandleFunc("/logout", es.Logout)

	es.Handler = router
	es.Handler = PanicMiddleware(es.Handler)

	return es
}

func (es *EventServer) ServeEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var filter event.EventFilter
		var decoder = schema.NewDecoder()
		fmt.Printf("%#v", r.URL.Query())
		err := decoder.Decode(&filter, r.URL.Query())
		if err != nil {
			http.Error(w, "Error in GET parameters", http.StatusBadRequest)
		} else {
			log.Println("GET parameters : ", filter)
		}
		evs := es.Store.GetEvents(r.Context(), filter)
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
	eventId, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/event/"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut:
		es.SaveEvent(w, *r, eventId)
	case http.MethodGet:
		es.GetEvent(r.Context(), w, eventId)
	case http.MethodDelete:
		es.DeleteEvent(w, eventId)
	}
}

func (es *EventServer) ServeUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Wrong method type", http.StatusBadRequest)
		return
	}
	var userEntity user.User
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Wrong body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &userEntity)
	if err != nil {
		http.Error(w, "Wrong entity", http.StatusBadRequest)
		return
	}
	err = es.UserStore.UpdateTimezone(userEntity.Login, userEntity.Timezone)
	if err != nil {
		http.Error(w, "Wrong Timezone", http.StatusBadRequest)
		return
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
	if es.Store.IsExist(ev.ID) {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func (es *EventServer) GetEvent(ctx context.Context, w http.ResponseWriter, id int) {
	ev, err := es.Store.GetEventById(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if ev.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
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
	var creds user.Credentials
	defer r.Body.Close()
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
		Timezone: userEntity.Timezone,
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
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
	})
}

func (es *EventServer) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now(),
	})
}
