package models

import (
	"time"
)

// GetMillis returns the current time in milliseconds
func GetMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
