package storage

import (
	"calendar/event"
	"encoding/json"
	"sync"
)
type EventStore interface {
	GetEventById(id int) event.Event
	Save(jsonEvent string)
}

func NewEventStorage() *InMemoryEventStorage {
	return &InMemoryEventStorage{map[int]event.Event{}, sync.RWMutex{}}
}

type InMemoryEventStorage struct {
	store map[int]event.Event
	// A mutex is used to synchronize read/write access to the map
	lock sync.RWMutex
}

func (i *InMemoryEventStorage) GetEventById(id int) event.Event{
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.store[id]
}

func (i *InMemoryEventStorage) Save (jsonEvent string)  {
	var ev event.Event
	json.Unmarshal([]byte(jsonEvent), &ev)
	i.lock.Lock()
	defer i.lock.Unlock()
	i.store[ev.Id] = ev
}