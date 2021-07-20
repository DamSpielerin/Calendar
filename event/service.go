package event

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	shortForm = "2006-01-02"
	longForm  = "2006-01-02 15:04:05"
	h24       = time.Hour * 24
)

type EventFilter struct {
	Timezone string  `schema:"timezone"` // Location string "America/Chicago", "Europe/Riga"
	DateFrom *string `schema:"dateFrom"` // format "2006-01-02"
	DateTo   *string `schema:"dateTo"`   // format "2006-01-03"
	TimeFrom *string `schema:"timeFrom"` // format "05:30"
	TimeTo   *string `schema:"timeTo"`   // format "06:30"
	Title    *string `schema:"title"`    // filter if title of event contains this string
}
type EventHelper struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DateTime    string    `json:"time"`
	Timezone    string    `json:"timezone"`
	Duration    string    `json:"duration"`
	Notes       *[]string `json:"notes"`
}

// MarshalJSON convert event to JSON
func (ev *Event) MarshalJSON() ([]byte, error) {
	eh := EventHelper{ev.ID, ev.Title, ev.Description, ev.DateTime.Format(longForm), ev.Timezone.String(), ev.Duration.String(), ev.Notes}
	return json.Marshal(eh)
}

// UnmarshalJSON convert JSON to event
func (ev *Event) UnmarshalJSON(j []byte) error {
	var eh EventHelper
	err := json.Unmarshal(j, &eh)
	if err != nil {
		return err
	}
	ev.ID = eh.ID
	ev.Title = eh.Title
	ev.Description = eh.Description
	ev.Notes = eh.Notes
	ev.Duration, err = time.ParseDuration(eh.Duration)
	if err != nil {
		return err
	}
	loc, err := time.LoadLocation(eh.Timezone)
	if err != nil || loc == nil {
		return err
	}
	ev.Timezone = loc
	ev.DateTime, err = time.ParseInLocation(longForm, eh.DateTime, loc)
	if err != nil {
		return err
	}
	return nil
}

// IsFiltered check if event meet the criteria
func (ef *EventFilter) IsFiltered(event *Event) bool {

	isOk := true
	loc, err := time.LoadLocation(ef.Timezone)
	if err != nil || loc == nil {
		return false
	}
	et := event.DateTime.In(loc)
	if ef.DateFrom != nil {
		t, err := time.ParseInLocation(shortForm, *ef.DateFrom, loc)
		if err != nil {
			fmt.Print(err)
		}
		isOk = isOk && (t.Before(et) || t.Equal(et))
	}

	if ef.DateTo != nil {
		t, err := time.ParseInLocation(shortForm, *ef.DateTo, loc)
		if err != nil {
			fmt.Print(err)
		}
		isOk = isOk && et.Before(t.Add(h24))
	}

	if ef.Title != nil {
		isOk = isOk && strings.Contains(strings.ToLower(event.Title), strings.ToLower(*ef.Title))
	}

	if ef.TimeFrom != nil {
		temp := strings.Split(*ef.TimeFrom, ":")
		h, err := strconv.Atoi(temp[0])
		if err != nil {
			fmt.Print(err)
		}
		m, err := strconv.Atoi(temp[1])
		if err != nil {
			fmt.Print(err)
		}
		isOk = isOk && (h < et.Hour() || (h == et.Hour() && m >= et.Minute()))
	}

	if ef.TimeTo != nil {
		temp := strings.Split(*ef.TimeTo, ":")
		h, err := strconv.Atoi(temp[0])
		if err != nil {
			fmt.Print(err)
		}
		m, err := strconv.Atoi(temp[1])
		if err != nil {
			fmt.Print(err)
		}
		isOk = isOk && (h > et.Hour() || (h == et.Hour() && m >= et.Minute()))
	}

	return isOk
}
