package time_format

import (
	"encoding/json"
	"time"
)

type Time struct {
	t      time.Time
	format TimeFormat
}

func NewTime(timeStr string, format TimeFormat) Time {
	return Time{
		t:      parse(timeStr, format),
		format: format,
	}
}

func NewTimeFromTime(t time.Time, format TimeFormat) Time {
	return Time{
		t:      t,
		format: format,
	}
}

func (t Time) ConvertTo(format TimeFormat) {
	t.format = format
}

func (t Time) String() string {
	return format(t.t, t.format)
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
