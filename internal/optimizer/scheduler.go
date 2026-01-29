package optimizer

import (
	"math"
	"sort"
	"time"
)

type PollItem struct {
	Symbol        string
	ProviderScore float64 // [0,1]
	ConflictRate  float64 // [0,1]
	WindowWeight  float64 // [0,1] time-of-day prior (e.g. repo close window)
}

type PollPlanItem struct {
	Symbol   string
	Priority float64
	NextIn   time.Duration
}

type SchedulerConfig struct {
	BudgetPerMinute int
	MinInterval     time.Duration
	MaxInterval     time.Duration
}

// BuildPollPlan returns a deterministic plan sorted by priority.
// The idea: poll more frequently where (expected_value) is high:
// - higher providerScore
// - higher windowWeight
// - lower conflictRate (more stable)
func BuildPollPlan(items []PollItem, cfg SchedulerConfig) []PollPlanItem {
	if cfg.BudgetPerMinute <= 0 {
		cfg.BudgetPerMinute = 60
	}
	if cfg.MinInterval <= 0 {
		cfg.MinInterval = 1 * time.Second
	}
	if cfg.MaxInterval <= 0 {
		cfg.MaxInterval = 15 * time.Second
	}

	score := func(it PollItem) float64 {
		ps := clamp01(it.ProviderScore)
		cf := clamp01(it.ConflictRate)
		ww := clamp01(it.WindowWeight)
		// Penalize conflict more aggressively; avoid polling noisy symbols too hard.
		return ps*0.45 + ww*0.45 + (1-cf)*0.10
	}

	out := make([]PollPlanItem, 0, len(items))
	for _, it := range items {
		p := score(it)
		// Map priority->interval: higher p => closer to MinInterval.
		next := lerpDuration(cfg.MaxInterval, cfg.MinInterval, p)
		out = append(out, PollPlanItem{
			Symbol:   it.Symbol,
			Priority: p,
			NextIn:   next,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].Symbol < out[j].Symbol
		}
		return out[i].Priority > out[j].Priority
	})
	return out
}

func clamp01(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0
	}
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func lerpDuration(a, b time.Duration, t float64) time.Duration {
	t = clamp01(t)
	return time.Duration(float64(a) + (float64(b)-float64(a))*t)
}

