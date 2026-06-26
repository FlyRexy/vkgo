package domain

import "time"

type Comment struct {
	Body      string
	ID        uint64
	CreatedAt time.Time
	Author    User
}
