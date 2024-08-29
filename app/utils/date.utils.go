package utils

import (
	"time"
)

func FormatDueDateTime(t string) string {
	if len(t) == 0 {
		return ""
	}

	var layout = "2006-01-02T15:04:05Z0700" // ISO 8601 format
	var output = "Mon, 02 Jan 2006 15:04"

	check_t, _ := time.Parse(layout, t)
	checkDate_t, _ := time.Parse("2006-01-02", check_t.Format("2006-01-02"))
	nowDate_t, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))

	if len(t) == 0 {
		return ""
	}

	f := check_t.Format(output)

	if inTimeSpan(nowDate_t.AddDate(0, 0, -1), nowDate_t, checkDate_t) {
		f = "Yesterday at " + check_t.Format("3:04 PM")
	}
	if inTimeSpan(nowDate_t, nowDate_t.AddDate(0, 0, 1), checkDate_t) {
		f = "Today at " + check_t.Format("3:04 PM")
	}
	if inTimeSpan(nowDate_t.AddDate(0, 0, 1), nowDate_t.AddDate(0, 0, 2), checkDate_t) {
		f = "Tomorrow at " + check_t.Format("3:04 PM")
	}

	return f
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}
