package event

import (
	"calendar/user"
	"context"
	"encoding/json"
	"time"
)

type Event struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	DateTime    time.Time      `json:"time"`
	Timezone    *time.Location `json:"timezone,omitempty"`
	Duration    time.Duration  `json:"duration"`
	Notes       []string       `json:"notes,omitempty"`
	ownerUser   *user.User
	Unmarshaler
}
type EventHelper struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DateTime    string   `json:"time"`
	Timezone    string   `json:"timezone"`
	Duration    string   `json:"duration"`
	Notes       []string `json:"notes"`
}

type Unmarshaler interface {
	UnmarshalJSON([]byte) error
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

func (ev *Event) ChangeTimezoneFromContext(ctx context.Context) {
	if v := ctx.Value("timezone"); v != nil {
		var loc, err = time.LoadLocation(v.(string))
		if err != nil || loc == nil {
			panic("wrong timezone from context")
		}
		ev.ChangeTimezone(loc)
	}
}
func (ev *Event) ChangeTimezone(loc *time.Location) {
	ev.Timezone = loc
	ev.DateTime = ev.DateTime.In(loc)
}
