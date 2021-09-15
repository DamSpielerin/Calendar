package storage

import (
	"context"
	"log"
	"sync"
	"time"

	"calendar/event"

	"github.com/google/uuid"
)

const shortForm = "2006-01-02"

// NewEventStorage initialises an empty store
var once sync.Once
var instance *InMemoryEventStorage = nil

// EventStore stores information about events
type EventStore interface {
	GetEvents(ctx context.Context, ef event.EventFilter) ([]event.Event, error)
	GetEventById(ctx context.Context, id uuid.UUID) (event.Event, error)
	IsExist(ctx context.Context, id uuid.UUID) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Save(ctx context.Context, ev event.Event) (event.Event, error)
	Count() int
	UserStore
}

// NewEventStorage initialises an empty store only one time
func NewEventStorage() *InMemoryEventStorage {
	once.Do(func() {
		instance = &InMemoryEventStorage{map[uuid.UUID]event.Event{}, &sync.RWMutex{}}
	})
	return instance
}

// InMemoryEventStorage collects events to map by id
type InMemoryEventStorage struct {
	store map[uuid.UUID]event.Event
	// A mutex is used to synchronize read/write access to the map
	lock *sync.RWMutex
}

// GetEventById return event by id
func (i *InMemoryEventStorage) GetEventById(ctx context.Context, id uuid.UUID) (event.Event, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	ev := i.store[id]
	err := ev.ChangeTimezoneFromContext(ctx)

	return ev, err
}

// GetEvents return all events as slice
func (i *InMemoryEventStorage) GetEvents(ctx context.Context, ef event.EventFilter) ([]event.Event, error) {
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
			log.Println("Wrong date from ", ef.DateFrom, err)
			return nil, err
		}
	}

	if ef.DateTo != "" {
		dateTo, err = time.ParseInLocation(shortForm, ef.DateTo, loc)
		if err != nil {
			log.Println("Wrong date to ", ef.DateTo, err)
			return nil, err
		}
	}

	if ef.TimeFrom != "" {
		timeFrom, err = event.NewHoursMin(ef.TimeFrom)
		if err != nil {
			log.Println("Wrong time from ", ef.TimeFrom, err)
			return nil, err
		}
	}

	if ef.TimeTo != "" {
		timeTo, err = event.NewHoursMin(ef.TimeTo)
		if err != nil {
			log.Println("Wrong time to ", ef.TimeTo, err)
			return nil, err
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
	return events, nil
}

// Save event to store
func (i *InMemoryEventStorage) Save(ctx context.Context, ev event.Event) (event.Event, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.store[ev.ID] = ev
	return i.GetEventById(ctx, ev.ID)
}

// Delete event from store
func (i *InMemoryEventStorage) Delete(ctx context.Context, id uuid.UUID) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	delete(i.store, id)
	return nil
}

// IsExist check if event already in store
func (i *InMemoryEventStorage) IsExist(ctx context.Context, id uuid.UUID) (bool, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	_, exist := i.store[id]
	return exist, nil
}

// Count return number of events in storage
func (i *InMemoryEventStorage) Count(ctx context.Context) (int, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return len(i.store), nil
}
