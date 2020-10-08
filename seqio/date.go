package seqio

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Date represents a date stamp for record entries.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// FromTime creates a Date object from a time.Time object.
func FromTime(t time.Time) Date {
	return Date{t.Year(), t.Month(), t.Day()}
}

// ToTime converts the Date object into a time.Time object.
func (d Date) ToTime() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

var monthMap = map[string]time.Month{
	"JAN": time.January, "Jan": time.January, "01": time.January,
	"FEB": time.February, "Feb": time.February, "02": time.February,
	"MAR": time.March, "Mar": time.March, "03": time.March,
	"APR": time.April, "Apr": time.April, "04": time.April,
	"MAY": time.May, "May": time.May, "05": time.May,
	"JUN": time.June, "Jun": time.June, "06": time.June,
	"JUL": time.July, "Jul": time.July, "07": time.July,
	"AUG": time.August, "Aug": time.August, "08": time.August,
	"SEP": time.September, "Sep": time.September, "09": time.September,
	"OCT": time.October, "Oct": time.October, "10": time.October,
	"NOV": time.November, "Nov": time.November, "11": time.November,
	"DEC": time.December, "Dec": time.December, "12": time.December,
}

var dayMap = map[time.Month]int{
	time.January:   31,
	time.February:  28,
	time.March:     31,
	time.April:     30,
	time.May:       31,
	time.June:      30,
	time.July:      31,
	time.August:    31,
	time.September: 30,
	time.October:   31,
	time.November:  30,
	time.December:  31,
}

func isLeapYear(year int) bool {
	switch {
	case year%400 == 0:
		return true
	case year%100 == 0:
		return false
	default:
		return year%4 == 0
	}
}

func checkDate(year int, month time.Month, day int) error {
	dayMax, ok := dayMap[month]
	if !ok {
		return fmt.Errorf("bad month value: %q", month)
	}
	if month == time.February && isLeapYear(year) {
		dayMax++
	}
	if day < 1 {
		return fmt.Errorf("day cannot be less than 1, got %d", day)
	}
	if day > dayMax {
		return fmt.Errorf("%q has %d days: got %d", month, dayMax, day)
	}
	return nil
}

// AsDate interprets the given string as a Date.
func AsDate(s string) (Date, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		return Date{}, errors.New("expected 3 fields in date")
	}
	sday, smonth, syear := parts[0], parts[1], parts[2]
	day, err := strconv.Atoi(sday)
	if err != nil {
		return Date{}, fmt.Errorf("cannot interpret %q as day value", sday)
	}
	month, ok := monthMap[smonth]
	if !ok {
		return Date{}, fmt.Errorf("cannot interpret %q as month value", smonth)
	}
	year, err := strconv.Atoi(syear)
	if err != nil {
		return Date{}, fmt.Errorf("cannot interpret %q as year value", syear)
	}
	return Date{year, month, day}, checkDate(year, month, day)
}
