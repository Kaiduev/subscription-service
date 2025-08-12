package date

import (
	"fmt"
	"strings"
	"time"
)

func ParseMonth(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse("2006-01", s); err == nil {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}
	if t, err := time.Parse("01-2006", s); err == nil {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}
	return time.Time{}, fmt.Errorf("invalid month format (want YYYY-MM or MM-YYYY)")
}

func FormatMonth(t time.Time) string { return t.Format("2006-01") }
