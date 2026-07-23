package pdfimport

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type dayOffer struct {
	shifts   []Shift
	fullness int
	count    int
	callouts int
	source   int
	label    string
}

// DayConflict is a same-calendar-day clash where fullness is tied (or nearly)
// and the shift sets differ — UI should ask which option to keep.
type DayConflict struct {
	Day     time.Time
	DayKey  string
	Options []ConflictOption
}

// ConflictOption is one PDF's version of a day.
type ConflictOption struct {
	Label       string
	SourceIndex int
	Shifts      []Shift
	FullnessMin int
}

// MergeOutcome is the auto-merged roster plus days that need a user choice.
type MergeOutcome struct {
	Result    *Result
	Conflicts []DayConflict
}

// MergeResults combines several Velho parses into one roster.
// Clear winner (more worked minutes) wins automatically.
// Equal fullness + different shifts → DayConflict for the UI.
func MergeResults(results ...*Result) *Result {
	out, _ := MergeResultsWithConflicts(nil, results...)
	return out
}

// MergeResultsWithConflicts is like MergeResults but reports ambiguous days.
// sourceLabels[i] is a display name for results[i] (e.g. file base name).
func MergeResultsWithConflicts(sourceLabels []string, results ...*Result) (*Result, []DayConflict) {
	out := &Result{}
	if len(results) == 0 {
		return out, nil
	}

	best := map[string]dayOffer{}
	var conflicts []DayConflict
	pendingConflict := map[string][]dayOffer{}

	labelFor := func(i int) string {
		if i >= 0 && i < len(sourceLabels) && sourceLabels[i] != "" {
			return sourceLabels[i]
		}
		return fmt.Sprintf("PDF %d", i+1)
	}

	for i, res := range results {
		if res == nil {
			continue
		}
		if out.Person == "" && res.Person != "" {
			out.Person = res.Person
		}
		if out.Period == "" && res.Period != "" {
			out.Period = res.Period
		}
		if !res.From.IsZero() && (out.From.IsZero() || res.From.Before(out.From)) {
			out.From = res.From
		}
		if !res.To.IsZero() && (out.To.IsZero() || res.To.After(out.To)) {
			out.To = res.To
		}
		for _, w := range res.Warnings {
			out.Warnings = append(out.Warnings, w)
		}

		byDay := groupShiftsByDay(res.Shifts)
		for key, shifts := range byDay {
			shifts = normalizeShifts(shifts)
			f, n, c := dayFullness(shifts)
			offer := dayOffer{
				shifts: shifts, fullness: f, count: n, callouts: c,
				source: i, label: labelFor(i),
			}
			if list, ok := pendingConflict[key]; ok {
				pendingConflict[key] = append(list, offer)
				continue
			}
			prev, ok := best[key]
			if !ok {
				best[key] = offer
				continue
			}
			if shiftsEquivalent(prev.shifts, offer.shifts) {
				// Identical day — keep existing, no warning spam.
				continue
			}
			if offer.fullness > prev.fullness {
				out.Warnings = append(out.Warnings, fmt.Sprintf(
					"%s: käytetty täydempi jakso %s (%d min), ei %s (%d min)",
					key, offer.label, offer.fullness, prev.label, prev.fullness,
				))
				best[key] = offer
				continue
			}
			if offer.fullness < prev.fullness {
				out.Warnings = append(out.Warnings, fmt.Sprintf(
					"%s: pidetty %s (%d min), ohitettu %s (%d min)",
					key, prev.label, prev.fullness, offer.label, offer.fullness,
				))
				continue
			}
			// Equal fullness, different shifts → user chooses.
			delete(best, key)
			pendingConflict[key] = []dayOffer{prev, offer}
		}
	}

	for key, offers := range pendingConflict {
		// Dedupe equivalent offers inside the conflict list.
		uniq := make([]dayOffer, 0, len(offers))
		for _, o := range offers {
			dup := false
			for _, u := range uniq {
				if shiftsEquivalent(u.shifts, o.shifts) {
					dup = true
					break
				}
			}
			if !dup {
				uniq = append(uniq, o)
			}
		}
		if len(uniq) == 1 {
			best[key] = uniq[0]
			continue
		}
		// If one became clearly fuller after normalize, auto-resolve.
		sort.Slice(uniq, func(i, j int) bool {
			return dayOfferBetter(uniq[i], uniq[j])
		})
		if uniq[0].fullness > uniq[1].fullness {
			best[key] = uniq[0]
			out.Warnings = append(out.Warnings, fmt.Sprintf(
				"%s: käytetty täydempi jakso %s (%d min)",
				key, uniq[0].label, uniq[0].fullness,
			))
			continue
		}
		day, _ := time.ParseInLocation("2006-01-02", key, time.Local)
		opts := make([]ConflictOption, 0, len(uniq))
		for _, o := range uniq {
			opts = append(opts, ConflictOption{
				Label:       fmt.Sprintf("%s — %s (%d min)", o.label, summarizeShifts(o.shifts), o.fullness),
				SourceIndex: o.source,
				Shifts:      o.shifts,
				FullnessMin: o.fullness,
			})
		}
		conflicts = append(conflicts, DayConflict{Day: day, DayKey: key, Options: opts})
	}

	keys := make([]string, 0, len(best))
	for k := range best {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out.Shifts = append(out.Shifts, best[k].shifts...)
	}
	out.Shifts = normalizeShifts(out.Shifts)

	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].DayKey < conflicts[j].DayKey
	})
	return out, conflicts
}

