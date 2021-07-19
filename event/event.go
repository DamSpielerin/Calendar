package event

import "time"

type Event struct {
	Id          int
	Title       string
	Description string
	Time        time.Time
	Timezone    time.Location
	Duration    time.Duration
	Notes       *[]string
}
