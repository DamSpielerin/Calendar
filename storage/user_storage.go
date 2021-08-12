package storage

import (
	"calendar/user"
	"sync"
	"time"
)

type UserStorage struct {
	store map[string]user.User
	lock  sync.RWMutex
}

var Users = UserStorage{
	store: map[string]user.User{
		"User1": {
			ID:       1,
			Login:    "User1",
			Email:    "user@ukr.net",
			Password: "password1",
			Timezone: "Europe/Athens",
		}, "User2": {
			ID:       2,
			Login:    "User2",
			Email:    "user2@ukr.net",
			Password: "password2",
			Timezone: "Europe/Riga",
		}},
}

// GetUserByLogin get user entity from storage by login
func (us *UserStorage) GetUserByLogin(login string) (userEntity user.User, isExist bool) {
	us.lock.RLock()
	defer us.lock.RUnlock()
	userEntity, isExist = us.store[login]
	return
}

// IsExist check if event already in store
func (us *UserStorage) IsExist(login string) bool {
	us.lock.RLock()
	defer us.lock.RUnlock()
	_, exist := us.store[login]
	return exist
}

// UpdateTimezone
func (us *UserStorage) UpdateTimezone(login string, timezone string) (err error) {
	us.lock.Lock()
	defer us.lock.Unlock()
	if userEntity, exist := us.store[login]; exist == true {
		if _, err = time.LoadLocation(timezone); err == nil {
			userEntity.Timezone = timezone
			us.store[login] = userEntity
		}
	}
	return
}

// Count return number of users in storage
func (us *UserStorage) Count() int {
	us.lock.RLock()
	defer us.lock.RUnlock()
	return len(us.store)
}
