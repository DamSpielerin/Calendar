package event

import (
	"strconv"
	"strings"
	"time"
)

const (
	longForm = "2006-01-02 15:04:05"
	h24      = time.Hour * 24
)

type EventFilter struct {
	Timezone string `schema:"timezone"` // Location string "America/Chicago", "Europe/Riga"
	DateFrom string `schema:"dateFrom"` // format "2006-01-02"
	DateTo   string `schema:"dateTo"`   // format "2006-01-03"
	TimeFrom string `schema:"timeFrom"` // format "05:30"
	TimeTo   string `schema:"timeTo"`   // format "06:30"
	Title    string `schema:"title"`    // filter if title of event contains this string
}

type HoursMin struct {
	H int
	M int
}

func NewHoursMin(str string) (HoursMin, error) {

	temp := strings.Split(str, ":")
	h, err := strconv.Atoi(temp[0])
	if err != nil {
		return HoursMin{0, 0}, err
	}
	m, err := strconv.Atoi(temp[1])
	if err != nil {
		return HoursMin{0, 0}, err
	}
	return HoursMin{h, m}, nil
}

// IsFiltered check if event meet the criteria
func IsFiltered(event *Event, ef EventFilter, loc *time.Location, dateFrom *time.Time, dateTo *time.Time, timeFrom *HoursMin, timeTo *HoursMin) bool {
	et := event.DateTime.In(loc)
	if ef.DateFrom != "" && !(dateFrom.Before(et) || dateFrom.Equal(et)) {
		return false
	}

	if ef.DateTo != "" && !et.Before(dateTo.Add(h24)) {
		return false
	}

	if ef.Title != "" && !strings.Contains(strings.ToLower(event.Title), strings.ToLower(ef.Title)) {
		return false
	}

	if ef.TimeFrom != "" && !(timeFrom.H < et.Hour() || (timeFrom.H == et.Hour() && timeFrom.M >= et.Minute())) {
		return false
	}

	if ef.TimeTo != "" && !(timeTo.H > et.Hour() || (timeTo.H == et.Hour() && timeTo.H >= et.Minute())) {
		return false
	}

	return true
}
