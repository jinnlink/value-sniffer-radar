package signals

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/notifier"
	"value-sniffer-radar/internal/tushare"
)

type FundPremium struct {
	name        string
	tier        string
	minInterval time.Duration
	market          string
	pickTopByAmount int
	minAmount       float64
	premiumLow      float64
	premiumHigh     float64
	topN            int
}

func NewFundPremium(c config.SignalConfig) *FundPremium {
	topN := c.TopN
	if topN <= 0 {
		topN = 20
	}
	market := c.Market
	if market == "" {
		market = "E"
	}
	pick := c.PickTopByAmount
	if pick <= 0 {
		pick = 50
	}
	name := c.Name
	if name == "" {
		name = "fund_premium"
	}
	tier := c.Tier
	if tier == "" {
		tier = "action"
	}
	minInt := time.Duration(c.MinIntervalSeconds) * time.Second
	return &FundPremium{
		name:            name,
		tier:            tier,
		minInterval:     minInt,
		market:          market,
		pickTopByAmount: pick,
		minAmount:       c.MinAmount,
		premiumLow:      c.PremiumPctLow,
		premiumHigh:     c.PremiumPctHigh,
		topN:            topN,
	}
}

func (s *FundPremium) Name() string { return s.name }

func (s *FundPremium) MinInterval() time.Duration { return s.minInterval }

type fundRow struct {
	tsCode string
	close  float64
	amount float64
}

type fundAlert struct {
	tsCode     string
	close      float64
	nav        float64
	premiumPct float64
	amount     float64
}

func (s *FundPremium) Evaluate(ctx context.Context, client *tushare.Client, tradeDate string) ([]notifier.Event, error) {
	// Step 1: pick top funds by amount to limit fund_nav calls.
	params := map[string]any{
		"trade_date": tradeDate,
	}
	// Not sure if fund_daily supports market param for all accounts; keep optional.
	if s.market != "" {
		params["market"] = s.market
	}
	rows, err := client.Query(ctx, "fund_daily", params, []string{"ts_code", "close", "amount"})
	if err != nil {
		return nil, err
	}

	var funds []fundRow
	for _, r := range rows {
		tsCode := tushare.GetString(r, "ts_code")
		closeP := tushare.GetFloat(r, "close")
		amount := tushare.GetFloat(r, "amount")
		if tsCode == "" || closeP <= 0 {
			continue
		}
		if s.minAmount > 0 && amount < s.minAmount {
			continue
		}
		funds = append(funds, fundRow{tsCode: tsCode, close: closeP, amount: amount})
	}
	sort.Slice(funds, func(i, j int) bool { return funds[i].amount > funds[j].amount })
	if len(funds) > s.pickTopByAmount {
		funds = funds[:s.pickTopByAmount]
	}

	// Step 2: per fund fetch NAV.
	var alerts []fundAlert
	for _, f := range funds {
		navRows, err := client.Query(ctx, "fund_nav", map[string]any{
			"ts_code":    f.tsCode,
			"start_date": tradeDate,
			"end_date":   tradeDate,
		}, []string{"ts_code", "nav_date", "unit_nav"})
		if err != nil {
			continue
		}
		nav := 0.0
		for _, nr := range navRows {
			nav = tushare.GetFloat(nr, "unit_nav")
			if nav > 0 {
				break
			}
		}
		if nav <= 0 {
			continue
		}
		premiumPct := (f.close - nav) / nav * 100.0
		if premiumPct <= s.premiumLow || premiumPct >= s.premiumHigh {
			alerts = append(alerts, fundAlert{
				tsCode:     f.tsCode,
				close:      f.close,
				nav:        nav,
				premiumPct: premiumPct,
				amount:     f.amount,
			})
		}
	}

	sort.Slice(alerts, func(i, j int) bool { return math.Abs(alerts[i].premiumPct) > math.Abs(alerts[j].premiumPct) })
	if len(alerts) > s.topN {
		alerts = alerts[:s.topN]
	}
	if len(alerts) == 0 {
		return nil, nil
	}

	events := make([]notifier.Event, 0, len(alerts))
	for _, a := range alerts {
		body := fmt.Sprintf("close=%.4f\nnav=%.4f\npremium=%.2f%%\namount=%.0f\n", a.close, a.nav, a.premiumPct, a.amount)
		events = append(events, notifier.Event{
			Source:    s.name,
			TradeDate: tradeDate,
			Market:    "CN-A",
			Symbol:    a.tsCode,
			Title:     fmt.Sprintf("Fund premium %.2f%% (%s)", a.premiumPct, a.tsCode),
			Body:      body,
			Tags: map[string]string{
				"kind": "fund",
				"tier": s.tier,
			},
			Data: map[string]interface{}{
				"premium_pct": a.premiumPct,
				"close":       a.close,
				"nav":         a.nav,
				"amount":      a.amount,
			},
		})
	}
	return events, nil
}
