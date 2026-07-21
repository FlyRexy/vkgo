package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// тут писать SearchServer

type ServerUser struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ID        int    `json:"id,,string"`
	Age       int    `json:"age,,string"`
	About     string `json:"about"`
}

func GetUsers() ([]ServerUser, error) {
	file, err := os.ReadFile("./dataset.json")
	if err != nil {
		return []ServerUser{}, err
	}

	var users []ServerUser
	err = json.Unmarshal(file, &users)
	if err != nil {
		return []ServerUser{}, err
	}
	return users, nil
}

func FilterByQuery(users []ServerUser, query string) []ServerUser {
	var filteredUsers []ServerUser
	for _, user := range users {
		if strings.Contains(user.About, query) {
			filteredUsers = append(filteredUsers, user)
		}

		if strings.Contains(user.FirstName+user.LastName, query) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	return filteredUsers
}

func SortBy(users []ServerUser, orderField string, orderBy int) ([]ServerUser, error) {
	fmt.Println("MY ORDAH", orderBy)
	var isError bool
	sort.Slice(users, func(i, j int) bool {
		var result bool
		switch orderField {
		case "Id":
			result = users[i].ID < users[j].ID
		case "Name":
			result = users[i].FirstName+users[i].LastName < users[j].FirstName+users[j].LastName
		case "Age":
			result = users[i].Age < users[j].Age
		default:
			isError = true
		}

		if orderBy == -1 {
			return !result
		}
		return result
	})

	if isError {
		return nil, errors.New("unknown orderField " + orderField)
	}

	return users, nil
}

func PrepareUsers(users []ServerUser) []User {
	preparedUsers := make([]User, 0, len(users))

	for _, user := range users {
		preparedUsers = append(preparedUsers, User{
			Name:  user.FirstName + user.LastName,
			Age:   user.Age,
			About: user.About,
			ID:    user.ID,
		})
	}

	return preparedUsers
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if token := r.Header.Get("AccessToken"); token == "" {
		http.Error(w, "no access", http.StatusUnauthorized)
		return
	}

	users, err := GetUsers()
	if err != nil {
		bytes, _ := json.Marshal(SearchErrorResponse{Error: err.Error()})
		w.Write(bytes)
	}

	params := r.URL.Query()
	if query := params.Get("query"); query != "" {
		users = FilterByQuery(users, query)
	}

	if orderField := params.Get("order_field"); orderField != "" {
		orderBy, _ := strconv.Atoi(params.Get("order_by"))
		users, err = SortBy(slices.Clone(users), orderField, orderBy)
		if err != nil {
			bytes, _ := json.Marshal(SearchErrorResponse{Error: err.Error()})
			w.Write(bytes)
		}
	}

	if offsetStr := params.Get("offset"); offsetStr != "" {
		offset, _ := strconv.Atoi(offsetStr)
		users = users[offset:]
	}

	if limitStr := params.Get("limit"); limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		if len(users) < limit {
			limit = len(users)
		}
		users = users[:limit]
	}

	usersToSend := PrepareUsers(users)

	bytes, _ := json.Marshal(usersToSend)
	fmt.Println("FILTERED", bytes)

	w.Write(bytes)
}
