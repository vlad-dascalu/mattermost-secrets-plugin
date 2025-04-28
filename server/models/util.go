package models

import (
	"time"
)

// GetMillis returns the current time in milliseconds
var GetMillis = func() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
