package storage

import (
	"calendar/event"
	"sync"
)

// EventStore stores information about events
type EventStore interface {
	GetEvents(ef event.EventFilter) []event.Event
	GetEventById(id int) event.Event
	IsExist(id int) bool
	Delete(id int)
	Save(ev event.Event)
}

// NewEventStorage initialises an empty player store
func NewEventStorage() *InMemoryEventStorage {
	return &InMemoryEventStorage{map[int]event.Event{}, sync.RWMutex{}}
}

// InMemoryEventStorage collects events to map by id
type InMemoryEventStorage struct {
	store map[int]event.Event
	// A mutex is used to synchronize read/write access to the map
	lock sync.RWMutex
}

// GetEventById return event by id
func (i *InMemoryEventStorage) GetEventById(id int) event.Event {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.store[id]
}

// GetEvents return all events as slice
func (i *InMemoryEventStorage) GetEvents(ef event.EventFilter) []event.Event {
	i.lock.RLock()
	defer i.lock.RUnlock()
	events := make([]event.Event, len(i.store), len(i.store))
	idx := 0
	for _, ev := range i.store {
		if ef.IsFiltered(&ev) {
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
	i.store[ev.Id] = ev

}

// Delete event from store
func (i *InMemoryEventStorage) Delete(id int) {
	i.lock.Lock()
	defer i.lock.Unlock()
	delete(i.store, id)
}

// IsExist check if event already in store
func (i *InMemoryEventStorage) IsExist(id int) bool {
	_, exist := i.store[id]
	return exist
}
