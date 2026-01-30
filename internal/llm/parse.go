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
		// Many CLIs may print extra text. Best-effort: scan for a JSON object and decode it.
		if ee, ok := tryDecodeFirstJSONObject(s); ok {
			e = ee
		} else {
			return e, err
		}
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

func tryDecodeFirstJSONObject(s string) (Enrichment, bool) {
	var e Enrichment
	for i := 0; i < len(s); i++ {
		if s[i] != '{' {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(s[i:]))
		if err := dec.Decode(&e); err != nil {
			continue
		}
		return e, true
	}
	return Enrichment{}, false
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
