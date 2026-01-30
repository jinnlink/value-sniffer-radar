package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParseEnrichment(s string) (Enrichment, error) {
	var e Enrichment
	s = strings.TrimSpace(s)
	// Best-effort: strip leading BOM and code fences.
	s = strings.TrimPrefix(s, "\ufeff")
	s = strings.TrimSpace(strings.TrimPrefix(s, "```json"))
	s = strings.TrimSpace(strings.TrimPrefix(s, "```"))
	s = strings.TrimSpace(strings.TrimSuffix(s, "```"))
	if s == "" {
		return e, fmt.Errorf("empty output")
	}
	if err := json.Unmarshal([]byte(s), &e); err != nil {
		return e, err
	}
	e.Summary = strings.TrimSpace(e.Summary)
	if e.Summary == "" {
		return e, fmt.Errorf("missing summary")
	}
	// Normalize.
	e.Risks = compact(e.Risks)
	e.Checklist = compact(e.Checklist)
	return e, nil
}

func compact(in []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, s := range in {
		t := strings.TrimSpace(s)
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if !seen[key] {
			seen[key] = true
			out = append(out, t)
		}
	}
	return out
}
