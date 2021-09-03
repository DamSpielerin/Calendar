package storage

import (
	"calendar/event"
	"context"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var repo *repository = nil

type repository struct {
	db *gorm.DB
}

// NewDbStorage initialises an empty store only one time
func NewDbStorage(dsn string, idleConn, maxConn int) (*repository, error) {

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(idleConn)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(maxConn)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	once.Do(func() {
		repo = &repository{db}
	})

	return repo, nil
}

func (i *repository) GetEventById(ctx context.Context, id uuid.UUID) (event.Event, error) {
	var ev event.Event
	//i.db.First(&ev, id)
	//err := ev.ChangeTimezoneFromContext(ctx)
	return ev, nil
}

// GetEvents return all events as slice
func (i *repository) GetEvents(ctx context.Context, ef event.EventFilter) ([]event.Event, error) {
	return make([]event.Event, 1), nil
}

// Save event to store
func (i *repository) Save(ctx context.Context, ev event.Event) (event.Event, error) {
	exist, err := i.IsExist(ctx, ev.ID)
	if err != nil {
		return ev, err
	}
	if exist {
		i.db.Save(&ev)
	} else {
		i.db.Create(&ev)
	}
	return ev, nil
}

// Delete event from store
func (i *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// IsExist check if event already in store
func (i *repository) IsExist(ctx context.Context, id uuid.UUID) (bool, error) {
	return false, nil
}

// Count return number of events in storage
func (i *repository) Count(ctx context.Context) (int, error) {
	return 1, nil
}