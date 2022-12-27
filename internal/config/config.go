// Config package contains constants and shared metrics of the server module
//
//Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/internal/config
package config

import (
	"database/sql"
	"os"
)

// Database and Server context timeout values
const (
	ContextDBTimeout  = 5
	ContextSrvTimeout = 10
)

// Optional hashing Key
var Key string
// Shared SQL database instance
var DB *sql.DB
// Flag for SQL database use
var DBFlag bool

// GetEnv function is used for retrieving variables passed in the command prompt
func GetEnv(key string, fallback *string) *string {
	if value, ok := os.LookupEnv(key); ok {
		return &value
	}
	return fallback
}
