package reco

import (
	"path/filepath"
	"testing"
	"time"
)

func TestWriteReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "optimizer.reco.json")
	in := Recommendation{
		Version:          "reco.v1",
		GeneratedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPaper:       "paper.jsonl",
		InputLabels:      "labels.repo.jsonl",
		PrimaryWindowSec: 30,
		Slots:            30,
		Quotas: []SignalQuota{
			{Signal: "a", MeanReward: 0.5, N: 10, SuggestedDailyQuota: 12},
		},
	}

	if err := Write(path, in); err != nil {
		t.Fatalf("Write err=%v", err)
	}
	out, err := Read(path)
	if err != nil {
		t.Fatalf("Read err=%v", err)
	}
	if out.Version != in.Version || out.PrimaryWindowSec != in.PrimaryWindowSec || out.Slots != in.Slots {
		t.Fatalf("roundtrip mismatch out=%+v in=%+v", out, in)
	}
	if len(out.Quotas) != 1 || out.Quotas[0].Signal != "a" || out.Quotas[0].SuggestedDailyQuota != 12 {
		t.Fatalf("quotas mismatch out=%+v", out.Quotas)
	}
}
