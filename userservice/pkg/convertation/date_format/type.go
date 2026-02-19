package date_format

import (
	"encoding/json"
	"time"
)

type Date struct {
	t      time.Time
	format DateFormat
}

func NewDate(date string, format DateFormat) Date {
	return Date{
		t:      parse(date, format),
		format: format,
	}
}

func NewDateFromTime(t time.Time, format DateFormat) Date {
	return Date{
		t:      t,
		format: format,
	}
}

func (d Date) ConvertTo(format DateFormat) {
	d.format = format
}

func (d Date) String() string {
	return format(d.t, d.format)
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}
