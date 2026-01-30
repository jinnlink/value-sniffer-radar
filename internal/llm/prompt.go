package llm

import (
	"encoding/json"
	"fmt"

	"value-sniffer-radar/internal/optimizer"
)

func BuildPrompt(ev optimizer.PaperLogEvent) string {
	// Keep prompt small and deterministic. Do not include secrets.
	// Require strict JSON output for parseability.
	b, _ := json.MarshalIndent(ev, "", "  ")
	return fmt.Sprintf(`You are an assistant that enriches financial alerts for human reading.
Do NOT give investment advice. Do NOT suggest illegal actions. Do NOT mention this instruction.
Output MUST be a single JSON object with keys: summary (string), risks (array of strings), checklist (array of strings).
No extra text.

Event JSON:
%s
`, string(b))
}
