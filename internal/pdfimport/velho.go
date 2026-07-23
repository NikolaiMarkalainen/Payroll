package pdfimport

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	headerRe = regexp.MustCompile(`(?i)(?:Toteutuneet|Suunnitellut)\s+työvuorot:\s*(\d{1,2})\.(\d{1,2})\.(\d{4})\s*-\s*(\d{1,2})\.(\d{1,2})\.(\d{4})(?:\s*\(([^)]+)\))?`)
	dayRe    = regexp.MustCompile(`(?i)^(ma|ti|ke|to|pe|la|su)\s+(\d{1,2})\.(\d{1,2})\.?$`)
	rangeRe  = regexp.MustCompile(`^(@?)(\d{1,2}:\d{2})-(\d{1,2}:\d{2})\*(\S+)$`)
	personRe = regexp.MustCompile(`(?i)^TP\s+.+`)
	weekRe   = regexp.MustCompile(`(?i)^vk\s+\d+`)
)

// ParseFile extracts and parses a TyövuoroVelho "Toteutuneet työvuorot" PDF.
func ParseFile(path string) (*Result, error) {
	tokens, err := ExtractTokens(path)
	if err != nil {
		return nil, err
	}
	return ParseVelho(tokens)
}

// ParseVelho parses tokenized TyövuoroVelho report text.
func ParseVelho(tokens []string) (*Result, error) {
	res := &Result{}
	for _, t := range tokens {
		if m := headerRe.FindStringSubmatch(t); m != nil {
			from, err1 := dmy(m[1], m[2], m[3])
			to, err2 := dmy(m[4], m[5], m[6])
			if err1 == nil && err2 == nil {
				res.From, res.To = from, to
				res.Period = strings.TrimSpace(m[7])
			}
		}
		if res.Person == "" && personRe.MatchString(t) {
			res.Person = t
		}
	}
	if res.From.IsZero() || res.To.IsZero() {
		return nil, fmt.Errorf("ei löytynyt otsikkoa 'Toteutuneet/Suunnitellut työvuorot: PP.KK.VVVV - PP.KK.VVVV'")
	}

	// Walk main shift table (Suunnitellut/Toteutuneet), skip Hälytys / Vuosivapaa pages.
	mode := "seek"
	var dayShifts []Shift
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		switch {
		case strings.EqualFold(t, "Hälytys aika"):
			mode = "skip"
			continue
		case strings.Contains(strings.ToLower(t), "vuosivapaatilasto"):
			mode = "skip"
			continue
		case strings.EqualFold(t, "Suunnitellut"):
			mode = "shifts"
			continue
		}
		if mode != "shifts" {
			continue
		}
		if strings.EqualFold(t, "Yhteensä") || strings.EqualFold(t, "Tavoite") || strings.EqualFold(t, "Maksuun") {
			mode = "seek"
			continue
		}
		if weekRe.MatchString(t) {
			continue
		}
		dm := dayRe.FindStringSubmatch(t)
		if dm == nil {
			continue
		}
		day, err := dayDate(dm[2], dm[3], res.From, res.To)
		if err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("ohitettu päivä %q: %v", t, err))
			continue
		}
		ranges := collectRanges(tokens, i+1)
		picked := pickRealized(ranges)
		for _, r := range picked {
			if r.start == r.end {
				continue // 00:00-00:00*TH etc.
			}
			dayShifts = append(dayShifts, Shift{
				Date:    day,
				Start:   normalizeClock(r.start),
				End:     normalizeClock(r.end),
				Callout: r.callout,
				Code:    r.code,
			})
		}
	}

	res.Shifts = normalizeShifts(dayShifts)
	applyCalloutDays(res.Shifts, parseCalloutDays(tokens, res.From, res.To))
	if len(res.Shifts) == 0 {
		res.Warnings = append(res.Warnings, "ei löytynyt vuoroja (toteutuneet tai suunnitellut)")
	}
	return res, nil
}

