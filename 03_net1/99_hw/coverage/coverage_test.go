package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// тут писать код тестов

func TestStartServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer srv.Close()

	requests := []SearchRequest{
		{
			OrderField: "Age",
			OrderBy:    -1,
			Query:      "Wolf",
			Limit: 2,
		},
		{
			OrderField: "Age",
			OrderBy:    -1,
			Query:      "Wolf",
			Limit: -1,
		},
		{
			OrderField: "Age",
			OrderBy:    -1,
			Query:      "Wolf",
			Limit: 26,
		},
	}
	sc := SearchClient{URL: srv.URL}

	for _, request := range requests {
		fmt.Println(sc.FindUsers(request))
	}
}
