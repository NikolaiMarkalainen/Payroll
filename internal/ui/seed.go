package ui

import "time"

// demoRosterShifts is sample data from a real roster export (järjestyksenvalvoja),
// Jul-Aug 2026. "Vt" days are omitted (vuosiloma / vapaata).
// Reported total on the sheet: 98:50.
func demoRosterShifts(loc *time.Location) []calendarShift {
	if loc == nil {
		loc = time.Local
	}
	type row struct {
		month time.Month
		day int
		start string
		end string
	}
	rows := []row{
		{7, 20, "04:55", "16:55"},
		{7, 21, "04:55", "14:30"},
		{7, 23, "04:55", "12:20"},
		{7, 24, "04:55", "14:30"},
		{7, 30, "04:55", "16:50"},
		{7, 31, "04:55", "14:30"},
		{8, 1, "04:55", "14:30"},
		{8, 2, "05:45", "15:00"},
		{8, 3, "04:55", "16:50"},
		{8, 4, "08:00", "16:00"},
	}
	out := make([]calendarShift, 0, len(rows))
	for _, r := range rows {
		out = append(out, calendarShift{
			Date: time.Date(2026, r.month, r.day, 0, 0, 0, 0, loc),
			Start: r.start,
			End: r.end,
		})
	}
	return out
}

func demoFocusMonth(loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	return time.Date(2026, time.July, 1, 0, 0, 0, 0, loc)
}
