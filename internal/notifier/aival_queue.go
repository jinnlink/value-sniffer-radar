package notifier

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"value-sniffer-radar/internal/config"
)

// AivalQueue writes AI-Value (aival.event.v1) JSON files into a queue dir.
// The AstrBot plugin ai_value can poll this directory and push messages to QQ.
type AivalQueue struct {
	queueDir string
	market   string
	tags     []string
}

func NewAivalQueue(c config.NotifierConfig) (*AivalQueue, error) {
	if strings.TrimSpace(c.QueueDir) == "" {
		return nil, fmt.Errorf("aival_queue.queue_dir required")
	}
	market := strings.TrimSpace(c.Market)
	if market == "" {
		market = "CN-A"
	}
	var tags []string
	for _, t := range c.Tags {
		if s := strings.TrimSpace(t); s != "" {
			tags = append(tags, s)
		}
	}
	return &AivalQueue{
		queueDir: c.QueueDir,
		market:   market,
		tags:     tags,
	}, nil
}

func (q *AivalQueue) Name() string { return "aival_queue" }

func (q *AivalQueue) Notify(_ context.Context, events []Event) error {
	if err := os.MkdirAll(q.queueDir, 0o755); err != nil {
		return err
	}
	for _, e := range events {
		if err := q.dropOne(e); err != nil {
			return err
		}
	}
	return nil
}

func (q *AivalQueue) dropOne(e Event) error {
	now := time.Now()
	id := newID("evt", now)

	title := strings.TrimSpace(e.Title)
	if title == "" {
		title = "ValueSniffer alert"
	}

	text := strings.TrimSpace(e.Body)
	var header []string
	if s := strings.TrimSpace(e.Source); s != "" {
		header = append(header, "signal: "+s)
	}
	if s := strings.TrimSpace(e.TradeDate); s != "" {
		header = append(header, "trade_date: "+s)
	}
	if len(header) > 0 {
		if text != "" {
			text = strings.Join(header, "\n") + "\n" + text
		} else {
			text = strings.Join(header, "\n")
		}
	}

	metrics := map[string]any{}
	if e.TradeDate != "" {
		metrics["trade_date"] = e.TradeDate
	}
	if e.Source != "" {
		metrics["signal"] = e.Source
	}
	for k, v := range e.Data {
		metrics[k] = v
	}
	if len(metrics) == 0 {
		metrics = nil
	}

	tags := append([]string{}, q.tags...)
	if len(e.Tags) > 0 {
		keys := make([]string, 0, len(e.Tags))
		for k := range e.Tags {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := strings.TrimSpace(e.Tags[k])
			if v == "" {
				continue
			}
			tags = append(tags, fmt.Sprintf("%s=%s", k, v))
		}
	}
	if len(tags) == 0 {
		tags = nil
	}

	payload := map[string]any{
		"schema": "aival.event.v1",
		"kind":   "event",
		"id":     id,
		"ts":     now.Format(time.RFC3339),
		"title":  title,
		"market": q.market,
		"text":   text,
		"source": map[string]any{
			"app":    "value-sniffer-radar",
			"signal": e.Source,
		},
	}
	if e.Market != "" {
		payload["market"] = e.Market
	}
	if e.Symbol != "" {
		payload["symbol"] = e.Symbol
	}
	if metrics != nil {
		payload["metrics"] = metrics
	}
	if tags != nil {
		payload["tags"] = tags
	}

	filename := fmt.Sprintf("%d_%s.json", now.UnixMilli(), id)
	outPath := filepath.Join(q.queueDir, filename)
	tmpPath := outPath + ".tmp"

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmpPath, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, outPath)
}

func newID(prefix string, now time.Time) string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	suffix := hex.EncodeToString(b[:])
	base := fmt.Sprintf("%d_%s", now.UnixMilli(), suffix)
	p := strings.TrimSpace(prefix)
	if p == "" {
		return base
	}
	return p + "_" + base
}
