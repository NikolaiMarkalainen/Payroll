package ui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	persistStateFile    = "state.json"
	persistStateVersion = 1
)

// persistedState is the on-disk app snapshot.
type persistedState struct {
	Version  int                    `json:"version"`
	Settings persistedSettings      `json:"settings"`
	Shifts   persistedShifts        `json:"shifts"`
	Calc     persistedCalc          `json:"calc"`
}

type persistedSettings struct {
	TESFamily        string            `json:"tesFamily"`
	TESLevel         string            `json:"tesLevel"`
	TESRegion        string            `json:"tesRegion"`
	ExperienceTier   string            `json:"experienceTier"`
	HourlyWage       string            `json:"hourlyWage"`
	LevelPay         string            `json:"levelPay"`
	Experience       string            `json:"experience"`
	Personal         string            `json:"personal"`
	TrainingEnabled  bool              `json:"trainingEnabled"`
	Training         string            `json:"training"`
	OtherMode        string            `json:"otherMode"`
	OtherAllowance   string            `json:"otherAllowance"`
	Evening          string            `json:"evening"`
	EveningDouble    string            `json:"eveningDouble"`
	Night            string            `json:"night"`
	Saturday         string            `json:"saturday"`
	Sunday           string            `json:"sunday"`
	Holiday          string            `json:"holiday"`
	Perehdytys       string            `json:"perehdytys"`
	DailyRest        string            `json:"dailyRest"`
	RestViolation    string            `json:"restViolation"`
	Overtime50After  string            `json:"overtime50After"`
	Overtime100After string            `json:"overtime100After"`
	WeeklyOTThreshold string           `json:"weeklyOTThreshold"`
	EveningStart     string            `json:"eveningStart"`
	EveningEnd       string            `json:"eveningEnd"`
	NightStart       string            `json:"nightStart"`
	NightEnd         string            `json:"nightEnd"`
	SaturdayStart    string            `json:"saturdayStart"`
	SaturdayEnd      string            `json:"saturdayEnd"`
	ShiftOTEnabled   bool              `json:"shiftOTEnabled"`
	WeeklyOTEnabled  bool              `json:"weeklyOTEnabled"`
	EveExcludeSat    bool              `json:"eveningExcludeSaturday"`
	NightExcludeSun  bool              `json:"nightExcludeSunday"`
	NightExcludeHol  bool              `json:"nightExcludeHoliday"`
	EveDoubleFrom    int               `json:"eveningDoubleMonthFrom"`
	EveDoubleTo      int               `json:"eveningDoubleMonthTo"`
	EveDoubleSunOnly bool              `json:"eveningDoubleSundayOnly"`
	CalloutFixedH    float64           `json:"calloutFixedH"`
	PeriodOTEnabled  bool              `json:"periodOTEnabled"`
	PeriodMode       string            `json:"periodMode"`
	PeriodFirstThr   string            `json:"periodFirstThreshold"`
	PeriodAnchor     string            `json:"periodAnchor"`
	ColorTitles      bool              `json:"colorShiftTitles"`
	ColorOverrides   map[string]string `json:"colorOverrides"` // key -> #RRGGBB
	ColorManual      []string          `json:"colorManual"`
}

type persistedShift struct {
	ID              int    `json:"id"`
	Date            string `json:"date"` // YYYY-MM-DD
	Start           string `json:"start"`
	End             string `json:"end"`
	Callout         bool   `json:"callout"`
	Code            string `json:"code,omitempty"`
	PerehdytysStart string `json:"perehdytysStart,omitempty"`
	PerehdytysEnd   string `json:"perehdytysEnd,omitempty"`
}

type persistedShifts struct {
	Month  string           `json:"month"` // YYYY-MM-01
	NextID int              `json:"nextID"`
	Items  []persistedShift `json:"items"`
}

type persistedCalc struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Absence  string `json:"absence"`
	Anchor   string `json:"anchor,omitempty"`
}

func defaultDataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	// Prefer XDG data for app content; fall back to config if data dir missing.
	if data := os.Getenv("XDG_DATA_HOME"); data != "" {
		return filepath.Join(data, appID), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(base, appID), nil
	}
	return filepath.Join(home, ".local", "share", appID), nil
}

func statePath(dir string) string {
	return filepath.Join(dir, persistStateFile)
}

func loadPersistedState(dir string) (*persistedState, error) {
	path := statePath(dir)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var st persistedState
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, fmt.Errorf("state.json: %w", err)
	}
	return &st, nil
}

func savePersistedState(dir string, st *persistedState) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	st.Version = persistStateVersion
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp := statePath(dir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, statePath(dir))
}

func colorToHex(c color.NRGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func hexToColor(s string) (color.NRGBA, bool) {
	s = strings.TrimSpace(s)
	if len(s) != 7 || s[0] != '#' {
		return color.NRGBA{}, false
	}
	var r, g, b int
	if _, err := fmt.Sscanf(s[1:], "%02x%02x%02x", &r, &g, &b); err != nil {
		return color.NRGBA{}, false
	}
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xff}, true
}

// appPersister loads/saves settings, shifts and calc range.
type appPersister struct {
	dir      string
	settings *settingsTab
	shifts   *shiftsTab
	calc     *calcTab

	mu       sync.Mutex
	timer    *time.Timer
	loading  bool
	disabled bool
}

func newAppPersister(dir string, settings *settingsTab, shifts *shiftsTab, calc *calcTab) *appPersister {
	return &appPersister{dir: dir, settings: settings, shifts: shifts, calc: calc}
}

func (p *appPersister) load() error {
	if p == nil || p.disabled {
		return nil
	}
	st, err := loadPersistedState(p.dir)
	if err != nil || st == nil {
		return err
	}
	p.mu.Lock()
	p.loading = true
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		p.loading = false
		p.mu.Unlock()
	}()

	if p.settings != nil {
		p.settings.applyPersisted(st.Settings)
	}
	if p.shifts != nil {
		p.shifts.applyPersisted(st.Shifts)
	}
	if p.calc != nil {
		p.calc.applyPersisted(st.Calc)
	}
	return nil
}

func (p *appPersister) snapshot() *persistedState {
	st := &persistedState{Version: persistStateVersion}
	if p.settings != nil {
		st.Settings = p.settings.exportPersisted()
	}
	if p.shifts != nil {
		st.Shifts = p.shifts.exportPersisted()
	}
	if p.calc != nil {
		st.Calc = p.calc.exportPersisted()
	}
	return st
}

func (p *appPersister) saveNow() error {
	if p == nil || p.disabled {
		return nil
	}
	p.mu.Lock()
	if p.loading {
		p.mu.Unlock()
		return nil
	}
	p.mu.Unlock()
	return savePersistedState(p.dir, p.snapshot())
}

func (p *appPersister) scheduleSave() {
	if p == nil || p.disabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.loading {
		return
	}
	if p.timer != nil {
		p.timer.Stop()
	}
	p.timer = time.AfterFunc(400*time.Millisecond, func() {
		_ = p.saveNow()
	})
}

func (p *appPersister) flush() {
	if p == nil || p.disabled {
		return
	}
	p.mu.Lock()
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	p.mu.Unlock()
	_ = p.saveNow()
}