// ApplyConflictChoices merges chosen day options into base.
func ApplyConflictChoices(base *Result, choices map[string][]Shift) *Result {
	if base == nil {
		base = &Result{}
	}
	out := &Result{
		From:     base.From,
		To:       base.To,
		Period:   base.Period,
		Person:   base.Person,
		Warnings: append([]string(nil), base.Warnings...),
		Shifts:   append([]Shift(nil), base.Shifts...),
	}
	for _, shifts := range choices {
		out.Shifts = append(out.Shifts, shifts...)
	}
	out.Shifts = normalizeShifts(out.Shifts)
	return out
}

func summarizeShifts(shifts []Shift) string {
	if len(shifts) == 0 {
		return "ei vuoroja"
	}
	parts := make([]string, 0, len(shifts))
	for _, sh := range shifts {
		s := sh.Start + "-" + sh.End
		if sh.Callout {
			s += " H"
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, ", ")
}

func groupShiftsByDay(shifts []Shift) map[string][]Shift {
	out := make(map[string][]Shift)
	for _, sh := range shifts {
		key := dayKey(sh.Date)
		out[key] = append(out[key], sh)
	}
	return out
}

func dayFullness(shifts []Shift) (minutes, count, callouts int) {
	for _, sh := range shifts {
		m := shiftMinutes(sh.Start, sh.End)
		if m <= 0 {
			continue
		}
		minutes += m
		count++
		if sh.Callout {
			callouts++
		}
	}
	return minutes, count, callouts
}

func dayOfferBetter(a, b dayOffer) bool {
	if a.fullness != b.fullness {
		return a.fullness > b.fullness
	}
	if a.count != b.count {
		return a.count > b.count
	}
	if a.callouts != b.callouts {
		return a.callouts > b.callouts
	}
	return a.source > b.source
}

func shiftsEquivalent(a, b []Shift) bool {
	if len(a) != len(b) {
		return false
	}
	type k struct{ start, end string; callout bool }
	count := map[k]int{}
	for _, sh := range a {
		count[k{sh.Start, sh.End, sh.Callout}]++
	}
	for _, sh := range b {
		key := k{sh.Start, sh.End, sh.Callout}
		if count[key] == 0 {
			return false
		}
		count[key]--
	}
	return true
}

func shiftMinutes(start, end string) int {
	sm, err1 := clockMinutes(start)
	em, err2 := clockMinutes(end)
	if err1 != nil || err2 != nil {
		return 0
	}
	if em <= sm {
		em += 24 * 60
	}
	return em - sm
}
