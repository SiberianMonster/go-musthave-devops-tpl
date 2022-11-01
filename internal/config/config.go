package config

import (
	"database/sql"
	"os"
)

const (
	ContextDBTimeout  = 5
	ContextSrvTimeout = 10
)

var Key string
var DB *sql.DB
var DBFlag bool

func GetEnv(key string, fallback *string) *string {
	if value, ok := os.LookupEnv(key); ok {
		return &value
	}
	return fallback
}
