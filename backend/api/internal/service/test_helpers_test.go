package service

import "time"

func now() time.Time {
	return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
}
