package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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
	ID              int
	Date            time.Time
	Start           string
	End             string
	Callout         bool   // Hälytysvuoro
	Code            string // roster place/code label from import, if any
	PerehdytysStart string // orientation mentoring start HH:MM; empty = none
	PerehdytysEnd   string // orientation mentoring end HH:MM
}

type shiftsTab struct {
	window fyne.Window
	month time.Time
	shifts []calendarShift
	nextID int
	title *widget.Label
	grid *fyne.Container
	content fyne.CanvasObject
	rules func() allowanceRules
	colorTitles func() bool
	colorFor    func(code string) color.NRGBA
	periodAnchor    func() (time.Time, bool)
	periodThreshold func(day time.Time) (float64, bool)
	selectedPeriod  func() (from, to time.Time, ok bool)
	onChanged   func()
	onDemoLoaded func()
}

func newShiftsTab(w fyne.Window) *shiftsTab {
	now := time.Now()
	s := &shiftsTab{
		window: w,
		month: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
		title: widget.NewLabel(""),
		nextID: 1,
		rules: defaultAllowanceRules,
	}
	s.title.TextStyle = fyne.TextStyle{Bold: true}
	s.title.Alignment = fyne.TextAlignCenter
	s.grid = container.NewGridWithColumns(7)
	s.refresh()
	return s
}

func (s *shiftsTab) loadDemoSeed() {
	loc := time.Local
	if s.month.Location() != nil {
		loc = s.month.Location()
	}
	s.replaceShifts(demoRosterShifts(loc), demoFocusMonth(loc))
	if s.onDemoLoaded != nil {
		s.onDemoLoaded()
	}
}

func (s *shiftsTab) replaceShifts(shifts []calendarShift, month time.Time) {
	s.shifts = make([]calendarShift, 0, len(shifts))
	s.nextID = 1
	for _, sh := range shifts {
		sh.ID = s.nextID
		s.nextID++
		s.shifts = append(s.shifts, sh)
	}
	s.month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	s.refresh()
}

// mergeImportedShifts keeps existing calendar days and overlays incoming shifts
// only on the days present in the import (so a new PDF does not wipe other periods).
func (s *shiftsTab) mergeImportedShifts(incoming []calendarShift, focusMonth time.Time) (kept, replacedDays, added int) {
	incomingDays := map[string]bool{}
	for _, sh := range incoming {
		incomingDays[calendarDayKey(sh.Date)] = true
	}
	replacedDays = len(incomingDays)

	keptShifts := make([]calendarShift, 0, len(s.shifts))
	for _, sh := range s.shifts {
		key := calendarDayKey(sh.Date)
		if incomingDays[key] {
			continue
		}
		keptShifts = append(keptShifts, sh)
		kept++
	}
	added = len(incoming)
	merged := make([]calendarShift, 0, len(keptShifts)+len(incoming))
	merged = append(merged, keptShifts...)
	merged = append(merged, incoming...)
	s.replaceShifts(merged, focusMonth)
	return kept, replacedDays, added
}

func calendarDayKey(t time.Time) string {
	return t.Format("2006-01-02")
}

func (s *shiftsTab) currentRules() allowanceRules {
	if s.rules != nil {
		return s.rules()
	}
	return defaultAllowanceRules()
}

func (s *shiftsTab) canvas() fyne.CanvasObject {
	hint := widget.NewLabel("Klikkaa päivää lisätäksesi vuoron. Klikkaa vuoroa muokataksesi. Poista X-napilla. J-numerot = 3 vk jaksot (ylityökorvaukset lasketaan jaksoittain).")
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
		container.NewVBox(
			heading,
			hint,
			allowanceLegend(),
			periodLegend(),
			widget.NewSeparator(),
			nav,
			headerRow,
		),
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
	if s.onChanged != nil {
		s.onChanged()
	}
}

func (s *shiftsTab) dayCell(date time.Time, isToday bool) fyne.CanvasObject {
	allow := summarizeDay(date, s.shifts, s.currentRules())
	items := []fyne.CanvasObject{s.dayHeader(date, isToday)}
	if chips := allow.chips(); len(chips) > 0 {
		items = append(items, chipRow(chips))
	}
	for _, seg := range s.segmentsOn(date) {
		items = append(items, s.shiftRow(seg))
	}
	// Pad empty area so the cell stays tall and easy to hit.
	items = append(items, layout.NewSpacer())

	inner := container.NewVBox(items...)
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 4
	if isToday {
		bg.StrokeColor = theme.Color(theme.ColorNamePrimary)
		bg.StrokeWidth = 1.5
	}

	d := date
	cell := container.NewStack(bg, container.NewPadded(inner))
	return newTapBox(cell, func() {
		s.openAddShiftDialog(d)
	}).withMinSize(0, 96)
}