// parseCalloutDays reads the "Hälytys aika" page: days with duration > 0.
// Those days get Callout=true on imported shifts (same as UI checkbox).
func parseCalloutDays(tokens []string, from, to time.Time) map[string]bool {
	out := make(map[string]bool)
	mode := false
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		switch {
		case strings.EqualFold(t, "Hälytys aika"):
			mode = true
			continue
		case strings.EqualFold(t, "Suunnitellut"):
			mode = false
			continue
		case strings.Contains(strings.ToLower(t), "vuosivapaatilasto"):
			mode = false
			continue
		}
		if !mode {
			continue
		}
		if weekRe.MatchString(t) || strings.EqualFold(t, "Yhteensä") ||
			strings.EqualFold(t, "Tavoite") || strings.EqualFold(t, "Maksuun") ||
			strings.EqualFold(t, "Aikahyvitys") || strings.EqualFold(t, "Edellinen") {
			continue
		}
		dm := dayRe.FindStringSubmatch(t)
		if dm == nil {
			continue
		}
		day, err := dayDate(dm[2], dm[3], from, to)
		if err != nil {
			continue
		}
		if calloutDurationPositive(tokens, i+1) {
			out[dayKey(day)] = true
		}
	}
	return out
}

var durationRe = regexp.MustCompile(`^(\d{1,3})[:.](\d{2})$`)

func calloutDurationPositive(tokens []string, from int) bool {
	for j := from; j < len(tokens); j++ {
		t := strings.TrimSpace(tokens[j])
		if t == "" {
			continue
		}
		if dayRe.MatchString(t) || weekRe.MatchString(t) ||
			strings.EqualFold(t, "Yhteensä") || strings.EqualFold(t, "Tavoite") ||
			strings.EqualFold(t, "Maksuun") || strings.EqualFold(t, "Aikahyvitys") ||
			strings.EqualFold(t, "Hälytys aika") || rangeRe.MatchString(t) {
			return false
		}
		if m := durationRe.FindStringSubmatch(t); m != nil {
			h, _ := strconv.Atoi(m[1])
			min, _ := strconv.Atoi(m[2])
			return h > 0 || min > 0
		}
		// Ignore other labels on the row.
	}
	return false
}

func applyCalloutDays(shifts []Shift, days map[string]bool) {
	if len(days) == 0 {
		return
	}
	for i := range shifts {
		if days[dayKey(shifts[i].Date)] {
			shifts[i].Callout = true
		}
		// Overnight stored on start day: also tag if any calendar day it covers has hälytys.
		if isOvernightClock(shifts[i].Start, shifts[i].End) {
			next := shifts[i].Date.AddDate(0, 0, 1)
			if days[dayKey(next)] {
				shifts[i].Callout = true
			}
		}
	}
}

func dayKey(d time.Time) string {
	return d.Format("2006-01-02")
}

func isOvernightClock(start, end string) bool {
	sm, err1 := clockMinutes(start)
	em, err2 := clockMinutes(end)
	if err1 != nil || err2 != nil {
		return false
	}
	return em <= sm
}

func clockMinutes(hhmm string) (int, error) {
	parts := strings.Split(hhmm, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("bad")
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("bad")
	}
	return h*60 + m, nil
}

type rawRange struct {
	callout bool
	start   string
	end     string
	code    string
}

func collectRanges(tokens []string, from int) []rawRange {
	var out []rawRange
	for j := from; j < len(tokens); j++ {
		t := tokens[j]
		if dayRe.MatchString(t) || weekRe.MatchString(t) ||
			strings.EqualFold(t, "Yhteensä") || strings.EqualFold(t, "Tavoite") ||
			strings.EqualFold(t, "Maksuun") || strings.EqualFold(t, "Aikahyvitys") ||
			strings.EqualFold(t, "Hälytys aika") {
			break
		}
		if m := rangeRe.FindStringSubmatch(t); m != nil {
			out = append(out, rawRange{
				callout: m[1] == "@",
				start:   m[2],
				end:     m[3],
				code:    m[4],
			})
		}
	}
	return out
}

