package repository

type Repository struct {
	UserRepo    *UserRepo
	SessionRepo *SessionRepo
}

func NewRepository() *Repository {
	return &Repository{
		UserRepo:    NewUserRepository(),
		SessionRepo: NewSessionRepo(),
	}
}
