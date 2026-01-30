package llm

import "testing"

func TestParseEnrichment(t *testing.T) {
	in := `{"summary":"ok","risks":["a","a"," "],"checklist":["x","y"]}`
	e, err := ParseEnrichment(in)
	if err != nil {
		t.Fatalf("ParseEnrichment err=%v", err)
	}
	if e.Summary != "ok" {
		t.Fatalf("summary=%q", e.Summary)
	}
	if len(e.Risks) != 1 || e.Risks[0] != "a" {
		t.Fatalf("risks=%v", e.Risks)
	}
	if len(e.Checklist) != 2 {
		t.Fatalf("checklist=%v", e.Checklist)
	}
}
