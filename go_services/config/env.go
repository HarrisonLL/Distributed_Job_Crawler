package config

import (
	"fmt"
	"os"
	"strconv"
)

// GetConcurrency returns the concurrency level from environment variables
func GetConcurrency() int {
	concurrency := 1
	if val := os.Getenv("CONCURRENCY"); val != "" {
		concurrency, _ = strconv.Atoi(val)
	}
	return concurrency
}

// GetMode returns the mode from environment variables
func GetMode() string {
	mode := "host"
	if val := os.Getenv("MODE"); val != "" {
		mode = val
	}
	return mode
}

// GetLocation returns the location from environment variables
func GetLocation() string {
	location := "USA"
	if val := os.Getenv("LOCATION"); val != "" {
		location = val
	}
	return location
}

// GetCrawlingTimeInterval returns the crawling time interval from environment variables
func GetCrawlingTimeInterval() int {
	timeInterval := 2
	if val := os.Getenv("TIMEINTERVAL"); val != "" {
		timeInterval, _ = strconv.Atoi(val)
	}
	return timeInterval
}

// GetURL returns the URL for a given service from environment variables
func GetURL(service string) (string, string) {
	if val := os.Getenv(service); val != "" {
		return val, ""
	}
	return "", fmt.Sprintf("env variable %s is not set", service)
}
