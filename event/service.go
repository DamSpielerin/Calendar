package event

import (
	"fmt"
	"strings"
	"time"
)

type EventFilter struct {
	DateFrom *string `schema:"dateFrom"` // format "2006-01-02"
	DateTo   *string `schema:"dateTo"`   // format "2006-01-03"
	TimeFrom *string `schema:"timeFrom"` // format "05:30"
	TimeTo   *string `schema:"timeTo"`   // format "06:30"
	Title    *string `schema:"title"`    // filter if title of event contains this string
}

// IsFiltered check if event meet the criteria
func (ef *EventFilter) IsFiltered(event *Event) bool {
	const shortForm = "2006-01-02"
	const h24 = time.Hour * 24
	isOk := true
	if ef.DateFrom != nil {
		t, err := time.Parse(shortForm, *ef.DateFrom)
		if err != nil {
			fmt.Print(err)
		}
		isOk = isOk && (t.Before(event.Time) || t.Equal(event.Time))
	}
	if ef.DateTo != nil {
		t, err := time.Parse(shortForm, *ef.DateTo)
		if err != nil {
			fmt.Print(err)
		}
		isOk = isOk && t.After(event.Time.Add(h24))
	}
	if ef.Title != nil {
		isOk = isOk && strings.Contains(event.Title, *ef.Title)
	}

	if ef.TimeFrom != nil || ef.TimeTo != nil {
		t := event.Time
		t = t.Truncate(h24)
		if ef.TimeFrom != nil {
			temp := strings.Split(*ef.TimeFrom, ":")
			h, _ := time.ParseDuration(temp[0] + "h")
			m, _ := time.ParseDuration(temp[1] + "m")
			t = t.Add(h + m - time.Second)
			isOk = isOk && t.Before(event.Time)
		}
		if ef.TimeTo != nil {
			temp := strings.Split(*ef.TimeTo, ":")
			h, _ := time.ParseDuration(temp[0] + "h")
			m, _ := time.ParseDuration(temp[1] + "m")
			t = t.Add(h + m + time.Second)
			isOk = isOk && t.After(event.Time)
		}
	}

	return isOk
}
