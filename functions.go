package main

import (
	"log"
	"os"
)

// Getenv ... Retrives an environment variable but provides a default fallback value if empty
func Getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// FailOnError ... A simple function to handle errors
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
