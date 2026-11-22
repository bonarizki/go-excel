package excel

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

// DefaultDateLayouts returns the layouts tried, in order, when importing
// into a time.Time field without WithDateLayouts. They are deliberately
// unambiguous: ISO forms plus the month-first forms excelize itself
// renders for built-in date formats 14 ("01-02-06") and 22
// ("1/2/06 15:04"). Day-first files (common in Indonesia and Europe)
// need WithDateLayouts("02-01-2006", "02/01/2006").
func DefaultDateLayouts() []string {
	return []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02",
		"01-02-06",
		"01-02-06 15:04",
		"1/2/06 15:04",
		"01/02/06 15:04",
	}
}

// parseTime parses s with each layout in turn, falling back to an Excel
// serial date number (e.g. "45984") when no layout matches. Layouts
// without a zone are interpreted in loc (UTC if nil).
func parseTime(s string, layouts []string, loc *time.Location) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	if loc == nil {
		loc = time.UTC
	}
	if len(layouts) == 0 {
		layouts = DefaultDateLayouts()
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}

	if serial, err := strconv.ParseFloat(s, 64); err == nil && serial > 0 {
		t, err := excelize.ExcelDateToTime(serial, false)
		if err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc), nil
		}
	}

	return time.Time{}, fmt.Errorf("%w: could not parse time: %s", ErrTypeMismatch, s)
}
