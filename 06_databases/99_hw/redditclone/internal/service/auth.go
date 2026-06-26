package service

import (
	"context"

	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/lib"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/repository"
)

type UserDTO struct {
	Username string
	Password string
}

type AuthService struct {
	usersRepo   repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthService(ur repository.UserRepository, sr repository.SessionRepository) AuthService {
	return AuthService{
		usersRepo:   ur,
		sessionRepo: sr,
	}
}

func (s AuthService) Login(ctx context.Context, userDTO UserDTO) (domain.User, lib.Token, error) {
	user, err := s.usersRepo.Login(ctx, domain.User{
		Username: userDTO.Username,
		Password: userDTO.Password,
	})
	if err != nil {
		return domain.User{}, "", err
	}

	_, err = s.sessionRepo.Create(ctx, user)
	if err != nil {
		return domain.User{}, "", err
	}

	token := lib.CreateSessionToken(user)

	return user, token, nil
}

func (s AuthService) Signup(ctx context.Context, userDTO UserDTO) (domain.User, lib.Token, error) {
	user, err := s.usersRepo.Signup(ctx, domain.User{
		Username: userDTO.Username,
		Password: userDTO.Password,
	})
	if err != nil {
		return domain.User{}, "", err
	}

	_, err = s.sessionRepo.Create(ctx, user)
	if err != nil {
		return domain.User{}, "", err
	}

	token := lib.CreateSessionToken(user)

	return user, token, nil
}
