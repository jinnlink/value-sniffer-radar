package optimizer

import (
	"fmt"
	"strings"
	"time"
)

type Report struct {
	GeneratedAt time.Time
	InputPath   string
	LabelsPath  string
	// PrimaryWindowSec is the label window used for bandit updates (0 means labels not used).
	PrimaryWindowSec int

	Warnings []string

	UniqueEvents      int
	RewardsUsed       int
	RewardsFromLabels int
	RewardsFromPaper  int

	LabelWarnings []string
	Coverage      []CoverageStat
	RewardRates   []RewardRateStat

	ArmsTotal int
	Alloc     []Allocation
}

type CoverageStat struct {
	WindowSec     int
	TotalEvents   int
	LabeledEvents int
	CoveragePct   float64
}

type RewardRateStat struct {
	Signal    string
	WindowSec int
	N         int
	RewardSum int
	RatePct   float64
}

func RenderMarkdown(r Report) string {
	var b strings.Builder
	b.WriteString("# Optimizer Report\n\n")
	b.WriteString(fmt.Sprintf("- generated_at: `%s`\n", r.GeneratedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("- input: `%s`\n", r.InputPath))
	if r.LabelsPath != "" {
		b.WriteString(fmt.Sprintf("- labels: `%s`\n", r.LabelsPath))
		if r.PrimaryWindowSec > 0 {
			b.WriteString(fmt.Sprintf("- labels_primary_window_sec: `%d`\n", r.PrimaryWindowSec))
		}
	}
	b.WriteString(fmt.Sprintf("- arms_total: `%d`\n", r.ArmsTotal))
	if r.UniqueEvents > 0 {
		b.WriteString(fmt.Sprintf("- unique_events: `%d`\n", r.UniqueEvents))
	}
	if r.RewardsUsed > 0 {
		b.WriteString(fmt.Sprintf("- rewards_used: `%d` (labels=%d, paper=%d)\n", r.RewardsUsed, r.RewardsFromLabels, r.RewardsFromPaper))
	}
	b.WriteString("\n")

	if len(r.Warnings) > 0 || len(r.LabelWarnings) > 0 {
		b.WriteString("## Warnings\n")
		for _, w := range r.Warnings {
			b.WriteString("- ")
			b.WriteString(w)
			b.WriteString("\n")
		}
		for _, w := range r.LabelWarnings {
			b.WriteString("- ")
			b.WriteString(w)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if r.LabelsPath != "" && len(r.Coverage) > 0 {
		b.WriteString("## Labeled Coverage (by window)\n")
		b.WriteString("| window_sec | labeled_events | total_events | coverage |\n|---:|---:|---:|---:|\n")
		for _, c := range r.Coverage {
			b.WriteString(fmt.Sprintf("| %d | %d | %d | %.2f%% |\n", c.WindowSec, c.LabeledEvents, c.TotalEvents, c.CoveragePct))
		}
		b.WriteString("\n")
	}

	if r.LabelsPath != "" && len(r.RewardRates) > 0 {
		b.WriteString("## Reward Rate (by signal/window)\n")
		b.WriteString("| signal | window_sec | reward_rate | n | reward_sum |\n|---|---:|---:|---:|---:|\n")
		for _, rr := range r.RewardRates {
			b.WriteString(fmt.Sprintf("| %s | %d | %.2f%% | %d | %d |\n", rr.Signal, rr.WindowSec, rr.RatePct, rr.N, rr.RewardSum))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Suggested Action Allocation (Thompson Sampling)\n")
	if len(r.Alloc) == 0 {
		b.WriteString("_(none)_\n")
		return b.String()
	}
	b.WriteString("| key | sample_score | mean | n |\n|---|---:|---:|---:|\n")
	for _, a := range r.Alloc {
		b.WriteString(fmt.Sprintf("| %s | %.4f | %.4f | %d |\n", a.Key, a.Score, a.Mean, a.N))
	}
	b.WriteString("\n")
	b.WriteString("## Notes\n")
	b.WriteString("- This tool consumes `paper_log` JSONL.\n")
	b.WriteString("- If `-labels` is provided, it prefers `labels.repo.jsonl` rewards (per `labels_primary_window_sec`) and falls back to `event.data.reward` when missing.\n")
	b.WriteString("- Use `-out-reco` to emit a machine-readable daily quota suggestion file for runtime consumption.\n")
	return b.String()
}
