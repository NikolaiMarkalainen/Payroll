package pdfimport

import "time"

// Shift is one realized work interval from a TyövuoroVelho report.
type Shift struct {
	Date    time.Time
	Start   string // HH:MM or 24:00
	End     string
	Callout bool
	Code    string // place/code label from report, if any
}

// Result is the parsed roster import.
type Result struct {
	From     time.Time
	To       time.Time
	Period   string // e.g. 13/2026
	Person   string
	Shifts   []Shift
	Warnings []string
}
