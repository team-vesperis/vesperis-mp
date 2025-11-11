package util

import (
	"fmt"
	"time"
)

// Example: 10 days, 2 hours, 3 minutes and 15 seconds
func FormatTimeUntil(t time.Time) string {
	return FormatDuration(time.Until(t))
}

func FormatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	return fmt.Sprintf("%d days, %d hours, %d minutes and %d seconds", days, hours, minutes, seconds)
}
