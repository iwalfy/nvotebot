package util

import (
	"os"
)

func Must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}

func Getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
