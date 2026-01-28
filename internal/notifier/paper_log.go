package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"value-sniffer-radar/internal/config"
)

// PaperLog appends every event as one JSON line (JSONL), for later evaluation/backtest.
type PaperLog struct {
	path string
}

func NewPaperLog(c config.NotifierConfig) (*PaperLog, error) {
	p := c.FilePath
	if p == "" {
		p = filepath.Join("state", "paper.jsonl")
	}
	return &PaperLog{path: p}, nil
}

func (p *PaperLog) Name() string { return "paper_log" }

func (p *PaperLog) Notify(_ context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p.path), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(p.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	now := time.Now().Format(time.RFC3339)
	for _, e := range events {
		rec := map[string]any{
			"ts":    now,
			"event": e,
		}
		b, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func (p *PaperLog) String() string { return fmt.Sprintf("paper_log(%s)", p.path) }