func (s *shiftsTab) shiftRow(seg shiftSegment) fyne.CanvasObject {
	body := s.shiftLabel(seg)

	target := seg.Shift
	edit := newTapBox(body, func() {
		s.openEditShiftDialog(target)
	})

	del := widget.NewButton("X", func() {
		s.confirmDeleteShift(target)
	})
	del.Importance = widget.LowImportance

	return container.NewBorder(nil, nil, nil, del, edit)
}

func (s *shiftsTab) shiftLabel(seg shiftSegment) fyne.CanvasObject {
	colorOn := s.colorTitles != nil && s.colorTitles() && seg.Shift.Code != ""
	titleColor := shiftTitleColor(seg.Shift.Code)
	if s.colorFor != nil {
		titleColor = s.colorFor(seg.Shift.Code)
	}

	spanObj := func() fyne.CanvasObject {
		if colorOn {
			txt := canvas.NewText(seg.Span, titleColor)
			txt.TextStyle = fyne.TextStyle{Italic: true}
			txt.TextSize = theme.TextSize()
			return txt
		}
		l := widget.NewLabel(seg.Span)
		l.TextStyle = fyne.TextStyle{Italic: true}
		return l
	}

	if seg.Title == "" {
		return spanObj()
	}

	var titleObj fyne.CanvasObject
	if colorOn {
		txt := canvas.NewText(seg.Title, titleColor)
		txt.TextStyle = fyne.TextStyle{Bold: true}
		txt.TextSize = theme.TextSize()
		titleObj = txt
	} else {
		l := widget.NewLabel(seg.Title)
		l.TextStyle = fyne.TextStyle{Bold: true}
		titleObj = l
	}
	return container.NewVBox(titleObj, spanObj())
}

func (s *shiftsTab) emptyCell() fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	bg.CornerRadius = 4
	return container.NewStack(bg, container.NewPadded(widget.NewLabel("")))
}

func (s *shiftsTab) openAddShiftDialog(date time.Time) {
	s.openShiftDialog(date, nil)
}

func (s *shiftsTab) openEditShiftDialog(existing calendarShift) {
	s.openShiftDialog(existing.Date, &existing)
}

