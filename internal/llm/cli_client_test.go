package llm

import (
	"context"
	"testing"
	"time"
)

func TestCLIClientComplete(t *testing.T) {
	// Use powershell.exe to output a fixed JSON blob (avoid cmd.exe escaping quirks).
	c := CLIClient{
		Command: "powershell.exe",
		Args:    []string{"-NoProfile", "-Command", "Write-Output '{\"summary\":\"ok\",\"risks\":[],\"checklist\":[]}'"},
		Timeout: 5 * time.Second,
	}
	out, err := c.Complete(context.Background(), "ignored")
	if err != nil {
		t.Fatalf("Complete err=%v", err)
	}
	_, err = ParseEnrichment(out)
	if err != nil {
		t.Fatalf("ParseEnrichment err=%v out=%q", err, out)
	}
}
