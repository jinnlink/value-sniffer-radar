package signals

import (
	"reflect"
	"testing"
	"time"
)

func TestNormalizeRepoCodes(t *testing.T) {
	in := []string{"SH204001", "SZ131810", "204001.SH", "131810.SZ", "204001", "131810", "  "}
	got := normalizeRepoCodes(in)
	want := []string{"204001.SH", "131810.SZ"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeRepoCodes: got=%v want=%v", got, want)
	}
}

func TestParseHHMM(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"14:45", 14*60 + 45, true},
		{"1445", 14*60 + 45, true},
		{"14：45", 14*60 + 45, true},
		{" 09:30 ", 9*60 + 30, true},
		{"9:30", 9*60 + 30, true},
		{"", 0, false},
		{"2460", 0, false},
		{"25:00", 0, false},
	}
	for _, c := range cases {
		got, ok := parseHHMM(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Fatalf("parseHHMM(%q): got=%d ok=%v want=%d ok=%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestWithinWindow(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	at := func(h, m int) time.Time {
		return time.Date(2026, 1, 29, h, m, 0, 0, loc)
	}

	if !withinWindow(at(14, 50), "14:45", "15:00") {
		t.Fatalf("expected within window")
	}
	if withinWindow(at(14, 30), "14:45", "15:00") {
		t.Fatalf("expected outside window")
	}
	if !withinWindow(at(2, 0), "23:00", "03:00") {
		t.Fatalf("expected within overnight window")
	}
	if withinWindow(at(12, 0), "23:00", "03:00") {
		t.Fatalf("expected outside overnight window")
	}
}
