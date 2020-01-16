package gts

import "time"

// Date represents a date stamp for record entries.
type Date struct {
	Year  int        `json:"year" yaml:"year" msgpack:"year"`
	Month time.Month `json:"month" yaml:"month" msgpack:"month"`
	Day   int        `json:"day" yaml:"day" msgpack:"day"`
}

// FromTime creates a Date from a time.Time object.
func FromTime(t time.Time) Date {
	return Date{t.Year(), t.Month(), t.Day()}
}

// ToTime returns the time.Time object for the Date.
func (d Date) ToTime() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}
