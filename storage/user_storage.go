package storage

import (
	"context"
	"sync"
	"time"

	"calendar/user"

	"github.com/google/uuid"
)

type UserStorage struct {
	store map[string]user.User
	lock  sync.RWMutex
}
type UserStore interface {
	GetUserByLogin(login string) (userEntity user.User, isExist bool, err error)
	IsUserExist(login string) (bool, error)
	UpdateTimezone(ctx context.Context, login string, timezone string) (err error)
	UsersCount() int
	GetAllUsers() []user.User
	CreateUser(userEntity user.User) error
}

var Users = UserStorage{
	store: map[string]user.User{
		"User1": {
			ID:           uuid.New(),
			Login:        "User1",
			Email:        "user@ukr.net",
			PasswordHash: "$2a$10$uBSNrcItnFExYytpKS10cekMn5FvyV/ajg9cLLVZOssBdlU.OyEtu",
			Timezone:     "Europe/Athens",
		}, "User2": {
			ID:           uuid.New(),
			Login:        "User2",
			Email:        "user2@ukr.net",
			PasswordHash: "$2a$10$Let584FS8GiToX2FkjIlSOZWfeZsQZYFE7b98uyo1J7W9TptPzS4S",
			Timezone:     "Europe/Riga",
		}},
}

// GetUserByLogin get user entity from storage by login
func (us *UserStorage) GetUserByLogin(login string) (userEntity user.User, isExist bool, err error) {
	us.lock.RLock()
	defer us.lock.RUnlock()
	userEntity, isExist = us.store[login]
	return
}

// IsUserExist check if event already in store
func (us *UserStorage) IsUserExist(login string) (bool, error) {
	us.lock.RLock()
	defer us.lock.RUnlock()
	_, exist := us.store[login]
	return exist, nil
}

// UpdateTimezone
func (us *UserStorage) UpdateTimezone(ctx context.Context, login string, timezone string) (err error) {
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
func (us *UserStorage) UsersCount() int {
	us.lock.RLock()
	defer us.lock.RUnlock()
	return len(us.store)
}

// GetAllUsers return all users in the store
func (us *UserStorage) GetAllUsers() []user.User {
	us.lock.RLock()
	defer us.lock.RUnlock()
	users := make([]user.User, 0, len(us.store))

	for _, tx := range us.store {
		users = append(users, tx)
	}
	return users
}
