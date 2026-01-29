package optimizer

import (
	"fmt"
	"strings"
	"time"
)

type Report struct {
	GeneratedAt time.Time
	InputPath   string
	Warnings    []string

	ArmsTotal int
	Alloc     []Allocation
}

func RenderMarkdown(r Report) string {
	var b strings.Builder
	b.WriteString("# Optimizer Report\n\n")
	b.WriteString(fmt.Sprintf("- generated_at: `%s`\n", r.GeneratedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("- input: `%s`\n", r.InputPath))
	b.WriteString(fmt.Sprintf("- arms_total: `%d`\n", r.ArmsTotal))
	b.WriteString("\n")

	if len(r.Warnings) > 0 {
		b.WriteString("## Warnings\n")
		for _, w := range r.Warnings {
			b.WriteString("- ")
			b.WriteString(w)
			b.WriteString("\n")
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
	b.WriteString("- This tool consumes `paper_log` JSONL. For real PnL, add a later ticket to label events with `reward` based on price windows after costs.\n")
	return b.String()
}

