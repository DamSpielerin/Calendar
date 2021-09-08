package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Event struct {
	gorm.Model
	ID          uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey;"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	DateTime    time.Time     `json:"time" gorm:"column:time"`
	Timezone    string        `json:"timezone,omitempty"`
	Duration    time.Duration `json:"duration" gorm:"type:string"`
	Notes       string        `json:"notes,omitempty" gorm:"type:string"`
	UserId      uuid.UUID     `json:"-" gorm:"type:uuid	"`
	Unmarshaler `json:"-" gorm:"-"`
}
type Helper struct {
	ID          uuid.UUID `json:"id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DateTime    string    `json:"time"`
	Timezone    string    `json:"timezone"`
	Duration    string    `json:"duration"`
	Notes       string    `json:"notes"`
}

type Unmarshaler interface {
	UnmarshalJSON([]byte) error
}

// MarshalJSON convert event to JSON
func (ev *Event) MarshalJSON() ([]byte, error) {
	eh := Helper{ev.ID, ev.Title, ev.Description, ev.DateTime.Format(longForm), ev.Timezone, ev.Duration.String(), ev.Notes}
	return json.Marshal(eh)
}

// UnmarshalJSON convert JSON to event
func (ev *Event) UnmarshalJSON(j []byte) error {
	var eh Helper
	err := json.Unmarshal(j, &eh)
	if err != nil {
		return err
	}
	if eh.ID != uuid.Nil {
		ev.ID = eh.ID
	} else {
		ev.ID = uuid.New()
	}

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
	ev.Timezone = loc.String()
	ev.DateTime, err = time.ParseInLocation(longForm, eh.DateTime, loc)
	if err != nil {
		return err
	}
	return nil
}

func (ev *Event) ChangeTimezoneFromContext(ctx context.Context) error {
	if v := ctx.Value("timezone"); v != nil {
		var loc, err = time.LoadLocation(v.(string))
		if err != nil || loc == nil {
			return fmt.Errorf("can't change timezone from context")
		}
		ev.ChangeTimezone(loc)
	}
	return nil
}
func (ev *Event) ChangeTimezone(loc *time.Location) {
	ev.Timezone = loc.String()
	ev.DateTime = ev.DateTime.In(loc)
}
