package engine

import (
	"testing"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/notifier"
)

func TestNetEdgePolicy_DowngradesActionBelowThreshold(t *testing.T) {
	e := &Engine{
		cfg: &config.Config{
			Engine: config.EngineConfig{
				ActionNetEdgeMinPct: 0.05,
				DefaultSpreadPct:    0.0,
				DefaultSlippagePct:  0.05,
				DefaultFeePct:       0.05,
			},
		},
		dailySent: map[string]int{},
	}

	in := []notifier.Event{
		{
			Source:    "sig",
			TradeDate: "20260101",
			Market:    "CN-A",
			Symbol:    "X",
			Title:     "t",
			Body:      "b",
			Tags:      map[string]string{"tier": "action"},
			Data:      map[string]interface{}{"expected_edge_pct": 0.08},
		},
	}

	out, downgraded := e.applyNetEdgePolicy(in)
	if downgraded != 1 {
		t.Fatalf("downgraded=%d want=1", downgraded)
	}
	if got := eventTier(out[0]); got != "observe" {
		t.Fatalf("tier=%s want=observe", got)
	}
	if out[0].Data == nil {
		t.Fatalf("expected data to exist")
	}
	if _, ok := out[0].Data["net_edge_pct"]; !ok {
		t.Fatalf("expected net_edge_pct in data")
	}
	if r, ok := out[0].Data["policy_downgrade_reason"]; !ok || r == "" {
		t.Fatalf("expected policy_downgrade_reason")
	}
}

func TestDailyCaps_PerSignalAndGlobalActionCaps(t *testing.T) {
	e := &Engine{
		cfg: &config.Config{
			Engine: config.EngineConfig{
				ActionMaxEventsPerDay:  30,
				ObserveMaxEventsPerDay: 1000,
				ActionMaxEventsPerSignalPerDay: map[string]int{
					"sigA": 10,
				},
			},
		},
		dailySent: map[string]int{},
	}

	var events []notifier.Event
	for i := 0; i < 20; i++ {
		events = append(events, notifier.Event{Source: "sigA", TradeDate: "20260101", Title: "a", Body: "a"})
	}
	for i := 0; i < 20; i++ {
		events = append(events, notifier.Event{Source: "sigB", TradeDate: "20260101", Title: "b", Body: "b"})
	}

	out, _, _ := e.applyDailyCaps(events, "20260101")

	action := 0
	observe := 0
	actionA := 0
	for _, ev := range out {
		switch eventTier(ev) {
		case "observe":
			observe++
		default:
			action++
			if ev.Source == "sigA" {
				actionA++
			}
		}
	}
	if action != 30 {
		t.Fatalf("action=%d want=30", action)
	}
	if observe != 10 {
		t.Fatalf("observe=%d want=10", observe)
	}
	if actionA != 10 {
		t.Fatalf("action(sigA)=%d want=10", actionA)
	}
}

func TestDailyCaps_ActionCapDowngradesToObserve(t *testing.T) {
	e := &Engine{
		cfg: &config.Config{
			Engine: config.EngineConfig{
				ActionMaxEventsPerDay:  5,
				ObserveMaxEventsPerDay: 1000,
			},
		},
		dailySent: map[string]int{},
	}

	var events []notifier.Event
	for i := 0; i < 10; i++ {
		events = append(events, notifier.Event{Source: "sig", TradeDate: "20260101", Title: "t", Body: "b"})
	}

	out, _, _ := e.applyDailyCaps(events, "20260101")

	action := 0
	observe := 0
	for _, ev := range out {
		if eventTier(ev) == "observe" {
			observe++
			continue
		}
		action++
	}
	if action != 5 || observe != 5 {
		t.Fatalf("action=%d observe=%d want action=5 observe=5", action, observe)
	}
}
