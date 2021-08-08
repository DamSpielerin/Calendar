package storage

import (
	"calendar/user"
	"sync"
)

type UserStorage struct {
	store map[string]user.User
	lock  sync.RWMutex
}

var Users = UserStorage{
	map[string]user.User{"User1": {
		1,
		"User1",
		"user@ukr.net",
		"password1",
		"America/New_York",
	}},
	sync.RWMutex{},
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
