package ui

import (
	"fmt"
	"math"
	"time"
)

// Allowance indicator codes shown in calendar cells.
const (
	codeCallout     = "H"    // Hälytys
	codeSunday      = "S"    // Sunnuntailisä
	codeHoliday     = "P"    // Pyhälisä
	codeEvening     = "I"    // Iltalisä
	codeNight       = "Y"    // Yölisä
	codeSaturday    = "L"    // Lauantailisä
	codeOvertime50  = "50%"  // Ylityö 50 %
	codeOvertime100 = "100%" // Ylityö 100 %
)

// allowanceRules drives how day chips are calculated.
type allowanceRules struct {
	eveningStartMin   int // minutes from midnight
	eveningEndMin     int
	nightStartMin     int
	nightEndMin       int // typically < nightStart (crosses midnight)
	overtime50AfterH  float64
	overtime100AfterH float64
	holidays          map[string]bool // keys: 2006-01-02
}

func defaultAllowanceRules() allowanceRules {
	return allowanceRules{
		eveningStartMin:   18 * 60,
		eveningEndMin:     22 * 60,
		nightStartMin:     22 * 60,
		nightEndMin:       6 * 60,
		overtime50AfterH:  8,  // TES: often after 8h/day
		overtime100AfterH: 10, // TES: often first 2 OT hours at 50%, then 100%
		holidays:          map[string]bool{},
	}
}

// withYearHolidays adds common fixed Finnish public holidays for the year.
func (r allowanceRules) withYearHolidays(year int) allowanceRules {
	if r.holidays == nil {
		r.holidays = map[string]bool{}
	}
	fixed := []time.Time{
		time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),   // uudenvuodenpäivä
		time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC),   // loppiainen
		time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC),   // vappu
		time.Date(year, 12, 6, 0, 0, 0, 0, time.UTC),  // itsenäisyyspäivä
		time.Date(year, 12, 24, 0, 0, 0, 0, time.UTC), // jouluaatto (often paid as holiday in TES contexts)
		time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), // joulupäivä
		time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), // tapaninpäivä
	}
	for _, d := range fixed {
		r.holidays[holidayKey(d)] = true
	}
	return r
}

// dayAllowances is premium-hour totals for one calendar day.
type dayAllowances struct {
	Total       float64
	Callout     float64
	Sunday      float64
	Holiday     float64
	Evening     float64
	Night       float64
	Saturday    float64
	Overtime50  float64
	Overtime100 float64
}

type allowanceChip struct {
	Code  string
	Hours float64
}

func (d dayAllowances) chips() []allowanceChip {
	type item struct {
		code string
		h    float64
	}
	order := []item{
		{codeCallout, d.Callout},
		{codeSunday, d.Sunday},
		{codeHoliday, d.Holiday},
		{codeEvening, d.Evening},
		{codeNight, d.Night},
		{codeSaturday, d.Saturday},
		{codeOvertime50, d.Overtime50},
		{codeOvertime100, d.Overtime100},
	}
	var out []allowanceChip
	for _, it := range order {
		if it.h > 0.0001 {
			out = append(out, allowanceChip{Code: it.code, Hours: roundHours(it.h)})
		}
	}
	return out
}

func roundHours(h float64) float64 {
	return math.Round(h*10) / 10
}

func formatChipHours(h float64) string {
	if math.Abs(h-math.Round(h)) < 0.05 {
		return fmt.Sprintf("%.0f", math.Round(h))
	}
	return fmt.Sprintf("%.1f", h)
}

func holidayKey(t time.Time) string {
	return t.Format("2006-01-02")
}

func (r allowanceRules) isHoliday(t time.Time) bool {
	if r.holidays == nil {
		return false
	}
	return r.holidays[holidayKey(t)]
}

