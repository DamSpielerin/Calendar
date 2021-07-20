package event

import (
	"reflect"
	"testing"
	"time"
)

func TestEventFilter_IsFiltered(t *testing.T) {
	type fields struct {
		Timezone string
		DateFrom *string
		DateTo   *string
		TimeFrom *string
		TimeTo   *string
		Title    *string
	}
	type args struct {
		event *Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ef := &EventFilter{
				Timezone: tt.fields.Timezone,
				DateFrom: tt.fields.DateFrom,
				DateTo:   tt.fields.DateTo,
				TimeFrom: tt.fields.TimeFrom,
				TimeTo:   tt.fields.TimeTo,
				Title:    tt.fields.Title,
			}
			if got := ef.IsFiltered(tt.args.event); got != tt.want {
				t.Errorf("IsFiltered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_MarshalJSON(t *testing.T) {
	type fields struct {
		ID          int
		Title       string
		Description string
		DateTime    time.Time
		Timezone    *time.Location
		Duration    time.Duration
		Notes       *[]string
		Unmarshaler Unmarshaler
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &Event{
				ID:          tt.fields.ID,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				DateTime:    tt.fields.DateTime,
				Timezone:    tt.fields.Timezone,
				Duration:    tt.fields.Duration,
				Notes:       tt.fields.Notes,
				Unmarshaler: tt.fields.Unmarshaler,
			}
			got, err := ev.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_UnmarshalJSON(t *testing.T) {
	type fields struct {
		ID          int
		Title       string
		Description string
		DateTime    time.Time
		Timezone    *time.Location
		Duration    time.Duration
		Notes       *[]string
		Unmarshaler Unmarshaler
	}
	type args struct {
		j []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &Event{
				ID:          tt.fields.ID,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				DateTime:    tt.fields.DateTime,
				Timezone:    tt.fields.Timezone,
				Duration:    tt.fields.Duration,
				Notes:       tt.fields.Notes,
				Unmarshaler: tt.fields.Unmarshaler,
			}
			if err := ev.UnmarshalJSON(tt.args.j); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
