package storage

import (
	"context"
	"errors"
	"log"
	"time"

	"calendar/event"
	"calendar/user"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var repo *repository = nil

type repository struct {
	db *gorm.DB
}

// NewDbStorage initialises an empty store only one time
func NewDbStorage(dsn string, idleConn, maxConn int) (*repository, error) {
	once.Do(func() {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			return
		}

		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		sqlDB.SetMaxIdleConns(idleConn)

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		sqlDB.SetMaxOpenConns(maxConn)

		// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		sqlDB.SetConnMaxLifetime(time.Hour)

		repo = &repository{db}

	})

	return repo, nil
}

// GetEventById return event according user's timezone
func (i *repository) GetEventById(ctx context.Context, id uuid.UUID) (event.Event, error) {
	var ev event.Event
	i.db.First(&ev, id)
	err := ev.ChangeTimezoneFromContext(ctx)
	return ev, err
}

// GetEvents return all events as slice
func (i *repository) GetEvents(ctx context.Context, ef event.EventFilter) ([]event.Event, error) {
	var events []event.Event
	if v := ctx.Value("user_id"); v != nil {
		var userEntity user.User
		var dateFrom, dateTo time.Time
		var timeFrom, timeTo event.HoursMin
		var loc *time.Location
		var err error

		result := i.db.First(&userEntity, v)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return events, errors.New("user not authorized")
		}
		result = i.db.Where("user_id = ? ", v)
		if ef.Title != "" {
			result = result.Where("title LIKE ?", "%"+ef.Title+"%")
		}

		if ef.Timezone != "" {
			loc, err = time.LoadLocation(ef.Timezone)
		} else if v := ctx.Value("timezone"); v != nil {
			loc, err = time.LoadLocation(v.(string))
		}
		if err != nil || loc == nil {
			loc, _ = time.LoadLocation("UTC")
		}

		if ef.DateFrom != "" {
			dateFrom, err = time.ParseInLocation(shortForm, ef.DateFrom, loc)
			if err != nil {
				log.Println("Wrong date from ", ef.DateFrom, err)
				return nil, err
			}
			result = result.Where("time >= ?", dateFrom)
		}

		if ef.DateTo != "" {
			dateTo, err = time.ParseInLocation(shortForm, ef.DateTo, loc)
			if err != nil {
				log.Println("Wrong date to ", ef.DateTo, err)
				return nil, err
			}
			result = result.Where("time < ?", dateTo)
		}
		result = result.Find(&events)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			result.Error = nil
		}

		if ef.TimeFrom != "" {
			timeFrom, err = event.NewHoursMin(ef.TimeFrom)
			if err != nil {
				log.Println("Wrong time from ", ef.TimeFrom, err)
				return nil, err
			}
		}
		if ef.TimeTo != "" {
			timeTo, err = event.NewHoursMin(ef.TimeTo)
			if err != nil {
				log.Println("Wrong time to ", ef.TimeTo, err)
				return nil, err
			}
		}
		for i := len(events) - 1; i >= 0; i-- {
			err = events[i].ChangeTimezoneFromContext(ctx)
			if err != nil {
				log.Println("can't change timezone for event", events[i].Title, err)
				return nil, err
			}
			et := events[i].DateTime
			if (ef.TimeFrom != "" && !(timeFrom.H < et.Hour() || (timeFrom.H == et.Hour() && timeFrom.M <= et.Minute()))) ||
				(ef.TimeTo != "" && !(timeTo.H > et.Hour() || (timeTo.H == et.Hour() && timeTo.H >= et.Minute()))) {
				events = append(events[:i], events[i+1:]...)
			}
		}
		return events, result.Error
	} else {
		return events, errors.New("user not authorized")
	}

}

// Save event to store
func (i *repository) Save(ctx context.Context, ev event.Event) (event.Event, error) {
	var result *gorm.DB
	exist, err := i.IsExist(ctx, ev.ID)
	if err != nil {
		return ev, err
	}
	if ev.UserId.String() == "00000000-0000-0000-0000-000000000000" {
		if v := ctx.Value("user_id"); v != nil {
			var userEntity user.User
			result = i.db.First(&userEntity, v)
			err = result.Error
			if err != nil {
				return ev, err
			}
			ev.UserId = userEntity.ID
		} else {
			return ev, errors.New("user not authorized")
		}

	}
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
	if v := ctx.Value("user_id"); v != nil {
		var ev event.Event
		var userEntity user.User

		result := i.db.First(&userEntity, v)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, errors.New("user not authorized")
		}
		result = i.db.Where("user_id = ?", v).First(&ev, id)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			result.Error = nil
		}
		return result.RowsAffected > 0, result.Error
	} else {
		return false, errors.New("user not authorized")
	}

}

// Count return number of events in storage
func (i *repository) Count() int {
	var events []event.Event
	result := i.db.Find(&events)
	return int(result.RowsAffected)
}

// GetUserByLogin get user entity from storage by login
func (i *repository) GetUserByLogin(login string) (userEntity user.User, isExist bool, err error) {
	result := i.db.Where("login = ?", login).First(&userEntity)
	err = result.Error
	isExist = result.RowsAffected > 0
	return
}

// IsUserExist check if event already in store
func (i *repository) IsUserExist(login string) (bool, error) {
	var userEntity user.User
	result := i.db.Where("login = ?", login).First(&userEntity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		result.Error = nil
	}
	return result.RowsAffected > 0, result.Error
}

// UpdateTimezone timezone of current user
func (i *repository) UpdateTimezone(ctx context.Context, login string, timezone string) (err error) {
	if userEntity, exist, err := i.GetUserByLogin(login); exist == true {
		if ctx.Value("user_id") == userEntity.ID {
			if _, err = time.LoadLocation(timezone); err == nil {
				userEntity.Timezone = timezone
				result := i.db.Save(&userEntity)
				err = result.Error
			}
		} else {
			err = errors.New("user have no access")
		}

	}
	return
}

// Count return number of users in storage
func (i *repository) UsersCount() int {
	return len(i.GetAllUsers())
}

// GetAllUsers return all users in the store
func (i *repository) GetAllUsers() []user.User {
	var users []user.User
	i.db.Find(&users)
	return users
}

// CreateUser Create new user from userEntity
func (i *repository) CreateUser(userEntity user.User) error {
	result := i.db.Create(&userEntity)
	return result.Error
}
