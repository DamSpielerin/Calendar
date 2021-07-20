package event

import (
	"time"
)

type Event struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	DateTime    time.Time      `json:"time"`
	Timezone    *time.Location `json:"timezone"`
	Duration    time.Duration  `json:"duration"`
	Notes       *[]string      `json:"notes"`
	Unmarshaler
}

func NewEvent(ID int, title string, description string, dateTime time.Time, timezone *time.Location, duration time.Duration, notes *[]string, unmarshaler Unmarshaler) *Event {
	return &Event{ID: ID, Title: title, Description: description, DateTime: dateTime, Timezone: timezone, Duration: duration, Notes: notes, Unmarshaler: unmarshaler}
}

type Unmarshaler interface {
	UnmarshalJSON([]byte) error
}
