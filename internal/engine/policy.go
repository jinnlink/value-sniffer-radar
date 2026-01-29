package engine

import (
	"math"
	"strconv"
	"strings"

	"value-sniffer-radar/internal/notifier"
)

func (e *Engine) applyNetEdgePolicy(events []notifier.Event) ([]notifier.Event, int) {
	if e.cfg.Engine.ActionNetEdgeMinPct <= 0 {
		// Still compute net_edge_pct best-effort for paper log/analysis when possible.
		out := make([]notifier.Event, 0, len(events))
		for _, ev := range events {
			out = append(out, withNetEdge(ev, e))
		}
		return out, 0
	}

	out := make([]notifier.Event, 0, len(events))
	downgraded := 0
	for _, ev := range events {
		ev2 := withNetEdge(ev, e)
		if eventTier(ev2) == "action" {
			net, ok := getFloat(ev2.Data, "net_edge_pct")
			if !ok {
				ev2 = downgrade(ev2, "missing_net_edge_pct", e.cfg.Engine.ActionNetEdgeMinPct)
				downgraded++
			} else if net < e.cfg.Engine.ActionNetEdgeMinPct {
				ev2 = downgrade(ev2, "net_edge_below_threshold", e.cfg.Engine.ActionNetEdgeMinPct)
				downgraded++
			}
		}
		out = append(out, ev2)
	}
	return out, downgraded
}

func withNetEdge(ev notifier.Event, e *Engine) notifier.Event {
	ev = ensureMaps(ev)

	// expected_edge_pct (required for meaningful net edge)
	expected, okExpected := getFloat(ev.Data, "expected_edge_pct")
	if !okExpected {
		// Do not invent expected edge; record and return.
		if _, ok := ev.Data["net_edge_pct"]; !ok {
			ev.Data["net_edge_pct"] = 0.0
			ev.Data["net_edge_reason"] = "missing_expected_edge_pct"
		}
		return ev
	}

	spread, okSpread := getFloat(ev.Data, "spread_pct")
	if !okSpread {
		spread = e.cfg.Engine.DefaultSpreadPct
	}
	slippage, okSlip := getFloat(ev.Data, "slippage_pct")
	if !okSlip {
		slippage = e.cfg.Engine.DefaultSlippagePct
	}
	fee, okFee := getFloat(ev.Data, "fee_pct")
	if !okFee {
		fee = e.cfg.Engine.DefaultFeePct
		if e.cfg.Engine.FeePctByMarket != nil {
			if v, ok := e.cfg.Engine.FeePctByMarket[ev.Market]; ok {
				fee = v
			}
		}
	}

	net := expected - spread - slippage - fee
	if math.IsNaN(net) || math.IsInf(net, 0) {
		net = 0
	}

	ev.Data["expected_edge_pct"] = expected
	ev.Data["spread_pct"] = spread
	ev.Data["slippage_pct"] = slippage
	ev.Data["fee_pct"] = fee
	ev.Data["net_edge_pct"] = net
	return ev
}

func downgrade(ev notifier.Event, reason string, minNetEdge float64) notifier.Event {
	ev = ensureMaps(ev)
	ev.Tags["tier"] = "observe"
	ev.Tags["policy"] = "net_edge"
	ev.Data["policy_downgrade_reason"] = reason
	ev.Data["policy_action_net_edge_min_pct"] = minNetEdge
	return ev
}

func ensureMaps(ev notifier.Event) notifier.Event {
	if ev.Tags == nil {
		ev.Tags = map[string]string{}
	}
	if ev.Data == nil {
		ev.Data = map[string]interface{}{}
	}
	return ev
}

func getFloat(m map[string]interface{}, key string) (float64, bool) {
	if m == nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(s, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
