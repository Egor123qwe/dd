package date_format

import (
	"errors"
	"strings"
)

type DateFormat string

const (
	DateFormatSlashDDMMYYYY DateFormat = "dd/mm/yyyy"
	DateFormatSlashYYYYMMDD DateFormat = "yyyy/mm/dd"
	DateFormatDDMMYYYY      DateFormat = "dd-mm-yyyy"
	DateFormatMMDDYYYY      DateFormat = "mm-dd-yyyy"
	DateFormatYYYYMMDD      DateFormat = "yyyy-mm-dd"
)

var ValidDateFormatValueObjects = []DateFormat{
	DateFormatSlashDDMMYYYY,
	DateFormatSlashYYYYMMDD,
	DateFormatDDMMYYYY,
	DateFormatMMDDYYYY,
	DateFormatYYYYMMDD,
}

func NewDateFormat(code string) (DateFormat, error) {
	for _, validType := range ValidDateFormatValueObjects {
		if strings.EqualFold(string(DateFormat(code)), string(validType)) {
			return validType, nil
		}
	}

	return "", errors.New("invalid date format value object")
}

func (p DateFormat) Value() string {
	return string(p)
}
