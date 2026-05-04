package repository

type Post struct {
	ID     uint64
	Author User
	Category string `decodeo `
}
