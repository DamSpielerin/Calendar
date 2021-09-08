package storage

import (
	"calendar/event"
	"calendar/user"
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
		err = repo.db.AutoMigrate(&event.Event{})
		if err != nil {
			return
		}
		err = repo.db.AutoMigrate(&user.User{})
		if err != nil {
			return
		}
	})

	return repo, nil
}

func (i *repository) GetEventById(ctx context.Context, id uuid.UUID) (event.Event, error) {
	var ev event.Event
	i.db.First(&ev, id)
	err := ev.ChangeTimezoneFromContext(ctx)
	return ev, err
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
	if ev.UserId.String() == "00000000-0000-0000-0000-000000000000" {
		if v := ctx.Value("user_id"); v != nil {
			ev.UserId = v.(uuid.UUID)
		}
	}

	var result *gorm.DB
	if exist {
		result = i.db.Save(&ev)
	} else {
		result = i.db.Create(&ev)
	}
	return ev, result.Error
}

// Delete event from store
func (i *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := i.db.Delete(&event.Event{}, id)
	return result.Error
}

// IsExist check if event already in store
func (i *repository) IsExist(ctx context.Context, id uuid.UUID) (bool, error) {
	var ev event.Event
	result := i.db.First(&ev, id)
	return result.RowsAffected > 0, result.Error
}

// Count return number of events in storage
func (i *repository) Count(ctx context.Context) (int, error) {
	return 0, nil
}
