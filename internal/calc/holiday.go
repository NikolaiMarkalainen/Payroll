package calc

import "time"

// Holiday is a Finnish public / pay-relevant holiday.
type Holiday struct {
	Date time.Time
	Name string
}

// HolidaysInYear returns Finnish holidays commonly relevant for TES pay
// (official red days + jouluaatto and juhannusaatto).
func HolidaysInYear(year int, loc *time.Location) []Holiday {
	if loc == nil {
		loc = time.UTC
	}
	easter := easterSunday(year, loc)
	return []Holiday{
		{date(year, 1, 1, loc), "Uudenvuodenpäivä"},
		{date(year, 1, 6, loc), "Loppiainen"},
		{easter.AddDate(0, 0, -2), "Pitkäperjantai"},
		{easter, "Pääsiäispäivä"},
		{easter.AddDate(0, 0, 1), "2. pääsiäispäivä"},
		{date(year, 5, 1, loc), "Vappu"},
		{easter.AddDate(0, 0, 39), "Helatorstai"},
		{midsummerEve(year, loc), "Juhannusaatto"},
		{midsummerEve(year, loc).AddDate(0, 0, 1), "Juhannuspäivä"},
		{allSaintsDay(year, loc), "Pyhäinpäivä"},
		{date(year, 12, 6, loc), "Itsenäisyyspäivä"},
		{date(year, 12, 24, loc), "Jouluaatto"},
		{date(year, 12, 25, loc), "Joulupäivä"},
		{date(year, 12, 26, loc), "Tapaninpäivä"},
	}
}

// HolidaySet maps YYYY-MM-DD -> holiday name for quick lookup.
func HolidaySet(year int, loc *time.Location) map[string]string {
	out := make(map[string]string)
	for _, h := range HolidaysInYear(year, loc) {
		out[Key(h.Date)] = h.Name
	}
	return out
}

// HolidaySetRange covers every year touched by [from, to] (inclusive dates).
func HolidaySetRange(from, to time.Time) map[string]string {
	if to.Before(from) {
		from, to = to, from
	}
	loc := from.Location()
	out := make(map[string]string)
	for y := from.Year(); y <= to.Year(); y++ {
		for k, name := range HolidaySet(y, loc) {
			out[k] = name
		}
	}
	return out
}

func Key(t time.Time) string {
	return t.Format("2006-01-02")
}

func IsHoliday(t time.Time, set map[string]string) bool {
	_, ok := set[Key(t)]
	return ok
}

func date(year int, month time.Month, day int, loc *time.Location) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

// easterSunday uses the Anonymous Gregorian algorithm.
func easterSunday(year int, loc *time.Location) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return date(year, time.Month(month), day, loc)
}

// midsummerEve is the Friday between 19 and 25 June.
func midsummerEve(year int, loc *time.Location) time.Time {
	d := date(year, 6, 19, loc)
	for d.Weekday() != time.Friday {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

// allSaintsDay is the Saturday between 31 Oct and 6 Nov.
func allSaintsDay(year int, loc *time.Location) time.Time {
	d := date(year, 10, 31, loc)
	for d.Weekday() != time.Saturday {
		d = d.AddDate(0, 0, 1)
	}
	return d
}
