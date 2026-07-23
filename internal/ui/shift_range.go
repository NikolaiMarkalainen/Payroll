package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// shiftSegment is one visual block of a shift on a calendar day.
// Overnight shifts (e.g. 22:00-06:00) render as two segments on consecutive days.
type shiftSegment struct {
	Shift        calendarShift
	Title        string // place/code above time, e.g. "*3AAA"
	Span         string // time range, e.g. "04:55-14:30"
	Continues    bool   // ends at midnight, continues next day
	Continuation bool   // started previous day
}

func clockToMinutes(hhmm string) (int, error) {
	parts := strings.Split(strings.TrimSpace(hhmm), ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time %q", hhmm)
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("invalid time %q", hhmm)
	}
	if h == 24 && m == 0 {
		return 24 * 60, nil
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid time %q", hhmm)
	}
	return h*60 + m, nil
}

func isOvernight(start, end string) bool {
	s, err1 := clockToMinutes(start)
	e, err2 := clockToMinutes(end)
	if err1 != nil || err2 != nil {
		return false
	}
	return e <= s
}

// absoluteRange returns the half-open interval [start, end) for a stored shift.
func (sh calendarShift) absoluteRange() (start, end time.Time, err error) {
	startMin, err1 := clockToMinutes(sh.Start)
	endMin, err2 := clockToMinutes(sh.End)
	if err1 != nil {
		return time.Time{}, time.Time{}, err1
	}
	if err2 != nil {
		return time.Time{}, time.Time{}, err2
	}
	if startMin == endMin {
		return time.Time{}, time.Time{}, fmt.Errorf("zero-length shift")
	}

	day := time.Date(sh.Date.Year(), sh.Date.Month(), sh.Date.Day(), 0, 0, 0, 0, sh.Date.Location())
	start = day.Add(time.Duration(startMin) * time.Minute)
	end = day.Add(time.Duration(endMin) * time.Minute)
	if endMin <= startMin {
		end = end.Add(24 * time.Hour)
	}
	return start, end, nil
}

func rangesOverlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

func (s *shiftsTab) overlapsExisting(sh calendarShift) bool {
	aStart, aEnd, err := sh.absoluteRange()
	if err != nil {
		return true
	}
	for _, other := range s.shifts {
		if other.ID != 0 && other.ID == sh.ID {
			continue
		}
		bStart, bEnd, err := other.absoluteRange()
		if err != nil {
			continue
		}
		if rangesOverlap(aStart, aEnd, bStart, bEnd) {
			return true
		}
	}
	return false
}

// segmentsOn returns display blocks for a calendar day, including overnight tails.
func (s *shiftsTab) segmentsOn(date time.Time) []shiftSegment {
	var out []shiftSegment
	for _, sh := range s.shifts {
		overnight := isOvernight(sh.Start, sh.End)
		title := shiftTitleLine(sh)

		if sameDate(sh.Date, date) {
			if overnight {
				out = append(out, shiftSegment{
					Shift:     sh,
					Title:     title,
					Span:      sh.Start + "-24:00",
					Continues: true,
				})
			} else {
				out = append(out, shiftSegment{
					Shift: sh,
					Title: title,
					Span:  sh.Start + "-" + sh.End,
				})
			}
		}

		if overnight {
			next := sh.Date.AddDate(0, 0, 1)
			if sameDate(next, date) {
				out = append(out, shiftSegment{
					Shift:        sh,
					Title:        title,
					Span:         "00:00-" + sh.End,
					Continuation: true,
				})
			}
		}
	}
	return out
}

// shiftTitleLine is the place/code line shown above the time span.
func shiftTitleLine(sh calendarShift) string {
	switch {
	case sh.Callout && sh.Code != "":
		return "H *" + sh.Code
	case sh.Callout:
		return "H"
	case sh.Code != "":
		return "*" + sh.Code
	default:
		return ""
	}
}
