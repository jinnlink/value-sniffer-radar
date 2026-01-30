package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"value-sniffer-radar/internal/llm"
	"value-sniffer-radar/internal/optimizer"
)

func main() {
	var inPath string
	var outPath string
	var mode string
	var provider string
	var maxEvents int

	// api provider
	var apiBaseURL string
	var apiModel string
	var apiKeyEnv string

	// cli provider
	var cliCmd string
	var cliArgs string

	flag.StringVar(&inPath, "in", "", "Input paper_log JSONL path")
	flag.StringVar(&outPath, "out", "", "Output path (enriched.jsonl or digest.md)")
	flag.StringVar(&mode, "mode", "enrich", "Mode: enrich|digest")
	flag.StringVar(&provider, "provider", "cli", "Provider: api|cli")
	flag.IntVar(&maxEvents, "max-events", 50, "How many latest events to process")

	flag.StringVar(&apiBaseURL, "api-base-url", "", "API base URL (OpenAI-compatible). Default https://api.openai.com/v1")
	flag.StringVar(&apiModel, "api-model", "", "API model name")
	flag.StringVar(&apiKeyEnv, "api-key-env", "LLM_API_KEY", "Env var that stores API key")

	flag.StringVar(&cliCmd, "cli-cmd", "", "CLI command (e.g. codex.cmd, gemini.cmd)")
	flag.StringVar(&cliArgs, "cli-args", "", "CLI args string (space-separated)")
	flag.Parse()

	if inPath == "" {
		fmt.Fprintln(os.Stderr, "[error] missing -in")
		os.Exit(2)
	}
	if outPath == "" {
		if strings.EqualFold(strings.TrimSpace(mode), "digest") {
			outPath = ".\\state\\llm.digest.md"
		} else {
			outPath = ".\\state\\llm.enriched.jsonl"
		}
	}

	rows, warns, err := readPaper(inPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error] read:", err.Error())
		os.Exit(1)
	}
	if len(warns) > 0 {
		fmt.Fprintf(os.Stderr, "[warn] paper: %s\n", strings.Join(warns, ","))
	}
	if maxEvents > 0 && len(rows) > maxEvents {
		rows = rows[len(rows)-maxEvents:]
	}

	c, err := buildClient(provider, apiBaseURL, apiModel, apiKeyEnv, cliCmd, cliArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error] client:", err.Error())
		os.Exit(2)
	}

	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "enrich":
		if err := runEnrich(rows, c, outPath); err != nil {
			fmt.Fprintln(os.Stderr, "[error] enrich:", err.Error())
			os.Exit(1)
		}
	case "digest":
		if err := runDigest(rows, c, outPath); err != nil {
			fmt.Fprintln(os.Stderr, "[error] digest:", err.Error())
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "[error] unknown -mode:", mode)
		os.Exit(2)
	}
}

func readPaper(path string) ([]optimizer.PaperRow, []string, error) {
	if path == "-" {
		return optimizer.ReadJSONL(os.Stdin)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return optimizer.ReadJSONL(f)
}

func buildClient(provider, baseURL, model, keyEnv, cliCmd, cliArgs string) (llm.Client, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "api":
		key := os.Getenv(strings.TrimSpace(keyEnv))
		return llm.APIClient{
			BaseURL: baseURL,
			APIKey:  key,
			Model:   model,
			Timeout: 30 * time.Second,
		}, nil
	case "cli":
		if strings.TrimSpace(cliCmd) == "" {
			return nil, fmt.Errorf("missing -cli-cmd for provider=cli")
		}
		args := splitArgs(cliArgs)
		return llm.CLIClient{
			Command: strings.TrimSpace(cliCmd),
			Args:    args,
			Timeout: 60 * time.Second,
		}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func splitArgs(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// Simple split; for complex quoting, pass a wrapper cmd.
	return strings.Fields(s)
}

func runEnrich(rows []optimizer.PaperRow, c llm.Client, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	ctx := context.Background()
	processed := 0
	skipped := 0
	for _, pr := range rows {
		prompt := llm.BuildPrompt(pr.Event)
		raw, err := c.Complete(ctx, prompt)
		if err != nil {
			skipped++
			continue
		}
		en, err := llm.ParseEnrichment(raw)
		if err != nil {
			skipped++
			continue
		}
		rec := map[string]any{
			"ts":       pr.TS,
			"event":    pr.Event,
			"llm":      en,
			"llm_v":    "llm.enrich.v1",
			"provider": c.Name(),
		}
		b, _ := json.Marshal(rec)
		if _, err := out.Write(append(b, '\n')); err != nil {
			return err
		}
		processed++
	}
	fmt.Printf("enriched_written=%d skipped=%d out=%s provider=%s\n", processed, skipped, outPath, c.Name())
	return nil
}

func runDigest(rows []optimizer.PaperRow, c llm.Client, outPath string) error {
	// Minimal digest: ask LLM to summarize the last N action events.
	var items []optimizer.PaperLogEvent
	for _, pr := range rows {
		items = append(items, pr.Event)
	}
	b, _ := json.MarshalIndent(items, "", "  ")
	prompt := fmt.Sprintf(`Summarize these events into a short digest for a human.
Do NOT give investment advice. Output MUST be Markdown.

Events JSON:
%s
`, string(b))

	ctx := context.Background()
	md, err := c.Complete(ctx, prompt)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(md), 0o644)
}
