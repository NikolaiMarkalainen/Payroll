package ui

import (
	"hash/fnv"
	"image/color"
	"strings"
	"unicode"
)

// Stable palette for shift title colors (calendar labels).
var shiftTitlePalette = []color.NRGBA{
	{R: 0x2f, G: 0x6f, B: 0xad, A: 0xff}, // blue
	{R: 0x2a, G: 0x7a, B: 0x6a, A: 0xff}, // teal
	{R: 0xb8, G: 0x86, B: 0x0b, A: 0xff}, // gold
	{R: 0xb3, G: 0x2d, B: 0x3a, A: 0xff}, // crimson
	{R: 0x6b, G: 0x3f, B: 0xa0, A: 0xff}, // purple
	{R: 0xc4, G: 0x5c, B: 0x26, A: 0xff}, // burnt orange
	{R: 0x3d, G: 0x7a, B: 0x3d, A: 0xff}, // green
	{R: 0x8b, G: 0x1e, B: 0x3f, A: 0xff}, // wine
	{R: 0x1a, G: 0x6b, B: 0x8a, A: 0xff}, // ocean
	{R: 0x8a, G: 0x5a, B: 0x2b, A: 0xff}, // amber brown
	{R: 0x5c, G: 0x3a, B: 0x21, A: 0xff}, // dark brown
	{R: 0x4a, G: 0x55, B: 0x68, A: 0xff}, // slate
}

var defaultShiftTitleColor = color.NRGBA{R: 0x44, G: 0x44, B: 0x44, A: 0xff}

// shiftColorKey groups codes for coloring:
//   - leading digits → that number ("3AAA" and "3AAB" → "3")
//   - no leading digits → full code ("ALPHA", "EXTRA" each own key)
func shiftColorKey(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	i := 0
	for _, r := range code {
		if !unicode.IsDigit(r) {
			break
		}
		i += len(string(r))
	}
	if i > 0 {
		return code[:i]
	}
	return strings.ToUpper(code)
}

func shiftTitleColor(code string) color.NRGBA {
	return shiftTitleColorFor(nil, code)
}

func shiftTitleColorFor(overrides map[string]color.NRGBA, code string) color.NRGBA {
	key := shiftColorKey(code)
	if key == "" {
		return defaultShiftTitleColor
	}
	if overrides != nil {
		if c, ok := overrides[key]; ok {
			return c
		}
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return shiftTitlePalette[int(h.Sum32())%len(shiftTitlePalette)]
}
