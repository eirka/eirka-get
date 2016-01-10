package utils

import (
	"time"
)

var StartTime = time.Now()

func GetTime() string {
	return time.Since(StartTime).String()
}
