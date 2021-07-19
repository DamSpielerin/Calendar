package event

import "time"

type Event struct {
	Id int
	Title string
	Description string
	Time time.Time
	Duration time.Duration
}