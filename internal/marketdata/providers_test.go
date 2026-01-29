package marketdata

import "testing"

func TestEastmoneySecID(t *testing.T) {
	got, err := eastmoneySecID("204001.SH")
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.204001" {
		t.Fatalf("got=%s", got)
	}
	got, err = eastmoneySecID("131810.SZ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "0.131810" {
		t.Fatalf("got=%s", got)
	}
}

func TestTencentCode(t *testing.T) {
	got, err := tencentCode("204001.SH")
	if err != nil {
		t.Fatal(err)
	}
	if got != "sh204001" {
		t.Fatalf("got=%s", got)
	}
}

func TestParseTencentLine(t *testing.T) {
	line := `v_sh204001="1~GC001~204001~5.120~";`
	r, _, err := parseTencentLine(line)
	if err != nil {
		t.Fatal(err)
	}
	if r != 5.120 {
		t.Fatalf("got=%.3f", r)
	}
}

