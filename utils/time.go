package utils

import (
	"time"
)

var StartTime time.Time

func InitTime() {
	StartTime = time.Now()
}

func GetTime() (curtime string) {
	curtime = time.Since(StartTime).String()

	return
}
