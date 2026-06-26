package domain

type Event struct {
	Timestamp int64
	Consumer  string
	Method    string
	Host      string
}
