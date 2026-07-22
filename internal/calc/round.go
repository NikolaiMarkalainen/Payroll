package calc

import (
	"math"
	"time"
)

func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func roundHours(h float64) float64 {
	return math.Round(h*100) / 100
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func roundDay(dh *DayHours) {
	dh.Total = roundHours(dh.Total)
	dh.Evening = roundHours(dh.Evening)
	dh.EveningDouble = roundHours(dh.EveningDouble)
	dh.Night = roundHours(dh.Night)
	dh.Saturday = roundHours(dh.Saturday)
	dh.Sunday = roundHours(dh.Sunday)
	dh.Holiday = roundHours(dh.Holiday)
	dh.Callout = roundHours(dh.Callout)
	dh.Overtime50 = roundHours(dh.Overtime50)
	dh.Overtime100 = roundHours(dh.Overtime100)
}
