package models

import (
	"testing"
	"time"
)

func TestGetMillis(t *testing.T) {
	// Get current time in milliseconds using time.Now()
	now := time.Now().UnixNano() / int64(time.Millisecond)

	// Get time using our GetMillis function
	millis := GetMillis()

	// The difference should be very small (less than 100ms)
	diff := millis - now
	if diff < -100 || diff > 100 {
		t.Errorf("GetMillis returned time too far from current time. Difference: %v ms", diff)
	}
}

func TestGetMillisMonotonic(t *testing.T) {
	// Test that GetMillis is monotonic (always increasing)
	first := GetMillis()
	time.Sleep(1 * time.Millisecond)
	second := GetMillis()

	if second <= first {
		t.Errorf("GetMillis is not monotonic: first=%v, second=%v", first, second)
	}
}
