package optimizer

import (
	"strings"
	"testing"
)

func TestReadJSONLAndReward(t *testing.T) {
	in := `{"ts":"x","event":{"source":"a","data":{"reward":1}}}
{"ts":"y","event":{"source":"b","data":{"reward":0}}}
{"ts":"z","event":{"source":"c","data":{"reward":"true"}}}
`
	rows, warns, err := ReadJSONL(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(warns) != 0 {
		t.Fatalf("unexpected warns: %v", warns)
	}
	if len(rows) != 3 {
		t.Fatalf("rows=%d", len(rows))
	}
	r, ok := RewardFromRow(rows[0])
	if !ok || r != 1 {
		t.Fatalf("reward0=%d ok=%v", r, ok)
	}
	r, ok = RewardFromRow(rows[1])
	if !ok || r != 0 {
		t.Fatalf("reward1=%d ok=%v", r, ok)
	}
	r, ok = RewardFromRow(rows[2])
	if !ok || r != 1 {
		t.Fatalf("reward2=%d ok=%v", r, ok)
	}
}

