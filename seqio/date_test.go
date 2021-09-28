package seqio

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-gts/gts/internal/testutils"
)

func TestDate(t *testing.T) {
	now := time.Now()
	in := FromTime(now)
	out := FromTime(in.ToTime())
	testutils.Equals(t, in, out)
}

var isLeapYearTests = []struct {
	in  int
	out bool
}{
	{2000, true},  // case 1
	{2100, false}, // case 2
	{2020, true},  // case 3
	{2021, false}, // case 4
}

func TestIsLeapYear(t *testing.T) {
	for i, tt := range isLeapYearTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := isLeapYear(tt.in)
			if out != tt.out {
				t.Errorf("isLeapYear(%q) = %v, want %v", tt.in, out, tt.out)
			}
		})
	}
}

var checkDateTests = []struct {
	year  int
	month time.Month
	day   int
	pass  bool
}{
	{2020, 13, 29, false},            // case 1
	{2020, time.February, 0, false},  // case 2
	{2029, time.February, 29, false}, // case 3
	{2020, time.February, 29, true},  // case 4
}

func TestCheckDate(t *testing.T) {
	for i, tt := range checkDateTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			err := checkDate(tt.year, tt.month, tt.day)
			if tt.pass && err != nil {
				t.Errorf("checkDate(%d, %s, %d): %v", tt.year, tt.month, tt.day, err)
			}
			if !tt.pass && err == nil {
				t.Errorf("checkDate(%d, %s, %d): expected an error", tt.year, tt.month, tt.day)
			}
		})
	}
}

var asDatePassTests = []struct {
	in  string
	out Date
}{
	{"02-JAN-2006", Date{2006, time.January, 2}}, // case 1
	{"02-Jan-2006", Date{2006, time.January, 2}}, // case 2
	{"02-01-2006", Date{2006, time.January, 2}},  // case 3
}

var asDateFailTests = []string{
	"02",
	"foo-JAN-2006",
	"02-foo-2006",
	"02-JAN-foo",
}

func TestAsDate(t *testing.T) {
	for i, tt := range asDatePassTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out, err := AsDate(tt.in)
			if err != nil {
				t.Errorf("AsDate(%q): %v", tt.in, err)
				return
			}
			if !reflect.DeepEqual(out, tt.out) {
				t.Errorf("AsDate(%q) = %v, want %v", tt.in, out, tt.out)
			}
		})
	}

	for _, in := range asDateFailTests {
		_, err := AsDate(in)
		if err == nil {
			t.Errorf("AsDate(%q): expected an error", in)
		}
	}
}
