package repository

import (
	"errors"
	"fmt"
	"sync"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	ID       int64  `json:"-"`
}

type UserRepo struct {
	data   []User
	mu     *sync.RWMutex
	nextID int64
}

func NewUserRepository() *UserRepo {
	return &UserRepo{
		data:   make([]User, 0),
		mu:     &sync.RWMutex{},
		nextID: 1,
	}
}

func (repo *UserRepo) CheckUserExist(username, password string) (User, error) {
	var checkedUser User
	for _, user := range repo.data {
		if user.Username == username {
			checkedUser = user
			break
		}
	}

	if checkedUser.Username == "" {
		return User{}, errors.New("such user doesn't exist")
	}

	if checkedUser.Password != password {
		return User{}, errors.New("password doesn't match")
	}

	return checkedUser, nil
}

func (repo *UserRepo) CreateUser(username, password string) (User, error) {
	var checkedUser User
	repo.mu.RLock()
	for _, user := range repo.data {
		if user.Username == username {
			checkedUser = user
			break
		}
	}
	repo.mu.RUnlock()

	if checkedUser.Username != "" {
		return User{}, fmt.Errorf("user with username %s already exist", username)
	}

	repo.mu.Lock()
	newUser := User{
		Username: username,
		Password: password,
		ID:       repo.nextID,
	}
	repo.nextID++
	repo.data = append(repo.data, newUser)
	repo.mu.Unlock()

	return newUser, nil
}
