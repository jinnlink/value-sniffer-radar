package signals

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/marketdata"
	"value-sniffer-radar/internal/notifier"
	"value-sniffer-radar/internal/tushare"
)

// CNRepoRealtime alerts on reverse repo yield spikes using realtime marketdata fusion.
// It will only emit tier=action when fusion confidence passes (multi-source consensus).
type CNRepoRealtime struct {
	name        string
	tier        string
	minInterval time.Duration

	repoCodes   []string
	minYieldPct float64
	topN        int

	windowStart string
	windowEnd   string

	confirmK int
	streaks  map[string]int
}

func NewCNRepoRealtime(c config.SignalConfig) *CNRepoRealtime {
	name := c.Name
	if name == "" {
		name = "cn_repo_realtime"
	}
	tier := c.Tier
	if tier == "" {
		tier = "action"
	}
	topN := c.TopN
	if topN <= 0 {
		topN = 5
	}
	minYield := c.MinYieldPct
	if minYield <= 0 {
		minYield = 4.0
	}
	repoCodes := normalizeRepoCodes(c.RepoCodes)
	if len(repoCodes) == 0 {
		repoCodes = []string{"204001.SH", "131810.SZ"}
	}
	minInt := time.Duration(c.MinIntervalSeconds) * time.Second
	if minInt <= 0 {
		minInt = 3 * time.Second
	}
	confirmK := c.ConfirmK
	if confirmK <= 0 {
		confirmK = 1
	}

	return &CNRepoRealtime{
		name:        name,
		tier:        tier,
		minInterval: minInt,
		repoCodes:   repoCodes,
		minYieldPct: minYield,
		topN:        topN,
		windowStart: strings.TrimSpace(c.WindowStart),
		windowEnd:   strings.TrimSpace(c.WindowEnd),
		confirmK:    confirmK,
		streaks:     map[string]int{},
	}
}

func (s *CNRepoRealtime) Name() string { return s.name }

func (s *CNRepoRealtime) MinInterval() time.Duration { return s.minInterval }

type repoRTAlert struct {
	tsCode    string
	ratePct   float64
	conf      marketdata.Confidence
	reason    string
	providers []marketdata.ProviderResult
}

func (s *CNRepoRealtime) Evaluate(ctx context.Context, _ *tushare.Client, tradeDate string, md marketdata.Fusion) ([]notifier.Event, error) {
	if md == nil {
		return nil, fmt.Errorf("marketdata disabled: enable config.marketdata and providers for %s", s.name)
	}
	if !withinWindow(time.Now(), s.windowStart, s.windowEnd) {
		return nil, nil
	}

	var alerts []repoRTAlert
	for _, code := range s.repoCodes {
		fs, err := md.FetchFusion(ctx, code)
		if err != nil {
			continue
		}

		pass := fs.Confidence == marketdata.ConfidencePass
		thr := fs.ConsensusRatePct >= s.minYieldPct
		if pass && thr {
			s.streaks[code]++
		} else {
			s.streaks[code] = 0
		}

		if !pass || !thr {
			continue
		}
		if s.streaks[code] < s.confirmK {
			continue
		}

		alerts = append(alerts, repoRTAlert{
			tsCode:    code,
			ratePct:   fs.ConsensusRatePct,
			conf:      fs.Confidence,
			reason:    fs.Reason,
			providers: fs.Providers,
		})
	}

	if len(alerts) == 0 {
		return nil, nil
	}

	sort.Slice(alerts, func(i, j int) bool { return alerts[i].ratePct > alerts[j].ratePct })
	if len(alerts) > s.topN {
		alerts = alerts[:s.topN]
	}

	events := make([]notifier.Event, 0, len(alerts))
	for _, a := range alerts {
		body := fmt.Sprintf("consensus_rate=%.4f%%\nconfidence=%s\nreason=%s\n", a.ratePct, a.conf, a.reason)
		for _, pr := range a.providers {
			if pr.Error != "" {
				body += fmt.Sprintf("- %s: error=%s\n", pr.Provider, pr.Error)
				continue
			}
			body += fmt.Sprintf("- %s: rate=%.4f%% inlier=%v outlier=%v\n", pr.Provider, pr.Snapshot.RatePct, pr.Inlier, pr.Outlier)
		}

		events = append(events, notifier.Event{
			Source:    s.name,
			TradeDate: tradeDate,
			Market:    "CN-A",
			Symbol:    a.tsCode,
			Title:     fmt.Sprintf("Repo realtime %.2f%% (%s)", a.ratePct, a.tsCode),
			Body:      body,
			Tags: map[string]string{
				"kind":       "repo",
				"strategy":   "yield_spike",
				"tier":       s.tier,
				"confidence": string(a.conf),
			},
			Data: map[string]any{
				"consensus_rate_pct":  a.ratePct,
				"threshold_yield_pct": s.minYieldPct,
				"expected_edge_pct":   a.ratePct - s.minYieldPct,
				"confidence":          string(a.conf),
				"reason":              a.reason,
				"providers":           a.providers,
			},
		})
	}
	return events, nil
}
