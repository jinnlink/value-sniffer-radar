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

// CNRepoSniper alerts on CN reverse repo "yield" spikes.
//
// NOTE:
// - This is not cross-market arbitrage. Treat as cash-management / mean-reversion opportunity.
// - We use Tushare `repo_daily` weighted price (`weight`) as the "rate" (%).
type CNRepoSniper struct {
	name        string
	tier        string
	minInterval time.Duration

	repoCodes   []string
	minYieldPct float64
	minAmount   float64
	topN        int

	windowStart string
	windowEnd   string
}

func NewCNRepoSniper(c config.SignalConfig) *CNRepoSniper {
	name := c.Name
	if name == "" {
		name = "cn_repo_sniper"
	}
	tier := c.Tier
	if tier == "" {
		tier = "action"
	}
	topN := c.TopN
	if topN <= 0 {
		topN = 10
	}
	minYield := c.MinYieldPct
	if minYield <= 0 {
		// Default threshold is deliberately conservative: only alert when rates look "interesting".
		minYield = 4.0
	}
	repoCodes := normalizeRepoCodes(c.RepoCodes)
	if len(repoCodes) == 0 {
		// Commonly traded baseline repos:
		// - 204001: SH 1-day reverse repo (GC001 in colloquial naming)
		// - 131810: SZ 1-day reverse repo (R-001)
		repoCodes = []string{"204001.SH", "131810.SZ"}
	}
	minInt := time.Duration(c.MinIntervalSeconds) * time.Second
	return &CNRepoSniper{
		name:        name,
		tier:        tier,
		minInterval: minInt,
		repoCodes:   repoCodes,
		minYieldPct: minYield,
		minAmount:   c.MinAmount,
		topN:        topN,
		windowStart: strings.TrimSpace(c.WindowStart),
		windowEnd:   strings.TrimSpace(c.WindowEnd),
	}
}

func (s *CNRepoSniper) Name() string { return s.name }

func (s *CNRepoSniper) MinInterval() time.Duration { return s.minInterval }

type repoAlert struct {
	tsCode     string
	tradeDate  string
	ratePct    float64
	close      float64
	weight     float64
	amount     float64
	vol        float64
	avgAmt     float64
	amountHint string
}

func (s *CNRepoSniper) Evaluate(ctx context.Context, client *tushare.Client, tradeDate string, _ marketdata.Fusion) ([]notifier.Event, error) {
	if !withinWindow(time.Now(), s.windowStart, s.windowEnd) {
		return nil, nil
	}

	var alerts []repoAlert
	for _, code := range s.repoCodes {
		rows, err := client.Query(ctx, "repo_daily", map[string]any{
			"ts_code":    code,
			"trade_date": tradeDate,
		}, []string{"ts_code", "trade_date", "close", "amount", "vol", "avg_amt", "weight"})
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			tsCode := tushare.GetString(r, "ts_code")
			if tsCode == "" {
				tsCode = code
			}
			weight := tushare.GetFloat(r, "weight")
			closeP := tushare.GetFloat(r, "close")
			rate := weight
			if rate <= 0 {
				rate = closeP
			}
			amount := tushare.GetFloat(r, "amount")
			if s.minAmount > 0 && amount > 0 && amount < s.minAmount {
				continue
			}
			if rate < s.minYieldPct {
				continue
			}
			alerts = append(alerts, repoAlert{
				tsCode:    tsCode,
				tradeDate: tushare.GetString(r, "trade_date"),
				ratePct:   rate,
				close:     closeP,
				weight:    weight,
				amount:    amount,
				vol:       tushare.GetFloat(r, "vol"),
				avgAmt:    tushare.GetFloat(r, "avg_amt"),
			})
		}
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
		td := tradeDate
		if a.tradeDate != "" {
			td = a.tradeDate
		}
		body := fmt.Sprintf(
			"rate=%.4f%%\nts_code=%s\nclose=%.4f\nweight=%.4f\namount=%.0f\nvol=%.0f\navg_amt=%.0f\n",
			a.ratePct, a.tsCode, a.close, a.weight, a.amount, a.vol, a.avgAmt,
		)
		events = append(events, notifier.Event{
			Source:    s.name,
			TradeDate: td,
			Market:    "CN-A",
			Symbol:    a.tsCode,
			Title:     fmt.Sprintf("Repo rate %.2f%% (%s)", a.ratePct, a.tsCode),
			Body:      body,
			Tags: map[string]string{
				"kind":     "repo",
				"strategy": "yield_spike",
				"tier":     s.tier,
			},
			Data: map[string]interface{}{
				"rate_pct":            a.ratePct,
				"threshold_yield_pct": s.minYieldPct,
				"expected_edge_pct":   a.ratePct - s.minYieldPct,
				"close":               a.close,
				"weight":              a.weight,
				"amount":              a.amount,
				"vol":                 a.vol,
				"avg_amt":             a.avgAmt,
			},
		})
	}
	return events, nil
}

func normalizeRepoCodes(in []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, raw := range in {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		u := strings.ToUpper(s)
		// Allow "SH204001"/"SZ131810" style.
		if strings.HasPrefix(u, "SH") && len(u) > 2 {
			u = u[2:] + ".SH"
		} else if strings.HasPrefix(u, "SZ") && len(u) > 2 {
			u = u[2:] + ".SZ"
		}
		// Allow "204001" -> assume SH by default (common for 204xxx repos).
		if !strings.Contains(u, ".") && len(u) == 6 {
			if strings.HasPrefix(u, "204") {
				u = u + ".SH"
			} else if strings.HasPrefix(u, "131") {
				u = u + ".SZ"
			}
		}
		if !strings.Contains(u, ".") {
			continue
		}
		if !seen[u] {
			seen[u] = true
			out = append(out, u)
		}
	}
	return out
}

func withinWindow(now time.Time, start, end string) bool {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if start == "" && end == "" {
		return true
	}
	startMin, okS := parseHHMM(start)
	endMin, okE := parseHHMM(end)
	if !okS || !okE {
		// If config is malformed, fail open (still allow alerts) rather than silently disabling.
		return true
	}
	cur := now.Hour()*60 + now.Minute()
	if startMin <= endMin {
		return cur >= startMin && cur <= endMin
	}
	// Overnight window (rare for CN markets, but support anyway).
	return cur >= startMin || cur <= endMin
}

func parseHHMM(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	s = strings.ReplaceAll(s, "：", ":")
	s = strings.ReplaceAll(s, " ", "")
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return 0, false
		}
		h, ok1 := parseInt(parts[0])
		m, ok2 := parseInt(parts[1])
		if !ok1 || !ok2 || h < 0 || h > 23 || m < 0 || m > 59 {
			return 0, false
		}
		return h*60 + m, true
	}
	if len(s) != 4 {
		return 0, false
	}
	h, ok1 := parseInt(s[:2])
	m, ok2 := parseInt(s[2:])
	if !ok1 || !ok2 || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

func parseInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}
