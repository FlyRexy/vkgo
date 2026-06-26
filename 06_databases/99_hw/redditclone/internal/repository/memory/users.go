package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
)

type UserEntity struct {
	Username string
	Password string
	ID       int64
}

type UserRepo struct {
	AnyRepo[[]UserEntity]
	nextID int64
}

func NewUserRepository() *UserRepo {
	return &UserRepo{
		AnyRepo: AnyRepo[[]UserEntity]{
			data: make([]UserEntity, 0),
			mu:   &sync.RWMutex{},
		},
		nextID: 1,
	}
}

func (repo *UserRepo) Login(ctx context.Context, user domain.User) (domain.User, error) {
	var checkedUser domain.User
	for _, currentUser := range repo.data {
		if currentUser.Username == user.Username {
			checkedUser = domain.User{
				ID:       currentUser.ID,
				Username: currentUser.Username,
				Password: currentUser.Password,
			}
			break
		}
	}

	if checkedUser.Username == "" {
		return domain.User{}, errors.New("such user doesn't exist")
	}

	if checkedUser.Password != user.Password {
		return domain.User{}, errors.New("password doesn't match")
	}

	return checkedUser, nil
}

func (repo *UserRepo) Signup(ctx context.Context, user domain.User) (domain.User, error) {
	var checkedUser domain.User
	repo.mu.RLock()
	for _, currentUser := range repo.data {
		if user.Username == currentUser.Username {
			checkedUser = domain.User{
				ID:       currentUser.ID,
				Username: currentUser.Username,
				Password: currentUser.Password,
			}
			break
		}
	}
	repo.mu.RUnlock()

	if checkedUser.Username != "" {
		return domain.User{}, fmt.Errorf("user with username %s already exist", checkedUser.Username)
	}

	repo.mu.Lock()
	newUser := UserEntity{
		Username: user.Username,
		Password: user.Password,
		ID:       repo.nextID,
	}
	repo.nextID++
	repo.data = append(repo.data, newUser)
	repo.mu.Unlock()

	return domain.User{
		ID:       newUser.ID,
		Username: newUser.Username,
		Password: newUser.Password,
	}, nil
}
