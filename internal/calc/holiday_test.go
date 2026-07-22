package calc

import (
	"testing"
	"time"
)

func TestEasterSundayKnownYears(t *testing.T) {
	cases := map[int]string{
		2024: "2024-03-31",
		2025: "2025-04-20",
		2026: "2026-04-05",
		2027: "2027-03-28",
	}
	for year, want := range cases {
		got := easterSunday(year, time.UTC)
		if Key(got) != want {
			t.Fatalf("%d: got %s want %s", year, Key(got), want)
		}
	}
}

func TestHolidaysInYear2026IncludesIndependenceAndEaster(t *testing.T) {
	set := HolidaySet(2026, time.UTC)
	if set["2026-12-06"] != "Itsenäisyyspäivä" {
		t.Fatalf("independence=%q", set["2026-12-06"])
	}
	if set["2026-04-05"] != "Pääsiäispäivä" {
		t.Fatalf("easter=%q", set["2026-04-05"])
	}
	if set["2026-04-03"] != "Pitkäperjantai" {
		t.Fatalf("good friday=%q", set["2026-04-03"])
	}
	// Midsummer Eve 2026 = Friday 19 Jun
	if set["2026-06-19"] != "Juhannusaatto" {
		t.Fatalf("midsummer eve=%q", set["2026-06-19"])
	}
}

func TestMidsummerAndAllSaints(t *testing.T) {
	if Key(midsummerEve(2025, time.UTC)) != "2025-06-20" {
		t.Fatalf("midsummer 2025=%s", Key(midsummerEve(2025, time.UTC)))
	}
	if Key(allSaintsDay(2026, time.UTC)) != "2026-10-31" {
		t.Fatalf("all saints 2026=%s", Key(allSaintsDay(2026, time.UTC)))
	}
}
