package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var finnishWeekdays = []string{"Ma", "Ti", "Ke", "To", "Pe", "La", "Su"}

var finnishMonths = []string{
	"", "tammikuu", "helmikuu", "maaliskuu", "huhtikuu", "toukokuu", "kesäkuu",
	"heinäkuu", "elokuu", "syyskuu", "lokakuu", "marraskuu", "joulukuu",
}

// calendarShift is a shift shown inside a day cell.
type calendarShift struct {
	Date  time.Time
	Start string
	End   string
}

type shiftsTab struct {
	month   time.Time
	shifts  []calendarShift
	title   *widget.Label
	grid    *fyne.Container
	content fyne.CanvasObject
}

func newShiftsTab() *shiftsTab {
	now := time.Now()
	s := &shiftsTab{
		month: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
		title: widget.NewLabel(""),
	}
	s.title.TextStyle = fyne.TextStyle{Bold: true}
	s.title.Alignment = fyne.TextAlignCenter
	s.grid = container.NewGridWithColumns(7)
	s.refresh()
	return s
}

func (s *shiftsTab) canvas() fyne.CanvasObject {
	hint := widget.NewLabel("Kuukausinäkymä. Vuorojen ajat näkyvät solussa, kun vuoroja on lisätty. Päiväklikkaus tulee seuraavassa vaiheessa.")
	hint.Wrapping = fyne.TextWrapWord

	heading := widget.NewLabel("Vuorokalenteri")
	heading.TextStyle = fyne.TextStyle{Bold: true}

	prev := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		s.month = s.month.AddDate(0, -1, 0)
		s.refresh()
	})
	next := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		s.month = s.month.AddDate(0, 1, 0)
		s.refresh()
	})
	nav := container.NewBorder(nil, nil, prev, next, s.title)

	headers := make([]fyne.CanvasObject, 7)
	for i, name := range finnishWeekdays {
		l := widget.NewLabel(name)
		l.TextStyle = fyne.TextStyle{Bold: true}
		l.Alignment = fyne.TextAlignCenter
		headers[i] = l
	}
	headerRow := container.NewGridWithColumns(7, headers...)

	s.content = container.NewBorder(
		container.NewVBox(heading, hint, widget.NewSeparator(), nav, headerRow),
		nil,
		nil,
		nil,
		container.NewPadded(s.grid),
	)
	return s.content
}

func (s *shiftsTab) refresh() {
	s.title.SetText(fmt.Sprintf("%s %d", finnishMonths[s.month.Month()], s.month.Year()))

	cells := make([]fyne.CanvasObject, 0, 42)
	first := time.Date(s.month.Year(), s.month.Month(), 1, 0, 0, 0, 0, s.month.Location())
	// Monday=0 ... Sunday=6
	startOffset := (int(first.Weekday()) + 6) % 7

	for i := 0; i < startOffset; i++ {
		cells = append(cells, s.emptyCell())
	}

	daysInMonth := daysIn(s.month.Year(), s.month.Month())
	today := time.Now()
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(s.month.Year(), s.month.Month(), day, 0, 0, 0, 0, s.month.Location())
		isToday := sameDate(date, today)
		cells = append(cells, s.dayCell(date, isToday))
	}
	for len(cells)%7 != 0 {
		cells = append(cells, s.emptyCell())
	}

	s.grid.Objects = cells
	s.grid.Refresh()
}

func (s *shiftsTab) dayCell(date time.Time, isToday bool) fyne.CanvasObject {
	dayLabel := widget.NewLabel(fmt.Sprintf("%d", date.Day()))
	dayLabel.TextStyle = fyne.TextStyle{Bold: true}
	if isToday {
		dayLabel.Importance = widget.HighImportance
	}

	items := []fyne.CanvasObject{dayLabel}
	for _, sh := range s.shiftsOn(date) {
		t := widget.NewLabel(sh.Start + "–" + sh.End)
		t.TextStyle = fyne.TextStyle{Italic: true}
		items = append(items, t)
	}
	if len(items) == 1 {
		// Keep cell height more even when empty of shifts.
		items = append(items, widget.NewLabel(""))
	}

	inner := container.NewVBox(items...)
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 4
	if isToday {
		bg.StrokeColor = theme.Color(theme.ColorNamePrimary)
		bg.StrokeWidth = 1.5
	}

	return container.NewStack(bg, container.NewPadded(inner))
}

func (s *shiftsTab) emptyCell() fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	bg.CornerRadius = 4
	return container.NewStack(bg, container.NewPadded(widget.NewLabel("")))
}

func (s *shiftsTab) shiftsOn(date time.Time) []calendarShift {
	var out []calendarShift
	for _, sh := range s.shifts {
		if sameDate(sh.Date, date) {
			out = append(out, sh)
		}
	}
	return out
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func finnishMonthName(m time.Month) string {
	if m < 1 || int(m) >= len(finnishMonths) {
		return ""
	}
	return finnishMonths[m]
}
