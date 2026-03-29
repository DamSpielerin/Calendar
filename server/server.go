package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"calendar/event"
	"calendar/storage"
	"calendar/user"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
)

const jsonContentType = "application/json"

type EventServer struct {
	Store storage.EventStore
	http.Handler
}

func NewEventServer(store storage.EventStore) *EventServer {

	es := new(EventServer)

	es.Store = store

	privateRouter := http.NewServeMux()
	privateRouter.HandleFunc("/api/event/", es.ServeEvent)
	privateRouter.HandleFunc("/api/events", es.ServeEvents)
	privateRouter.HandleFunc("/api/user", es.ServeUser)

	privatHandler := AuthMiddleware(privateRouter)
	router := http.NewServeMux()
	router.Handle("/api/", privatHandler)
	router.HandleFunc("/login", es.Login)
	router.HandleFunc("/logout", es.Logout)
	router.HandleFunc("/signup", es.Signup)

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
		es.SaveEvent(w, r, uuid.New())
	}
	switch r.Method {
	case http.MethodPut:
		es.SaveEvent(w, r, eventId)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Wrong body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &userEntity)
	if err != nil {
		http.Error(w, "Wrong entity", http.StatusBadRequest)
		return
	}
	err = es.Store.UpdateTimezone(r.Context(), userEntity.Login, userEntity.Timezone)
	if err != nil {
		http.Error(w, "Wrong Timezone", http.StatusBadRequest)
		return
	}
}

func (es *EventServer) SaveEvent(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Wrong body", http.StatusBadRequest)
		return
	}
	err = ev.UnmarshalJSON(body)
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
	w.Header().Set("content-type", jsonContentType)

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
	userEntity, ok, err := es.Store.GetUserByLogin(creds.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ok = ok && creds.PasswordVerify(userEntity.PasswordHash)
	if !ok || err != nil {
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

func (es *EventServer) Signup(w http.ResponseWriter, r *http.Request) {
	var creds user.Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	passwordHash, err := creds.PasswordHash()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userEntity := user.User{
		ID:           uuid.New(),
		Login:        creds.Username,
		Email:        creds.Email,
		PasswordHash: passwordHash,
		Timezone:     creds.Timezone,
	}
	if userEntity.Timezone == "" {
		userEntity.Timezone = "UTC"
	}

	err = es.Store.CreateUser(userEntity)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
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
