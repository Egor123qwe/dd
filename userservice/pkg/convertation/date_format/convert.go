package date_format

import (
	"time"
)

func format(date time.Time, format DateFormat) string {
	var layout string

	switch format {
	case DateFormatSlashDDMMYYYY:
		layout = "02/01/2006"
	case DateFormatSlashYYYYMMDD:
		layout = "2006/01/02"
	case DateFormatDDMMYYYY:
		layout = "02-01-2006"
	case DateFormatMMDDYYYY:
		layout = "01-02-2006"
	case DateFormatYYYYMMDD:
		layout = "2006-01-02"
	}

	return date.Format(layout)
}

func parse(dateStr string, format DateFormat) time.Time {
	var layout string

	switch format {
	case DateFormatSlashDDMMYYYY:
		layout = "02/01/2006"
	case DateFormatSlashYYYYMMDD:
		layout = "2006/01/02"
	case DateFormatDDMMYYYY:
		layout = "02-01-2006"
	case DateFormatMMDDYYYY:
		layout = "01-02-2006"
	case DateFormatYYYYMMDD:
		layout = "2006-01-02"
	}

	t, _ := time.Parse(layout, dateStr)
	return t
}
