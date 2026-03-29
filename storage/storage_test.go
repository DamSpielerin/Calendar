package storage

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"calendar/event"
)

type StubEventStorage struct {
	store map[int]event.Event
	lock  *sync.RWMutex
}

func TestIsExist(t *testing.T) {
	store := NewStubEventStorage()
	type args struct {
		id int
	}
	tests := []struct {
		name   string
		fields StubEventStorage
		args   args
		want   bool
	}{
		{"Check event by id 1", *store, args{1}, true},
		{"Check event by id 5", *store, args{5}, true},
		{"Check event by id 6", *store, args{6}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &InMemoryEventStorage{
				store: tt.fields.store,
				lock:  tt.fields.lock,
			}
			if got := i.IsExist(tt.args.id); got != tt.want {
				t.Errorf("IsUserExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEventById(t *testing.T) {
	store := NewStubEventStorage()
	type args struct {
		id int
	}
	tests := []struct {
		name   string
		fields StubEventStorage
		args   args
		want   event.Event
	}{
		{"Get event by id 1", *store, args{1}, GetStubEventById(1)},
		{"Get event by id 5", *store, args{5}, GetStubEventById(5)},
		{"Get event by id 6", *store, args{6}, event.Event{}},
	}
	ctx := context.WithValue(context.Background(), "timezone", "Europe/Riga")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &InMemoryEventStorage{
				store: tt.fields.store,
				lock:  tt.fields.lock,
			}
			if got, _ := i.GetEventById(ctx, tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEventById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func NewStubEventStorage() *StubEventStorage {
	return &StubEventStorage{FillNewStubEvents(), &sync.RWMutex{}}
}

func FillNewStubEvents() map[int]event.Event {
	s := map[int]event.Event{}
	for i := 1; i <= 5; i++ {
		s[i] = GetStubEventById(i)
	}
	return s
}

func GetStubEventById(i int) event.Event {
	loc, _ := time.LoadLocation("America/New_York")
	return event.Event{
		ID:          i,
		Title:       "Test title " + string(rune(i)),
		Description: "Some descr",
		DateTime:    time.Date(2021, time.August, i, i, i, i, 0, loc),
		Timezone:    loc,
		Duration:    time.Hour,
		Notes:       []string{"test", "test2"},
		Unmarshaler: nil,
	}
}
