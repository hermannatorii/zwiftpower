package zp

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRiderMonthsAgoString(t *testing.T) {
	cases := []struct {
		d        time.Time
		expected string
	}{
		{d: time.Now(), expected: "This month"},
		{d: time.Now().Add(-1 * 30 * 24 * time.Hour), expected: "Last month"},
		{d: time.Now().Add(-2 * 30 * 24 * time.Hour), expected: "2 months ago"},
		{d: time.Now().Add(-3 * 30 * 24 * time.Hour), expected: "3 months ago"},
		{d: time.Now().Add(-13 * 30 * 24 * time.Hour), expected: "Over a year ago"},
		{expected: "No latest event"},
	}

	for i, c := range cases {
		r := Rider{
			LatestEventDate: c.d,
		}
		result := r.MonthsAgo()
		if result != c.expected {
			t.Fatalf("Case %d: got %s expected %s", i, result, c.expected)
		}
	}
}

func TestRiderStrings(t *testing.T) {
	ss := strings.Split("Liz Rice,98588,2020-04-15,This month,ZZRC SUB 2.0 Ride,160,https://www.zwiftpower.com/profile.php?z=98588,2.5,2.7,0,3,47,Stage 4 Race - Tour of Watopia 2020,2020-03-21", ",")

	r := Rider{
		Name:            "Liz Rice",
		Zwid:            98588,
		LatestEventDate: time.Now(),
	}
	rr := r.Strings()
	if len(rr) != len(ss) {
		t.Fatalf("Strings length %d, expected %d", len(rr), len(ss))
	}

	for i := range rr {
		switch i {
		case 0, 1:
			if rr[i] != ss[i] {
				t.Errorf("Got %s expected %s", rr[i], ss[i])
			}
			// TODO! Test more fields
		default:
			continue

		}

	}

}

func TestFormatAsExpected(t *testing.T) {
	ss := strings.Split("Some name,98588,2020-04-15,This month,ZZRC SUB 2.0 Ride,160,https://www.zwiftpower.com/profile.php?z=98588,2.5,2.7,0,3,47,Stage 4 Race - Tour of Watopia 2020,2020-03-21", ",")

	rider, err := ImportRider(98588)
	if err != nil {
		t.Fatalf("Failed to get data from ZwiftPower: %v", err)
	}

	result := rider.Strings()
	if len(result) != len(ss) {
		t.Errorf("Data length %d, expected %d", len(result), len(ss))
	}

	for i := range ss {
		switch i {
		case 0:
			// Name not included in the data from this URL
			continue
		case 1, 6: // ID, URL
			// Field should be identical
			if ss[i] != result[i] {
				t.Errorf("Unexpected data field %d, got %s expected %s", i, result[i], ss[i])
			}
		case 2, 13:
			// Check for a date format
			v := strings.Split(result[i], "-")
			if len(v) != 3 {
				t.Errorf("Unexpected format for field %d, got %s expected date in format 2006-01-02", i, result[i])
			}
		case 3:
			if !strings.Contains(result[i], "month") {
				t.Errorf("Unexpected data field %d, got %s expected %s", i, result[i], ss[i])
			}
		case 4, 12:
			// Strings that will vary wildly
			continue
		case 5, 9, 10, 11:
			// Check it's an integer number
			_, err := strconv.Atoi(result[i])
			if err != nil {
				t.Errorf("Error converting field %d containing %s to integer: %v", i, result[i], err)
			}
		case 7, 8:
			// Check for a decimal format that's plausible for W/kg
			v, err := strconv.ParseFloat(result[i], 32)
			if err != nil {
				t.Errorf("Error converting field %d containing %s to float: %v", i, result[i], err)
			}

			// Ftp30 will be zero if date was > 30 days ago
			switch result[3] {
			case "This month":
				if v < 0.1 || v > 6 {
					t.Errorf("Unexpected power value in field %d from string %s", i, result[i])
				}
			default:
				if i == 7 && result[i] != "0.0" {
					t.Errorf("Should have 0.0 for 30-day FTP since last event was over a month ago, got %s", result[i])
				}
			}
		}
	}
}
