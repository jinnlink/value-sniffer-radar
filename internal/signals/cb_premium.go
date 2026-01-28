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

type CBPremium struct {
	name       string
	tier       string
	minInterval time.Duration
	minAmount   float64
	premiumLow  float64
	premiumHigh float64
	topN        int
}

func NewCBPremium(c config.SignalConfig) *CBPremium {
	topN := c.TopN
	if topN <= 0 {
		topN = 20
	}
	name := c.Name
	if name == "" {
		name = "cb_premium"
	}
	tier := c.Tier
	if tier == "" {
		tier = "action"
	}
	minInt := time.Duration(c.MinIntervalSeconds) * time.Second
	return &CBPremium{
		name:        name,
		tier:        tier,
		minInterval: minInt,
		minAmount:   c.MinAmount,
		premiumLow:  c.PremiumPctLow,
		premiumHigh: c.PremiumPctHigh,
		topN:        topN,
	}
}

func (s *CBPremium) Name() string { return s.name }

func (s *CBPremium) MinInterval() time.Duration { return s.minInterval }

type cbBasic struct {
	tsCode    string
	stkCode   string
	convPrice float64
}

type cbAlert struct {
	tsCode     string
	bondClose  float64
	stkCode    string
	stkClose   float64
	convPrice  float64
	convValue  float64
	premiumPct float64
	amount     float64
}

func (s *CBPremium) Evaluate(ctx context.Context, client *tushare.Client, tradeDate string) ([]notifier.Event, error) {
	cbBasics, err := client.Query(ctx, "cb_basic", map[string]any{
		"list_status": "L",
	}, []string{"ts_code", "stk_code", "conv_price"})
	if err != nil {
		return nil, err
	}
	basicMap := map[string]cbBasic{}
	for _, r := range cbBasics {
		tsCode := tushare.GetString(r, "ts_code")
		if tsCode == "" {
			continue
		}
		basicMap[tsCode] = cbBasic{
			tsCode:    tsCode,
			stkCode:   tushare.GetString(r, "stk_code"),
			convPrice: tushare.GetFloat(r, "conv_price"),
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

	var alerts []cbAlert
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

		if premiumPct <= s.premiumLow || premiumPct >= s.premiumHigh {
			alerts = append(alerts, cbAlert{
				tsCode:     tsCode,
				bondClose:  closeP,
				stkCode:    b.stkCode,
				stkClose:   stkClose,
				convPrice:  b.convPrice,
				convValue:  convValue,
				premiumPct: premiumPct,
				amount:     amount,
			})
		}
	}

	sort.Slice(alerts, func(i, j int) bool {
		return math.Abs(alerts[i].premiumPct) > math.Abs(alerts[j].premiumPct)
	})
	if len(alerts) > s.topN {
		alerts = alerts[:s.topN]
	}
	if len(alerts) == 0 {
		return nil, nil
	}

	events := make([]notifier.Event, 0, len(alerts))
	for _, a := range alerts {
		body := fmt.Sprintf(
			"bond_close=%.2f\nstk=%s\nstk_close=%.2f\nconv_price=%.4f\nconv_value=%.2f\npremium=%.2f%%\namount=%.0f\n",
			a.bondClose, a.stkCode, a.stkClose, a.convPrice, a.convValue, a.premiumPct, a.amount,
		)
		events = append(events, notifier.Event{
			Source:    s.name,
			TradeDate: tradeDate,
			Market:    "CN-A",
			Symbol:    a.tsCode,
			Title:     fmt.Sprintf("CB premium %.2f%% (%s)", a.premiumPct, a.tsCode),
			Body:      body,
			Tags: map[string]string{
				"kind":      "cb",
				"underlying": a.stkCode,
				"tier":      s.tier,
			},
			Data: map[string]interface{}{
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
