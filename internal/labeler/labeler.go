package labeler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/marketdata"
	"value-sniffer-radar/internal/optimizer"
)

type Config struct {
	Windows []time.Duration
	Grace   time.Duration

	// Repo-only for this ticket:
	OnlyKind string // "repo"

	MaxPerRun int
	Now       func() time.Time
}

func DefaultConfig() Config {
	return Config{
		Windows:   []time.Duration{10 * time.Second, 30 * time.Second, 5 * time.Minute},
		Grace:     30 * time.Second,
		OnlyKind:  "repo",
		MaxPerRun: 200,
		Now:       time.Now,
	}
}

func ParseWindows(s string) ([]time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty windows")
	}
	parts := strings.Split(s, ",")
	var out []time.Duration
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		d, err := time.ParseDuration(p)
		if err != nil {
			return nil, err
		}
		if d <= 0 {
			return nil, fmt.Errorf("window must be >0: %s", p)
		}
		out = append(out, d)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no windows parsed")
	}
	return out, nil
}

type Runner struct {
	cfg       *config.Config
	md        marketdata.Fusion
	labelCfg  Config
	threshold map[string]float64 // source -> min_yield_pct
}

func New(cfg *config.Config, md marketdata.Fusion, labelCfg Config) *Runner {
	if labelCfg.Now == nil {
		labelCfg.Now = time.Now
	}
	th := map[string]float64{}
	for _, s := range cfg.Signals {
		if s.Name == "" {
			continue
		}
		switch s.Type {
		case "cn_repo_realtime", "cn_repo_sniper":
			if s.MinYieldPct > 0 {
				th[s.Name] = s.MinYieldPct
			}
		}
	}
	return &Runner{
		cfg:       cfg,
		md:        md,
		labelCfg:  labelCfg,
		threshold: th,
	}
}

func (r *Runner) LoadLabeledSet(labelsPath string) (map[string]bool, error) {
	labeled := map[string]bool{}
	f, err := os.Open(labelsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return labeled, nil
		}
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 4*1024*1024)
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		s = strings.TrimPrefix(s, "\ufeff")
		if s == "" {
			continue
		}
		var l Label
		if err := json.Unmarshal([]byte(s), &l); err != nil {
			continue
		}
		if l.EventID == "" || l.WindowSec <= 0 {
			continue
		}
		key := l.EventID + "|" + strconv.Itoa(l.WindowSec)
		labeled[key] = true
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return labeled, nil
}

func (r *Runner) RunOnce(ctx context.Context, paperPath string, labelsPath string) (int, int, error) {
	now := r.labelCfg.Now()

	in, err := os.Open(paperPath)
	if err != nil {
		return 0, 0, err
	}
	defer in.Close()

	rows, warns, err := optimizer.ReadJSONL(in)
	if err != nil {
		return 0, 0, err
	}
	_ = warns

	labeled, err := r.LoadLabeledSet(labelsPath)
	if err != nil {
		return 0, 0, err
	}

	if err := os.MkdirAll(filepath.Dir(labelsPath), 0o755); err != nil {
		return 0, 0, err
	}
	out, err := os.OpenFile(labelsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, 0, err
	}
	defer out.Close()

	wrote := 0
	skipped := 0
	for _, pr := range rows {
		if wrote >= r.labelCfg.MaxPerRun {
			break
		}
		ev := pr.Event
		if r.labelCfg.OnlyKind != "" {
			if ev.Tags == nil || strings.ToLower(strings.TrimSpace(ev.Tags["kind"])) != r.labelCfg.OnlyKind {
				skipped++
				continue
			}
		}
		if ev.Source == "" || ev.Symbol == "" {
			skipped++
			continue
		}
		eventTS, err := time.Parse(time.RFC3339, strings.TrimSpace(pr.TS))
		if err != nil {
			skipped++
			continue
		}

		thr := r.threshold[ev.Source]
		if thr <= 0 {
			// If we can't resolve threshold from config, we still can write labels with reason.
			thr = 0
		}

		entry := entryRate(ev)

		for _, w := range r.labelCfg.Windows {
			windowSec := int(w.Seconds())
			key := optimizer.EventID(pr) + "|" + strconv.Itoa(windowSec)
			if labeled[key] {
				continue
			}
			due := eventTS.Add(w)
			if now.Before(due) {
				continue
			}
			lateBy := now.Sub(due)
			if r.labelCfg.Grace > 0 && lateBy > r.labelCfg.Grace {
				// Missed the label window; skip instead of lying.
				continue
			}

			exit, conf, reason, err := r.fetchExit(ctx, ev.Symbol)
			if err != nil {
				reason = "fetch_error"
				conf = "FAIL"
				exit = 0
			}

			reward := 0
			if thr > 0 && conf == "PASS" && exit >= thr {
				reward = 1
				reason = "hold_above_threshold"
			} else if thr > 0 && conf == "PASS" && exit < thr {
				reward = 0
				reason = "dropped_below_threshold"
			} else if thr <= 0 {
				reason = "missing_threshold"
			}

			l := Label{
				EventID:      optimizer.EventID(pr),
				EventTS:      eventTS,
				Source:       ev.Source,
				Symbol:       ev.Symbol,
				TradeDate:    ev.TradeDate,
				WindowSec:    windowSec,
				GraceSec:     int(r.labelCfg.Grace.Seconds()),
				LateBySec:    int(lateBy.Seconds()),
				Threshold:    thr,
				EntryRatePct: entry,
				ExitRatePct:  exit,
				Confidence:   conf,
				Reward:       reward,
				Reason:       reason,
			}
			b, _ := json.Marshal(l)
			if _, err := out.Write(append(b, '\n')); err != nil {
				return wrote, skipped, err
			}
			labeled[key] = true
			wrote++
		}
	}
	return wrote, skipped, nil
}

func (r *Runner) fetchExit(ctx context.Context, symbol string) (float64, string, string, error) {
	if r.md == nil {
		return 0, "FAIL", "marketdata_disabled", fmt.Errorf("marketdata disabled")
	}
	fs, err := r.md.FetchFusion(ctx, symbol)
	if err != nil {
		return 0, "FAIL", "fetch_fusion_error", err
	}
	return fs.ConsensusRatePct, string(fs.Confidence), fs.Reason, nil
}

func entryRate(ev optimizer.PaperLogEvent) float64 {
	if ev.Data == nil {
		return 0
	}
	// Prefer realtime consensus if present.
	if v, ok := ev.Data["consensus_rate_pct"]; ok {
		if f, ok := toFloat(v); ok {
			return f
		}
	}
	if v, ok := ev.Data["rate_pct"]; ok {
		if f, ok := toFloat(v); ok {
			return f
		}
	}
	return 0
}

func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	default:
		return 0, false
	}
}