// summarizeDay clips all shifts to the calendar day and splits premium hours.
func summarizeDay(date time.Time, shifts []calendarShift, rules allowanceRules) dayAllowances {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	var out dayAllowances
	weekday := date.Weekday()
	isSun := weekday == time.Sunday
	isSat := weekday == time.Saturday
	isHol := rules.isHoliday(date)

	for _, sh := range shifts {
		absStart, absEnd, err := sh.absoluteRange()
		if err != nil {
			continue
		}
		clipStart := absStart
		clipEnd := absEnd
		if clipStart.Before(dayStart) {
			clipStart = dayStart
		}
		if clipEnd.After(dayEnd) {
			clipEnd = dayEnd
		}
		if !clipStart.Before(clipEnd) {
			continue
		}

		hours := clipEnd.Sub(clipStart).Hours()
		out.Total += hours
		if sh.Callout {
			out.Callout += hours
		}
		if isSun {
			out.Sunday += hours
		}
		if isSat {
			out.Saturday += hours
		}
		if isHol {
			out.Holiday += hours
		}

		eve, night := splitEveningNight(clipStart, clipEnd, dayStart, rules)
		out.Evening += eve
		out.Night += night
	}

	out.Overtime50, out.Overtime100 = splitOvertime(out.Total, rules.overtime50AfterH, rules.overtime100AfterH)

	out.Total = roundHours(out.Total)
	out.Callout = roundHours(out.Callout)
	out.Sunday = roundHours(out.Sunday)
	out.Holiday = roundHours(out.Holiday)
	out.Evening = roundHours(out.Evening)
	out.Night = roundHours(out.Night)
	out.Saturday = roundHours(out.Saturday)
	out.Overtime50 = roundHours(out.Overtime50)
	out.Overtime100 = roundHours(out.Overtime100)
	return out
}

// splitOvertime splits daily hours into 50% and 100% overtime buckets.
// Example (TES-typical): after50=8, after100=10 → hours 8–10 at 50%, hours after 10 at 100%.
func splitOvertime(total, after50, after100 float64) (ot50, ot100 float64) {
	if after50 < 0 {
		after50 = 0
	}
	if after100 < after50 {
		after100 = after50
	}
	if total <= after50 {
		return 0, 0
	}
	if total <= after100 {
		return total - after50, 0
	}
	return after100 - after50, total - after100
}

// splitEveningNight counts evening/night hours inside [start,end) on dayStart's calendar day.
// Night wins on overlap (e.g. at 22:00).
func splitEveningNight(start, end, dayStart time.Time, rules allowanceRules) (evening, night float64) {
	eve0 := dayStart.Add(time.Duration(rules.eveningStartMin) * time.Minute)
	eve1 := dayStart.Add(time.Duration(rules.eveningEndMin) * time.Minute)

	// Night window for this calendar day: [nightStart, 24:00) U [00:00, nightEnd)
	var nightSpans [][2]time.Time
	if rules.nightEndMin <= rules.nightStartMin {
		nightSpans = [][2]time.Time{
			{dayStart, dayStart.Add(time.Duration(rules.nightEndMin) * time.Minute)},
			{dayStart.Add(time.Duration(rules.nightStartMin) * time.Minute), dayStart.Add(24 * time.Hour)},
		}
	} else {
		nightSpans = [][2]time.Time{
			{
				dayStart.Add(time.Duration(rules.nightStartMin) * time.Minute),
				dayStart.Add(time.Duration(rules.nightEndMin) * time.Minute),
			},
		}
	}

	night = overlapHours(start, end, nightSpans...)
	eveSpans := [][2]time.Time{{eve0, eve1}}
	eveningRaw := overlapHours(start, end, eveSpans...)
	// Remove any evening minutes that are also night (night priority).
	evening = eveningRaw - overlapHours(start, end, intersectSpans(eveSpans, nightSpans)...)
	if evening < 0 {
		evening = 0
	}
	return evening, night
}

func overlapHours(start, end time.Time, spans ...[2]time.Time) float64 {
	var total float64
	for _, sp := range spans {
		a, b := sp[0], sp[1]
		if !a.Before(b) {
			continue
		}
		cs, ce := start, end
		if cs.Before(a) {
			cs = a
		}
		if ce.After(b) {
			ce = b
		}
		if cs.Before(ce) {
			total += ce.Sub(cs).Hours()
		}
	}
	return total
}

func intersectSpans(a, b [][2]time.Time) [][2]time.Time {
	var out [][2]time.Time
	for _, x := range a {
		for _, y := range b {
			cs, ce := x[0], x[1]
			if cs.Before(y[0]) {
				cs = y[0]
			}
			if ce.After(y[1]) {
				ce = y[1]
			}
			if cs.Before(ce) {
				out = append(out, [2]time.Time{cs, ce})
			}
		}
	}
	return out
}
