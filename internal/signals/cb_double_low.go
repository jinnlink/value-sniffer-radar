package signals

import (
	"context"
	"fmt"
	"sort"
	"time"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/notifier"
	"value-sniffer-radar/internal/tushare"
)

// CBDoubleLow alerts when (bond_price + premium_pct) <= threshold.
// It is NOT arbitrage; it's a classic low-risk-ish selection heuristic for CN convertible bonds.
type CBDoubleLow struct {
	name        string
	tier        string
	minInterval time.Duration
	minAmount    float64
	maxDoubleLow float64
	topN         int
}

func NewCBDoubleLow(c config.SignalConfig) *CBDoubleLow {
	topN := c.TopN
	if topN <= 0 {
		topN = 20
	}
	thr := c.MaxDoubleLow
	if thr <= 0 {
		thr = 125
	}
	name := c.Name
	if name == "" {
		name = "cb_double_low"
	}
	tier := c.Tier
	if tier == "" {
		tier = "action"
	}
	minInt := time.Duration(c.MinIntervalSeconds) * time.Second
	return &CBDoubleLow{
		name:         name,
		tier:         tier,
		minInterval:  minInt,
		minAmount:    c.MinAmount,
		maxDoubleLow: thr,
		topN:         topN,
	}
}

func (s *CBDoubleLow) Name() string { return s.name }

func (s *CBDoubleLow) MinInterval() time.Duration { return s.minInterval }

type cbDLBasic struct {
	stkCode   string
	convPrice float64
	name      string
}

type cbDLAlert struct {
	tsCode     string
	name       string
	bondClose  float64
	stkCode    string
	stkClose   float64
	convPrice  float64
	convValue  float64
	premiumPct float64
	doubleLow  float64
	amount     float64
}

func (s *CBDoubleLow) Evaluate(ctx context.Context, client *tushare.Client, tradeDate string) ([]notifier.Event, error) {
	cbBasics, err := client.Query(ctx, "cb_basic", map[string]any{
		"list_status": "L",
	}, []string{"ts_code", "stk_code", "conv_price", "bond_short_name"})
	if err != nil {
		return nil, err
	}
	basicMap := map[string]cbDLBasic{}
	for _, r := range cbBasics {
		tsCode := tushare.GetString(r, "ts_code")
		if tsCode == "" {
			continue
		}
		basicMap[tsCode] = cbDLBasic{
			stkCode:   tushare.GetString(r, "stk_code"),
			convPrice: tushare.GetFloat(r, "conv_price"),
			name:      tushare.GetString(r, "bond_short_name"),
		}
	}

	cbDaily, err := client.Query(ctx, "cb_daily", map[string]any{
		"trade_date": tradeDate,
	}, []string{"ts_code", "close", "amount"})
	if err != nil {
		return nil, err
	}

	stocks, err := client.Query(ctx, "daily", map[string]any{
		"trade_date": tradeDate,
	}, []string{"ts_code", "close"})
	if err != nil {
		return nil, err
	}
	stockMap := map[string]float64{}
	for _, r := range stocks {
		tsCode := tushare.GetString(r, "ts_code")
		if tsCode == "" {
			continue
		}
		stockMap[tsCode] = tushare.GetFloat(r, "close")
	}

	var alerts []cbDLAlert
	for _, r := range cbDaily {
		tsCode := tushare.GetString(r, "ts_code")
		closeP := tushare.GetFloat(r, "close")
		amount := tushare.GetFloat(r, "amount")
		if tsCode == "" || closeP <= 0 {
			continue
		}
		if s.minAmount > 0 && amount < s.minAmount {
			continue
		}

		b, ok := basicMap[tsCode]
		if !ok || b.stkCode == "" || b.convPrice <= 0 {
			continue
		}
		stkClose := stockMap[b.stkCode]
		if stkClose <= 0 {
			continue
		}
		convValue := stkClose * (100.0 / b.convPrice)
		if convValue <= 0 {
			continue
		}
		premiumPct := (closeP - convValue) / convValue * 100.0
		doubleLow := closeP + premiumPct

		if doubleLow <= s.maxDoubleLow {
			alerts = append(alerts, cbDLAlert{
				tsCode:     tsCode,
				name:       b.name,
				bondClose:  closeP,
				stkCode:    b.stkCode,
				stkClose:   stkClose,
				convPrice:  b.convPrice,
				convValue:  convValue,
				premiumPct: premiumPct,
				doubleLow:  doubleLow,
				amount:     amount,
			})
		}
	}

	sort.Slice(alerts, func(i, j int) bool { return alerts[i].doubleLow < alerts[j].doubleLow })
	if len(alerts) > s.topN {
		alerts = alerts[:s.topN]
	}
	if len(alerts) == 0 {
		return nil, nil
	}

	events := make([]notifier.Event, 0, len(alerts))
	for _, a := range alerts {
		body := fmt.Sprintf(
			"name=%s\ndouble_low=%.2f\nbond_close=%.2f\npremium=%.2f%%\nstk=%s\nstk_close=%.2f\nconv_price=%.4f\nconv_value=%.2f\namount=%.0f\n",
			a.name, a.doubleLow, a.bondClose, a.premiumPct, a.stkCode, a.stkClose, a.convPrice, a.convValue, a.amount,
		)
		events = append(events, notifier.Event{
			Source:    s.name,
			TradeDate: tradeDate,
			Market:    "CN-A",
			Symbol:    a.tsCode,
			Title:     fmt.Sprintf("CB double-low %.2f (%s)", a.doubleLow, a.tsCode),
			Body:      body,
			Tags: map[string]string{
				"kind":      "cb",
				"strategy":  "double_low",
				"underlying": a.stkCode,
				"tier":      s.tier,
			},
			Data: map[string]interface{}{
				"double_low":  a.doubleLow,
				"premium_pct": a.premiumPct,
				"bond_close":  a.bondClose,
				"stk_code":    a.stkCode,
				"stk_close":   a.stkClose,
				"conv_price":  a.convPrice,
				"conv_value":  a.convValue,
				"amount":      a.amount,
			},
		})
	}
	return events, nil
}
