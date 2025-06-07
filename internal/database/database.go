package database

import (
	"database/sql"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

const (
	TIME_LAYOUT = "2006-01-02 15:04:05"
)

type Client struct {
	db *sql.DB
}

func NewClient(pathToDB string) (*Client, error) {
	db, err := sql.Open("libsql", pathToDB)
	if err != nil {
		return nil, err
	}

	return &Client{db: db}, nil
}
