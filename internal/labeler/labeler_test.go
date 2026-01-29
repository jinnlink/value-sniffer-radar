package labeler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/marketdata"
)

type fakeFusion struct {
	rate float64
	conf marketdata.Confidence
}

func (f fakeFusion) FetchFusion(ctx context.Context, symbol string) (marketdata.FusionSnapshot, error) {
	_ = ctx
	return marketdata.FusionSnapshot{
		Symbol:           symbol,
		TS:               time.Now(),
		ConsensusRatePct: f.rate,
		Confidence:       f.conf,
		Reason:           "consensus_pass",
	}, nil
}

func TestRunOnceWritesLabel(t *testing.T) {
	tmp := t.TempDir()
	paper := filepath.Join(tmp, "paper.jsonl")
	labels := filepath.Join(tmp, "labels.jsonl")

	// Single repo event at ts=now-20s so 10s window is due.
	now := time.Date(2026, 1, 29, 0, 0, 20, 0, time.UTC)
	paperTS := now.Add(-20 * time.Second).Format(time.RFC3339)
	content := `{"ts":"` + paperTS + `","event":{"source":"cn_repo_realtime_action","trade_date":"20260129","market":"CN-A","symbol":"204001.SH","title":"demo","body":"","tags":{"tier":"action","kind":"repo"},"data":{"consensus_rate_pct":5.0}}}` + "\n"
	if err := os.WriteFile(paper, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Engine: config.EngineConfig{TradeDateMode: "fixed", FixedTradeDate: "20260129"},
		Signals: []config.SignalConfig{
			{Type: "cn_repo_realtime", Name: "cn_repo_realtime_action", Enabled: true, MinYieldPct: 4.0},
		},
	}

	lcfg := DefaultConfig()
	lcfg.Windows = []time.Duration{10 * time.Second}
	lcfg.Grace = 60 * time.Second
	lcfg.MaxPerRun = 10
	lcfg.Now = func() time.Time { return now }

	r := New(cfg, fakeFusion{rate: 4.5, conf: marketdata.ConfidencePass}, lcfg)
	wrote, _, err := r.RunOnce(context.Background(), paper, labels)
	if err != nil {
		t.Fatal(err)
	}
	if wrote != 1 {
		t.Fatalf("wrote=%d", wrote)
	}
	b, _ := os.ReadFile(labels)
	s := string(b)
	if !strings.Contains(s, "\"reward\":1") {
		t.Fatalf("expected reward=1, got: %s", s)
	}
}

