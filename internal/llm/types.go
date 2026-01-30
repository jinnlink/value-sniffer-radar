package llm

type Enrichment struct {
	Summary   string   `json:"summary"`
	Risks     []string `json:"risks"`
	Checklist []string `json:"checklist"`
}
