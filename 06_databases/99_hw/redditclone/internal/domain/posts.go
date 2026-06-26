package domain

import "time"

type PostType string

const (
	Link PostType = "link"
	Text PostType = "text"
)

type Post struct {
	ID               uint64
	Author           User
	CreatedAt        time.Time
	Score            int64
	UpvotePercentage int8
	Views            uint64
	Comments         []Comment
	Votes            []Vote
	Title            string
	Text             string
	Type             PostType
	Category         string
	URL              string
}
