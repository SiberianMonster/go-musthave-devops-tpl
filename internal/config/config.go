package config

import (
	"os"
)

const (
	ContextDBTimeout  = 5
	ContextSrvTimeout = 10
)

func GetEnv(key string, fallback *string) *string {
	if value, ok := os.LookupEnv(key); ok {
		return &value
	}
	return fallback
}
