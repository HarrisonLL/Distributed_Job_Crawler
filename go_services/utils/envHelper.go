package utils

import (
	"os"
	"strconv"
)

func GetConcurrency() int {
	concurrency := 1
	if val := os.Getenv("CONCURRENCY"); val != "" {
		concurrency, _ = strconv.Atoi(val)
	}
	return concurrency
}

func GetMode() string {
	mode := "host"
	if val := os.Getenv("MODE"); val != "" {
		mode = val
	}
	return mode
}

func GetLocation() string {
	location := "USA"
	if val := os.Getenv("LOCATION"); val != "" {
		location = val
	}
	return location
}

func GetCrawlingTimeInterval() int {
	timeInterval := 2
	if val := os.Getenv("TIMEINTERVAL"); val != "" {
		timeInterval, _ = strconv.Atoi(val)
	}
	return timeInterval
}

func GetURL(service string) (string, string) {
	if val := os.Getenv(service); val != "" {
		return val, ""
	}
	return "", "Value is not set"
}
