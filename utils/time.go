package utils

import (
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

func StringToTime(timeStr string) (*time.Time, error) {
	t, err := time.Parse(timeLayout, timeStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func TimeToString(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
