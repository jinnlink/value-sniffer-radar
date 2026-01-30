package engine

import (
	"path/filepath"
	"testing"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/reco"
)

func TestEngineLoadRecoOverridesPerSignalQuotas(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "optimizer.reco.json")
	if err := reco.Write(path, reco.Recommendation{
		Version:          "reco.v1",
		InputPaper:       "paper.jsonl",
		InputLabels:      "labels.repo.jsonl",
		PrimaryWindowSec: 30,
		Slots:            30,
		Quotas: []reco.SignalQuota{
			{Signal: "sigA", MeanReward: 0.6, N: 10, SuggestedDailyQuota: 2},
		},
	}); err != nil {
		t.Fatalf("write reco err=%v", err)
	}

	e := &Engine{
		cfg: &config.Config{
			Engine: config.EngineConfig{
				RecoPath: path,
				ActionMaxEventsPerSignalPerDay: map[string]int{
					"sigA": 10,
				},
			},
		},
		dailySent: map[string]int{},
	}
	e.loadRecoIfConfigured()
	if e.recoQuotas == nil || e.recoQuotas["sigA"] != 2 {
		t.Fatalf("expected recoQuotas override, got=%v", e.recoQuotas)
	}
}
