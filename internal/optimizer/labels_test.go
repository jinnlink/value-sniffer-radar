package optimizer

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestReadLabelsJSONL_DefaultPrimaryWindowSec(t *testing.T) {
	in := "\ufeff" + `{"event_id":"e1","source":"s","symbol":"x","trade_date":"20260101","window_sec":10,"reward":1}` + "\n" +
		`{"event_id":"e1","source":"s","symbol":"x","trade_date":"20260101","window_sec":30,"reward":0}` + "\n" +
		`{"event_id":"e2","source":"s","symbol":"x","trade_date":"20260101","window_sec":30,"reward":1}` + "\n"

	li, err := ReadLabelsJSONL(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ReadLabelsJSONL err=%v", err)
	}
	if got := li.DefaultPrimaryWindowSec(); got != 30 {
		t.Fatalf("DefaultPrimaryWindowSec=%d want=30", got)
	}
	if !li.Has("e1", 10) || !li.Has("e1", 30) {
		t.Fatalf("expected labels to be indexed by event/window")
	}
}

func TestOptimizerRankingChangesWithLabels(t *testing.T) {
	makeRow := func(ts string, source string, reward int) PaperRow {
		return PaperRow{
			TS: ts,
			Event: PaperLogEvent{
				Source:    source,
				TradeDate: "20260101",
				Market:    "cn",
				Symbol:    "204001",
				Title:     fmt.Sprintf("%s-%s", source, ts),
				Data:      map[string]any{"reward": reward},
				Tags:      map[string]string{"kind": "repo"},
			},
		}
	}

	var rows []PaperRow
	for i := 0; i < 50; i++ {
		rows = append(rows, makeRow(fmt.Sprintf("2026-01-01T00:00:%02dZ", i), "sig_A", 1))
		rows = append(rows, makeRow(fmt.Sprintf("2026-01-01T00:01:%02dZ", i), "sig_B", 0))
	}

	// Without labels: sig_A should win.
	{
		b := NewBandit()
		for _, pr := range rows {
			r, ok, _ := ResolveReward(pr, LabelsIndex{}, 0)
			if ok {
				b.Update(pr.Event.Source, r)
			} else {
				b.Ensure(pr.Event.Source)
			}
		}
		alloc, err := b.SuggestAllocation(rand.New(rand.NewSource(7)), 1)
		if err != nil {
			t.Fatalf("SuggestAllocation err=%v", err)
		}
		if len(alloc) != 1 || alloc[0].Key != "sig_A" {
			t.Fatalf("paper-only top=%v want=sig_A", alloc)
		}
	}

	// With labels (window=30): invert outcomes so sig_B should win.
	{
		var sb strings.Builder
		for _, pr := range rows {
			want := 0
			if pr.Event.Source == "sig_B" {
				want = 1
			}
			sb.WriteString(fmt.Sprintf(`{"event_id":"%s","source":"%s","symbol":"%s","trade_date":"%s","window_sec":30,"reward":%d}`+"\n",
				EventID(pr), pr.Event.Source, pr.Event.Symbol, pr.Event.TradeDate, want))
		}
		li, err := ReadLabelsJSONL(strings.NewReader(sb.String()))
		if err != nil {
			t.Fatalf("ReadLabelsJSONL err=%v", err)
		}

		b := NewBandit()
		for _, pr := range rows {
			r, ok, _ := ResolveReward(pr, li, 30)
			if ok {
				b.Update(pr.Event.Source, r)
			} else {
				b.Ensure(pr.Event.Source)
			}
		}
		alloc, err := b.SuggestAllocation(rand.New(rand.NewSource(7)), 1)
		if err != nil {
			t.Fatalf("SuggestAllocation err=%v", err)
		}
		if len(alloc) != 1 || alloc[0].Key != "sig_B" {
			t.Fatalf("labels top=%v want=sig_B", alloc)
		}
	}
}
