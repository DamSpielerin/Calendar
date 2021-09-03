package server

import (
	"calendar/event"
	"calendar/storage"
	"calendar/user"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/schema"

	"io/ioutil"
	"log"
	"net/http"
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
		evs, err := es.Store.GetEvents(r.Context(), filter)
		if err != nil {
			http.Error(w, "something bad happen with db", http.StatusInternalServerError)
		}
		w.Header().Set("content-type", jsonContentType)
		err = json.NewEncoder(w).Encode(evs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Wrong Method Used!", http.StatusBadRequest)
	}

}

func (es *EventServer) ServeEvent(w http.ResponseWriter, r *http.Request) {
	eventId, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/api/event/"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut:
		es.SaveEvent(w, *r, eventId)
	case http.MethodGet:
		es.GetEvent(r.Context(), w, eventId)
	case http.MethodDelete:
		es.DeleteEvent(r.Context(), w, eventId)
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

func (es *EventServer) SaveEvent(w http.ResponseWriter, r http.Request, id uuid.UUID) {
	exists, err := es.Store.IsExist(r.Context(), id)
	if err != nil {
		http.Error(w, "something wrong happen", http.StatusInternalServerError)
		return
	}
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
	ev, err = es.Store.Save(r.Context(), ev)
	if err != nil {
		http.Error(w, "something wrong happen", http.StatusInternalServerError)
		return
	}
	jsonEvent, err := json.Marshal(&ev)
	if err != nil {
		w.WriteHeader(http.StatusInsufficientStorage)
	}
	_, err = w.Write(jsonEvent)
	if err != nil {
		http.Error(w, "Wrong json response", http.StatusInternalServerError)
	}
}

func (es *EventServer) GetEvent(ctx context.Context, w http.ResponseWriter, id uuid.UUID) {
	ev, err := es.Store.GetEventById(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

func (es *EventServer) DeleteEvent(ctx context.Context, w http.ResponseWriter, id uuid.UUID) {
	exist, err := es.Store.IsExist(ctx, id)
	if err != nil || !exist {
		http.Error(w, "Wrong json response", http.StatusNotFound)
		return
	}
	err = es.Store.Delete(ctx, id)
	if err != nil {
		http.Error(w, "something wrong happen", http.StatusInternalServerError)
		return
	}
}

func (es *EventServer) Login(w http.ResponseWriter, r *http.Request) {
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
		ID:       userEntity.ID,
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
		Name:     "token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	})
}
