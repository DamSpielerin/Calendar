package storage

import (
	"calendar/event"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const shortForm = "2006-01-02"

// EventStore stores information about events
type EventStore interface {
	GetEvents(ctx context.Context, ef event.EventFilter) []event.Event
	GetEventById(ctx context.Context, id int) event.Event
	IsExist(id int) bool
	Delete(id int)
	Save(ev event.Event)
	Count() int
}

// NewEventStorage initialises an empty store
func NewEventStorage() *InMemoryEventStorage {
	return &InMemoryEventStorage{map[int]event.Event{}, &sync.RWMutex{}}
}

// InMemoryEventStorage collects events to map by id
type InMemoryEventStorage struct {
	store map[int]event.Event
	// A mutex is used to synchronize read/write access to the map
	lock *sync.RWMutex
}

// GetEventById return event by id
func (i *InMemoryEventStorage) GetEventById(ctx context.Context, id int) event.Event {
	i.lock.RLock()
	defer i.lock.RUnlock()
	ev := i.store[id]
	fmt.Println(ev)
	ev.ChangeTimezoneFromContext(ctx)
	fmt.Println(ev)
	return ev
}

// GetEvents return all events as slice
func (i *InMemoryEventStorage) GetEvents(ctx context.Context, ef event.EventFilter) []event.Event {
	var dateFrom, dateTo time.Time
	var timeFrom, timeTo event.HoursMin

	i.lock.RLock()
	defer i.lock.RUnlock()
	events := make([]event.Event, len(i.store))

	var loc *time.Location
	var err error
	if ef.Timezone != "" {
		loc, err = time.LoadLocation(ef.Timezone)
	} else if v := ctx.Value("timezone"); v != nil {
		loc, err = time.LoadLocation(v.(string))
	}

	if err != nil || loc == nil {
		loc, _ = time.LoadLocation("UTC")
	}

	if ef.DateFrom != "" {
		dateFrom, err = time.ParseInLocation(shortForm, ef.DateFrom, loc)
		if err != nil {
			log.Fatal("Wrong date from ", ef.DateFrom, err)
		}
	}

	if ef.DateTo != "" {
		dateTo, err = time.ParseInLocation(shortForm, ef.DateTo, loc)
		if err != nil {
			log.Fatal("Wrong date to ", ef.DateTo, err)
		}
	}

	if ef.TimeFrom != "" {
		timeFrom, err = event.NewHoursMin(ef.TimeFrom)
		if err != nil {
			log.Fatal("Wrong time from ", ef.TimeFrom, err)
		}
	}

	if ef.TimeTo != "" {
		timeTo, err = event.NewHoursMin(ef.TimeTo)
		if err != nil {
			log.Fatal("Wrong time to ", ef.TimeTo, err)
		}
	}

	idx := 0
	for _, ev := range i.store {
		if event.IsFiltered(&ev, ef, loc, &dateFrom, &dateTo, &timeFrom, &timeTo) {
			events[idx] = ev
			idx++
		}
	}
	events = events[:idx]
	return events
}

// Save event to store
func (i *InMemoryEventStorage) Save(ev event.Event) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.store[ev.ID] = ev

}

// Delete event from store
func (i *InMemoryEventStorage) Delete(id int) {
	i.lock.Lock()
	defer i.lock.Unlock()
	delete(i.store, id)
}

// IsExist check if event already in store
func (i *InMemoryEventStorage) IsExist(id int) bool {
	i.lock.RLock()
	defer i.lock.RUnlock()
	_, exist := i.store[id]
	return exist
}

// Count return number of events in storage
func (i *InMemoryEventStorage) Count() int {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return len(i.store)
}
