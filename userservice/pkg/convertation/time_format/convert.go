package time_format

import (
	"time"
)

func format(t time.Time, format TimeFormat) string {
	var layout string

	switch format {
	case TimeFormat12h:
		layout = "03:04 PM"
	case TimeFormat24h:
		layout = "15:04"
	}

	return t.Format(layout)
}

func parse(timeStr string, format TimeFormat) time.Time {
	var layout string

	switch format {
	case TimeFormat12h:
		layout = "03:04 PM"
	case TimeFormat24h:
		layout = "15:04"
	}

	parsed, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}
	}

	now := time.Now()
	return time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		parsed.Hour(),
		parsed.Minute(),
		parsed.Second(),
		parsed.Nanosecond(),
		parsed.Location(),
	)
}
