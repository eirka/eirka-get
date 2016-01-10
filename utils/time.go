package utils

import (
	"fmt"
	"time"
)

var StartTime = time.Now()

func GetTime() string {
	return fmt.Sprintf("%.2f", time.Since(StartTime).Minutes())
}
