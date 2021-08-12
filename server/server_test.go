package server

import (
	"bytes"
	"calendar/event"
	"encoding/json"
	"fmt"
	"github.com/magiconair/properties/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
)

type StubEventStorage struct {
	store map[int]event.Event
	lock  sync.RWMutex
}

func (s *StubEventStorage) GetEvents(ef event.EventFilter) []event.Event {
	evs := make([]event.Event, len(s.store))
	idx := 0
	for _, ev := range s.store {
		evs[idx] = ev
		idx++
	}
	return evs
}

func (s *StubEventStorage) GetEventById(id int) event.Event {
	return getStubEventById(id)
}

func (s *StubEventStorage) IsExist(id int) bool {
	_, exist := s.store[id]
	return exist
}

func (s *StubEventStorage) Delete(id int) {
	delete(s.store, id)
}

func (s *StubEventStorage) Save(ev event.Event) {
	s.store[ev.ID] = ev
}

func TestEvenServeEvent(t *testing.T) {
	storage := NewStubEventStorage()
	server := NewEventServer(storage)
	t.Run("test get event by id", func(t *testing.T) {
		request := newGetEventByIdRequest("3")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		got := DecodeEventFromResponse(t, response.Body)
		assertStatus(t, response.Code, http.StatusOK)
		assert.Equal(t, got, getStubEventById(3))
	})
}
func TestEvenServeEvents(t *testing.T) {
	storage := NewStubEventStorage()
	server := NewEventServer(storage)
	t.Run("test all events", func(t *testing.T) {
		request := newGetEventsRequest()
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		got := DecodeEventsFromResponse(t, response.Body)
		assertStatus(t, response.Code, http.StatusOK)
		assertEqualEvents(t, got, storage.GetEvents(event.EventFilter{}))
	})
}
func DecodeEventFromResponse(t *testing.T, body *bytes.Buffer) (ev event.Event) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&ev)
	if err != nil {
		t.Fatalf("Unable to parse response from server %q into Event, '%v'", body, err)
	}
	return
}
func DecodeEventsFromResponse(t *testing.T, body *bytes.Buffer) (evs []event.Event) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&evs)
	if err != nil {
		t.Fatalf("Unable to parse response from server %q into Events, '%v'", body, err)
	}
	return
}
func newGetEventsRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/api/events/", nil)
	return req
}
func newGetEventByIdRequest(id string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/api/event/"+id, nil)
	return req
}
func NewStubEventStorage() *StubEventStorage {
	return &StubEventStorage{FillNewStubEvents(), sync.RWMutex{}}
}

func FillNewStubEvents() map[int]event.Event {
	s := map[int]event.Event{}
	for i := 1; i <= 5; i++ {
		s[i] = getStubEventById(i)
	}
	return s
}

func getStubEventById(i int) event.Event {
	loc, _ := time.LoadLocation("America/New_York")
	return event.Event{
		ID:          i,
		Title:       fmt.Sprintf("Test title %d", i),
		Description: "Some description",
		DateTime:    time.Date(2021, time.August, i, i, i, i, 0, loc),
		Timezone:    loc,
		Duration:    time.Hour,
		Notes:       []string{"test note", "test note2"},
		Unmarshaler: nil,
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}
func assertEqualEvents(t testing.TB, a, b []event.Event) bool {
	t.Helper()
	return reflect.DeepEqual(a, b)
}