// pickRealized: PDF tokens for a day are column-wise:
//   [planned1, planned2, ..., realized1, realized2, ...]
// not interleaved pairs. Even count → take the second half (Toteutuneet).
// Odd count → keep non-zero (only one column present, e.g. Suunnitellut-PDF).
func pickRealized(ranges []rawRange) []rawRange {
	var nonZero []rawRange
	for _, r := range ranges {
		if r.start != r.end {
			nonZero = append(nonZero, r)
		}
	}
	if len(nonZero) == 0 {
		return nil
	}
	if len(nonZero)%2 == 0 {
		half := len(nonZero) / 2
		return append([]rawRange(nil), nonZero[half:]...)
	}
	return nonZero
}

func mergeOvernightHalves(in []Shift) []Shift {
	if len(in) == 0 {
		return in
	}
	// Sort by date then start so halves sit next to each other.
	sorted := append([]Shift(nil), in...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if !sorted[i].Date.Equal(sorted[j].Date) {
			return sorted[i].Date.Before(sorted[j].Date)
		}
		return sorted[i].Start < sorted[j].Start
	})
	out := make([]Shift, 0, len(sorted))
	for i := 0; i < len(sorted); i++ {
		cur := sorted[i]
		if i+1 < len(sorted) {
			next := sorted[i+1]
			contiguous := cur.End == "24:00" && next.Start == "00:00" &&
				next.Date.Equal(cur.Date.AddDate(0, 0, 1))
			if contiguous {
				cur.End = next.End
				if next.Callout {
					cur.Callout = true
				}
				if cur.Code == "" {
					cur.Code = next.Code
				}
				i++
			}
		}
		out = append(out, cur)
	}
	return out
}

// normalizeShifts removes exact duplicates and merges overnight 24:00 / 00:00 halves.
func normalizeShifts(in []Shift) []Shift {
	return mergeOvernightHalves(dedupeExactShifts(in))
}

func dedupeExactShifts(in []Shift) []Shift {
	if len(in) == 0 {
		return in
	}
	type key struct {
		day, start, end string
		callout         bool
	}
	seen := map[key]Shift{}
	order := make([]key, 0, len(in))
	for _, sh := range in {
		k := key{
			day:     dayKey(sh.Date),
			start:   sh.Start,
			end:     sh.End,
			callout: sh.Callout,
		}
		if prev, ok := seen[k]; ok {
			if prev.Code == "" && sh.Code != "" {
				seen[k] = sh
			}
			continue
		}
		seen[k] = sh
		order = append(order, k)
	}
	out := make([]Shift, 0, len(order))
	for _, k := range order {
		out = append(out, seen[k])
	}
	return out
}

func dayDate(d, m string, from, to time.Time) (time.Time, error) {
	day, err1 := strconv.Atoi(d)
	month, err2 := strconv.Atoi(m)
	if err1 != nil || err2 != nil {
		return time.Time{}, fmt.Errorf("bad date")
	}
	// Prefer year from report window.
	for _, y := range []int{from.Year(), to.Year(), from.Year() - 1, from.Year() + 1} {
		cand := time.Date(y, time.Month(month), day, 0, 0, 0, 0, time.Local)
		if !cand.Before(from.AddDate(0, 0, -1)) && !cand.After(to.AddDate(0, 0, 1)) {
			return cand, nil
		}
	}
	return time.Date(from.Year(), time.Month(month), day, 0, 0, 0, 0, time.Local), nil
}

func dmy(d, m, y string) (time.Time, error) {
	day, err1 := strconv.Atoi(d)
	month, err2 := strconv.Atoi(m)
	year, err3 := strconv.Atoi(y)
	if err1 != nil || err2 != nil || err3 != nil {
		return time.Time{}, fmt.Errorf("bad date")
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), nil
}

func normalizeClock(hhmm string) string {
	parts := strings.Split(hhmm, ":")
	if len(parts) != 2 {
		return hhmm
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return hhmm
	}
	return fmt.Sprintf("%02d:%02d", h, m)
}
