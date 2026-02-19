package time_format

import (
	"errors"
	"strings"
)

type TimeFormat string

const (
	TimeFormat12h TimeFormat = "12 hours"
	TimeFormat24h TimeFormat = "24 hours"
)

var ValidTimeFormats = []TimeFormat{
	TimeFormat12h,
	TimeFormat24h,
}

func NewTimeFormat(code string) (TimeFormat, error) {
	for _, validType := range ValidTimeFormats {
		if strings.EqualFold(string(TimeFormat(code)), string(validType)) {
			return validType, nil
		}
	}

	return "", errors.New("invalid time format value object")
}

func (p TimeFormat) Value() string {
	return string(p)
}
