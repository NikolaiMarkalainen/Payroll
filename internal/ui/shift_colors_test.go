package ui

import "testing"

func TestShiftColorKeyGroupsByLeadingNumber(t *testing.T) {
	cases := []struct {
		code string
		want string
	}{
		{"3AAA", "3"},
		{"3AAB", "3"},
		{"2BBB", "2"},
		{"2BBC", "2"},
		{"1LOC", "1"},
		{"11XYZ", "11"},
		{"ALPHA", "ALPHA"},
		{"EXTRA", "EXTRA"},
		{"extra", "EXTRA"},
		{"BETA1X", "BETA1X"},
		{"", ""},
		{"  3AAA ", "3"},
	}
	for _, tc := range cases {
		if got := shiftColorKey(tc.code); got != tc.want {
			t.Errorf("shiftColorKey(%q)=%q want %q", tc.code, got, tc.want)
		}
	}
}

func TestShiftTitleColorSameNumberSameColor(t *testing.T) {
	a := shiftTitleColor("3AAA")
	b := shiftTitleColor("3AAB")
	if a != b {
		t.Fatalf("3AAA=%v 3AAB=%v want same", a, b)
	}
	c := shiftTitleColor("2BBB")
	if a == c {
		t.Fatal("group 3 and 2 should differ")
	}
}

func TestShiftTitleColorLetterCodesHaveOwnKeys(t *testing.T) {
	if shiftColorKey("ALPHA") == shiftColorKey("EXTRA") {
		t.Fatal("letter codes should not share a key")
	}
	if shiftColorKey("ALPHA") == shiftColorKey("3AAA") {
		t.Fatal("letter code must not collapse to a number group")
	}
	_ = shiftTitleColor("ALPHA")
	_ = shiftTitleColor("EXTRA")
}