func (s *shiftsTab) openShiftDialog(date time.Time, existing *calendarShift) {
	if s.window == nil {
		return
	}

	startVal, endVal := "06:00", "14:00"
	calloutChecked := false
	pereChecked := false
	pereStartVal, pereEndVal := startVal, endVal
	editing := existing != nil
	if editing {
		startVal = existing.Start
		endVal = existing.End
		calloutChecked = existing.Callout
		date = existing.Date
		pereStartVal, pereEndVal = existing.Start, existing.End
		if existing.PerehdytysStart != "" && existing.PerehdytysEnd != "" {
			pereChecked = true
			pereStartVal = existing.PerehdytysStart
			pereEndVal = existing.PerehdytysEnd
		}
	}

	startPicker := newTimePicker(startVal)
	endPicker := newTimePicker(endVal)
	calloutLabel := "Hälytystyö"
	if fixed := s.currentRules().calloutFixedH; fixed > 0 {
		calloutLabel = fmt.Sprintf("Hälytystyö (kiinteä %.0f h palkka)", fixed)
	}
	callout := widget.NewCheck(calloutLabel, nil)
	callout.SetChecked(calloutChecked)

	pereStart := newTimePicker(pereStartVal)
	pereEnd := newTimePicker(pereEndVal)
	pereTimes := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Perehdytys alkaa", pereStart.canvas()),
			widget.NewFormItem("Perehdytys loppuu", pereEnd.canvas()),
		),
	)
	if !pereChecked {
		pereTimes.Hide()
	}
	// Default perehdytys span = whole shift when the box is turned on
	// (full-day mentoring without editing times). Keep saved times on open.
	pereWasOn := pereChecked
	pere := widget.NewCheck("Perehdytyslisä", func(on bool) {
		if on {
			if !pereWasOn {
				if startPicker.valid() && endPicker.valid() {
					pereStart.set(startPicker.value())
					pereEnd.set(endPicker.value())
				}
			}
			pereWasOn = true
			pereTimes.Show()
		} else {
			pereWasOn = false
			pereTimes.Hide()
		}
	})
	pere.SetChecked(pereChecked)

	formErr := widget.NewLabel("")
	formErr.Importance = widget.DangerImportance
	formErr.Hide()

	dateLabel := widget.NewLabel(fmt.Sprintf("Päivä: %s", date.Format("02.01.2006")))
	dateLabel.TextStyle = fyne.TextStyle{Bold: true}

	form := widget.NewForm(
		widget.NewFormItem("Alkaa", startPicker.canvas()),
		widget.NewFormItem("Loppuu", endPicker.canvas()),
		widget.NewFormItem("", callout),
		widget.NewFormItem("", pere),
	)

	title, confirm := "Lisää vuoro", "Lisää"
	if editing {
		title, confirm = "Muokkaa vuoroa", "Tallenna"
	}

	body := container.NewVBox(dateLabel, form, pereTimes, formErr)
	var d dialog.Dialog
	d = dialog.NewCustomConfirm(
		title,
		confirm,
		"Peruuta",
		body,
		func(ok bool) {
			if !ok {
				return
			}
			startPicker.refreshError()
			endPicker.refreshError()
			if !startPicker.valid() || !endPicker.valid() {
				d.Show()
				return
			}
			sh := calendarShift{
				Date:    date,
				Start:   startPicker.value(),
				End:     endPicker.value(),
				Callout: callout.Checked,
			}
			if pere.Checked {
				pereStart.refreshError()
				pereEnd.refreshError()
				if !pereStart.valid() || !pereEnd.valid() {
					formErr.SetText("Perehdytyksen ajat eivät kelpaa")
					formErr.Show()
					d.Show()
					return
				}
				sh.PerehdytysStart = pereStart.value()
				sh.PerehdytysEnd = pereEnd.value()
				if h, err := clockSpanHours(sh.PerehdytysStart, sh.PerehdytysEnd); err != nil || h <= 0 {
					formErr.SetText("Perehdytyksen kesto pitää olla yli 0 h")
					formErr.Show()
					d.Show()
					return
				}
			}
			var err error
			if editing {
				sh.ID = existing.ID
				sh.Code = existing.Code
				err = s.updateShift(sh)
			} else {
				err = s.addShift(sh)
			}
			if err != nil {
				formErr.SetText(err.Error())
				formErr.Show()
				d.Show()
				return
			}
		},
		s.window,
	)
	d.Resize(fyne.NewSize(400, 420))
	d.Show()
}

func (s *shiftsTab) confirmDeleteShift(sh calendarShift) {
	if s.window == nil {
		s.removeShift(sh.ID)
		return
	}
	msg := fmt.Sprintf("Oletko varma, että haluat poistaa vuoron %s-%s?", sh.Start, sh.End)
	if sh.Callout {
		msg = fmt.Sprintf("Oletko varma, että haluat poistaa hälytysvuoron %s-%s?", sh.Start, sh.End)
	}
	dialog.ShowConfirm("Poista vuoro", msg, func(ok bool) {
		if ok {
			s.removeShift(sh.ID)
		}
	}, s.window)
}

func (s *shiftsTab) addShift(sh calendarShift) error {
	if err := s.validateShift(sh); err != nil {
		return err
	}
	sh.ID = s.nextID
	s.nextID++
	s.shifts = append(s.shifts, sh)
	s.refresh()
	return nil
}

func (s *shiftsTab) updateShift(sh calendarShift) error {
	if sh.ID == 0 {
		return fmt.Errorf("vuoroa ei löydy")
	}
	if err := s.validateShift(sh); err != nil {
		return err
	}
	for i := range s.shifts {
		if s.shifts[i].ID == sh.ID {
			s.shifts[i] = sh
			s.refresh()
			return nil
		}
	}
	return fmt.Errorf("vuoroa ei löydy")
}

func (s *shiftsTab) validateShift(sh calendarShift) error {
	if sh.Start == sh.End {
		return fmt.Errorf("alku ja loppu eivät voi olla sama aika")
	}
	if _, _, err := sh.absoluteRange(); err != nil {
		return fmt.Errorf("virheellinen vuoroaika")
	}
	if s.overlapsExisting(sh) {
		return fmt.Errorf("vuoro menee päällekkäin toisen vuoron kanssa")
	}
	return nil
}

func (s *shiftsTab) removeShift(id int) {
	out := s.shifts[:0]
	for _, sh := range s.shifts {
		if sh.ID != id {
			out = append(out, sh)
		}
	}
	s.shifts = out
	s.refresh()
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
