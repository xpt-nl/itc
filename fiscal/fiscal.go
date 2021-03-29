// Package fiscal contains functions to calculate the start and end dates of
// fiscal years, quarters and periods (a period is roughly a month) as used by
// Apple. The functions have been verfied by checking against quarterly
// reports since the start of the 2006 fiscal year on the 25th of September
// 2005.
package fiscal

import (
	"time"
)

const Day = 24 * time.Hour

// YearForDate returns the fiscal year (as used by Apple) for a given date.
func YearForDate(date time.Time) int {
	_, end := Year(date.Year())
	if end.Before(date) {
		return date.Year() + 1
	} else {
		return date.Year()
	}
}

// QuarterForDate returns the fiscal year and quarter (as used by Apple) for a
// given date.
func QuarterForDate(date time.Time) (year, quarter int) {
	year = date.Year()
	start, end := Year(year)
	if end.Before(date) {
		year++
		start = end.Add(time.Nanosecond)
	}
	daysSinceStart := int((date.Sub(start).Round(time.Hour).Hours()/24))
	for quarter = 1 + daysSinceStart/98; true; quarter++ {
		start, end = Quarter(year, quarter)
		if !date.Before(start) && !date.After(end) {
			return year, quarter
		}
	}
	return
}

// PeriodForDate returns the fiscal year and period (as used by Apple) for a
// given date.
func PeriodForDate(date time.Time) (year, period int) {
	year = date.Year()
	start, end := Year(year)
	if end.Before(date) {
		year++
		start = end.Add(time.Nanosecond)
	}
	daysSinceStart := int((date.Sub(start).Round(time.Hour).Hours()/24))
	for period = 1 + daysSinceStart/35; true; period++ {
		start, end = Period(year, period)
		if !date.Before(start) && !date.After(end) {
			return year, period
		}
	}
	return
}

// Year returns the start and end date of a fiscal year as used by Apple. The
// year must be 2006 or higher, returns start and end date of the year. End is
// the last nanosecond before the start of the next year.
func Year(year int) (start, end time.Time) {
	first := time.Date(2005, time.September, 25, 0, 0, 0, 0, time.UTC)
	for start = first; true; start = end {
		end = start.Add(364 * Day)
		if end.Day() < 25 {
			end = end.Add(7 * Day)
		}
		if end.Year() >= year {
			return start, end.Add(-time.Nanosecond)
		}
	}
	return
}

// Quarter returns the start and end date of a fiscal quarter as used by
// Apple. The year must be 2006 or higher, The quarter must be in the range
// 1..4. Quarter returns start and end date of the quarter. End is the last
// nanosecond before the start of the next quarter.
func Quarter(year, quarter int) (start, end time.Time) {
	start, end = Year(year)
	if quarter < 1 {
		quarter = 1
	}
	if quarter > 4 {
		quarter = 4
	}
	quarter--
	qstart := time.Duration(quarter * 91)
	qend := time.Duration((quarter + 1) * 91)
	if end.Sub(start).Hours()/24 > 364 {
		// 3rd period is 35 days so first quarter is 98 days
		qstart += time.Duration(((quarter + 3) / 4) * 7)
		qend += time.Duration(((quarter + 4) / 4) * 7)
	}
	return start.Add(qstart * Day), start.Add(qend * Day).Add(-time.Nanosecond)
}

// Period returns the start and end date of a fiscal period as used by Apple.
// A period is either 35 or 28 days and so is roughly equivalent to a single
// month. The year must be 2006 or higher. The period must be in the range
// 1..12. Period returns start and end date of the period. End is the last
// nanosecond before the start of the next period.
func Period(year, period int) (start, end time.Time) {
	start, end = Year(year)
	if period < 1 {
		period = 1
	}
	if period > 12 {
		period = 12
	}
	period--
	pstart := time.Duration(period*28 + ((period+2)/3)*7)
	pend := time.Duration((period+1)*28 + ((period+3)/3)*7)
	if end.Sub(start).Hours()/24 > 364 {
		// 3rd period is 35 days
		pstart += time.Duration(((period + 9) / 12) * 7)
		pend += time.Duration(((period + 10) / 12) * 7)
	}
	return start.Add(pstart * Day), start.Add(pend * Day).Add(-time.Nanosecond)
}
