package reco

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SignalQuota struct {
	Signal              string  `json:"signal"`
	MeanReward          float64 `json:"mean_reward"`
	N                   int     `json:"n"`
	SuggestedDailyQuota int     `json:"suggested_daily_quota"`
}

type Recommendation struct {
	Version          string        `json:"version"`
	GeneratedAt      time.Time     `json:"generated_at"`
	InputPaper       string        `json:"input_paper"`
	InputLabels      string        `json:"input_labels,omitempty"`
	PrimaryWindowSec int           `json:"primary_window_sec"`
	Slots            int           `json:"slots"`
	Quotas           []SignalQuota `json:"quotas"`
}

func Write(path string, r Recommendation) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func Read(path string) (Recommendation, error) {
	var r Recommendation
	if path == "" {
		return r, fmt.Errorf("empty path")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return r, err
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return r, err
	}
	return r, nil
}
