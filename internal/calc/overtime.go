package calc

import "time"

// overtime.go — daily, shift-extension, and weekly overtime helpers.

func splitOvertime(total, after50, after100 float64) (ot50, ot100 float64) {
	if after50 < 0 {
		after50 = 0
	}
	if total <= after50 {
		return 0, 0
	}
	// after100 <= after50 means "only 50% band" (e.g. Kaupan >10 h → 50%).
	if after100 <= after50 {
		return total - after50, 0
	}
	if total <= after100 {
		return total - after50, 0
	}
	return after100 - after50, total - after100
}

// weeklyOT50Hours returns hours over threshold per ISO week, minus daily/shift OT already
// attributed on days in that week (avoids double OT premium on the same hour).
func weeklyOT50Hours(days []DayHours, threshold float64) float64 {
	if threshold <= 0 || len(days) == 0 {
		return 0
	}
	type agg struct{ total, ot float64 }
	byWeek := map[[2]int]*agg{}
	for _, dh := range days {
		y, w := dh.Date.ISOWeek()
		k := [2]int{y, w}
		a := byWeek[k]
		if a == nil {
			a = &agg{}
			byWeek[k] = a
		}
		a.total += dh.Total
		a.ot += dh.Overtime50 + dh.Overtime100
	}
	var extra float64
	for _, a := range byWeek {
		over := a.total - threshold
		if over <= 0 {
			continue
		}
		add := over - a.ot
		if add > 0 {
			extra += add
		}
	}
	return extra
}

type extSlice struct {
	start time.Time
	end   time.Time
}

type extHours struct {
	total  float64
	slices []extSlice
}

// collectShiftExtensionHours finds the tail of each shift beyond afterH, clipped to window.
// Uses full (unclipped) shift duration to decide extension, then clips the extension span.
func collectShiftExtensionHours(shifts []Shift, windowStart, windowEnd time.Time, afterH float64) extHours {
	var out extHours
	if afterH <= 0 {
		return out
	}
	for _, sh := range shifts {
		if !sh.Start.Before(sh.End) {
			continue
		}
		dur := sh.End.Sub(sh.Start).Hours()
		if dur <= afterH {
			continue
		}
		extH := dur - afterH
		extStart := sh.End.Add(-time.Duration(extH * float64(time.Hour)))
		extEnd := sh.End
		if extStart.Before(windowStart) {
			extStart = windowStart
		}
		if extEnd.After(windowEnd) {
			extEnd = windowEnd
		}
		if !extStart.Before(extEnd) {
			continue
		}
		out.total += extEnd.Sub(extStart).Hours()
		out.slices = append(out.slices, extSlice{start: extStart, end: extEnd})
	}
	return out
}

// applyExtensionOTToDays attributes first capH of extension chronologically as 50%, rest 100%.
func applyExtensionOTToDays(days []DayHours, slices []extSlice, capH float64) {
	if len(days) == 0 || len(slices) == 0 {
		return
	}
	byKey := map[string]*DayHours{}
	for i := range days {
		days[i].Overtime50 = 0
		days[i].Overtime100 = 0
		byKey[Key(days[i].Date)] = &days[i]
	}
	remaining50 := capH
	if remaining50 < 0 {
		remaining50 = 0
	}
	for _, sp := range slices {
		for day := truncateDay(sp.start); day.Before(sp.end); day = day.AddDate(0, 0, 1) {
			dayEnd := day.Add(24 * time.Hour)
			cs, ce := sp.start, sp.end
			if cs.Before(day) {
				cs = day
			}
			if ce.After(dayEnd) {
				ce = dayEnd
			}
			if !cs.Before(ce) {
				continue
			}
			h := ce.Sub(cs).Hours()
			dh := byKey[Key(day)]
			if dh == nil {
				continue
			}
			if remaining50 > 0 {
				take := h
				if take > remaining50 {
					take = remaining50
				}
				dh.Overtime50 += take
				h -= take
				remaining50 -= take
			}
			if h > 0 {
				dh.Overtime100 += h
			}
		}
	}
	for i := range days {
		days[i].Overtime50 = roundHours(days[i].Overtime50)
		days[i].Overtime100 = roundHours(days[i].Overtime100)
	}
}
