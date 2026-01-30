package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/marketdata"
	"value-sniffer-radar/internal/notifier"
	"value-sniffer-radar/internal/signals"
	"value-sniffer-radar/internal/tushare"
)

type Engine struct {
	cfg        *config.Config
	client     *tushare.Client
	md         marketdata.Fusion
	notifiers  []notifier.Notifier
	sigs       []signals.Signal
	sent       map[string]time.Time
	symbolLast map[string]time.Time
	lastEval   map[string]time.Time
	dailySent  map[string]int
	recoQuotas map[string]int // optional overrides (signal -> daily action quota)
}

func New(cfg *config.Config) (*Engine, error) {
	var client *tushare.Client
	if cfg.RequiresTushare() {
		token, ok := os.LookupEnv(cfg.Tushare.TokenEnv)
		if !ok || strings.TrimSpace(token) == "" {
			return nil, fmt.Errorf("missing Tushare token env: %s", cfg.Tushare.TokenEnv)
		}
		client = tushare.New(tushare.Options{
			BaseURL:        cfg.Tushare.BaseURL,
			Token:          token,
			TimeoutSeconds: cfg.Tushare.TimeoutSeconds,
			MaxRetries:     cfg.Engine.MaxAPIRetries,
		})
	}

	notifs, err := notifier.BuildAll(cfg.Notifiers)
	if err != nil {
		return nil, err
	}

	sigs, err := signals.BuildAll(cfg.Signals)
	if err != nil {
		return nil, err
	}

	md, err := marketdata.Build(cfg.Marketdata)
	if err != nil {
		return nil, err
	}

	return &Engine{
		cfg:        cfg,
		client:     client,
		md:         md,
		notifiers:  notifs,
		sigs:       sigs,
		sent:       map[string]time.Time{},
		symbolLast: map[string]time.Time{},
		lastEval:   map[string]time.Time{},
		dailySent:  map[string]int{},
		recoQuotas: nil,
	}, nil
}

func (e *Engine) Run() error {
	ctx := context.Background()

	e.loadRecoIfConfigured()

	ticker := time.NewTicker(time.Duration(e.cfg.Engine.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		if err := e.runOnce(ctx); err != nil {
			log.Printf("runOnce error: %v", err)
		}
		<-ticker.C
	}
}

func (e *Engine) runOnce(ctx context.Context) error {
	tradeDate, err := e.resolveTradeDate(ctx)
	if err != nil {
		return err
	}

	var allEvents []notifier.Event
	now := time.Now()
	for _, sig := range e.sigs {
		if minInt := sig.MinInterval(); minInt > 0 {
			if last, ok := e.lastEval[sig.Name()]; ok && now.Sub(last) < minInt {
				continue
			}
			e.lastEval[sig.Name()] = now
		}
		evs, err := sig.Evaluate(ctx, e.client, tradeDate, e.md)
		if err != nil {
			log.Printf("signal %s error: %v", sig.Name(), err)
			continue
		}
		allEvents = append(allEvents, evs...)
	}

	if len(allEvents) == 0 {
		log.Printf("no events (trade_date=%s)", tradeDate)
		return nil
	}

	allEvents, dropped := e.applyDedupe(allEvents)
	if dropped > 0 {
		log.Printf("events=%d dropped=%d (trade_date=%s)", len(allEvents), dropped, tradeDate)
	} else {
		log.Printf("events=%d (trade_date=%s)", len(allEvents), tradeDate)
	}

	allEvents, cdDropped := e.applySymbolCooldown(allEvents)
	if cdDropped > 0 {
		log.Printf("cooldown_dropped=%d (trade_date=%s)", cdDropped, tradeDate)
	}

	allEvents, downgraded := e.applyNetEdgePolicy(allEvents)
	if downgraded > 0 {
		log.Printf("net_edge_downgraded=%d (trade_date=%s)", downgraded, tradeDate)
	}

	allEvents, capDropped := e.applyMaxEventsPerRun(allEvents)
	if capDropped > 0 {
		log.Printf("cap_dropped=%d max_events_per_run=%d (trade_date=%s)", capDropped, e.cfg.Engine.MaxEventsPerRun, tradeDate)
	}

	allEvents, dayDroppedA, dayDroppedO := e.applyDailyCaps(allEvents, tradeDate)
	if dayDroppedA > 0 || dayDroppedO > 0 {
		log.Printf("daily_cap_dropped action=%d observe=%d (trade_date=%s)", dayDroppedA, dayDroppedO, tradeDate)
	}

	if len(allEvents) == 0 {
		return nil
	}
	for _, n := range e.notifiers {
		if err := n.Notify(ctx, allEvents); err != nil {
			log.Printf("notifier %s error: %v", n.Name(), err)
		}
	}
	return nil
}

func (e *Engine) resolveTradeDate(ctx context.Context) (string, error) {
	switch e.cfg.Engine.TradeDateMode {
	case "fixed":
		return e.cfg.Engine.FixedTradeDate, nil
	case "latest_open":
		if e.client == nil {
			return "", fmt.Errorf("trade_date_mode=latest_open requires Tushare client (set %s)", e.cfg.Tushare.TokenEnv)
		}
		return e.client.LatestOpenTradeDate(ctx, 45)
	default:
		return "", fmt.Errorf("unknown trade_date_mode: %s", e.cfg.Engine.TradeDateMode)
	}
}

func (e *Engine) applyDedupe(events []notifier.Event) ([]notifier.Event, int) {
	if e.cfg.Engine.DedupeSeconds == -1 {
		return events, 0
	}
	ttl := time.Duration(e.cfg.Engine.DedupeSeconds) * time.Second
	if ttl <= 0 {
		return events, 0
	}

	now := time.Now()
	cutoff := now.Add(-ttl)
	for k, t := range e.sent {
		if t.Before(cutoff) {
			delete(e.sent, k)
		}
	}

	out := make([]notifier.Event, 0, len(events))
	dropped := 0
	for _, ev := range events {
		key := eventKey(ev)
		if t, ok := e.sent[key]; ok && !t.Before(cutoff) {
			dropped++
			continue
		}
		e.sent[key] = now
		out = append(out, ev)
	}
	return out, dropped
}

func (e *Engine) applySymbolCooldown(events []notifier.Event) ([]notifier.Event, int) {
	now := time.Now()
	out := make([]notifier.Event, 0, len(events))
	dropped := 0

	for _, ev := range events {
		if strings.TrimSpace(ev.Symbol) == "" {
			out = append(out, ev)
			continue
		}
		tier := eventTier(ev)
		ttlSeconds := e.cfg.Engine.ActionSymbolCooldownSeconds
		if tier == "observe" {
			ttlSeconds = e.cfg.Engine.ObserveSymbolCooldownSeconds
		}
		if ttlSeconds == -1 {
			out = append(out, ev)
			continue
		}
		ttl := time.Duration(ttlSeconds) * time.Second
		if ttl <= 0 {
			out = append(out, ev)
			continue
		}

		key := tier + "|" + ev.Source + "|" + ev.Symbol
		cutoff := now.Add(-ttl)
		if t, ok := e.symbolLast[key]; ok && !t.Before(cutoff) {
			dropped++
			continue
		}
		e.symbolLast[key] = now
		out = append(out, ev)
	}

	// best-effort cleanup to bound map size
	// (use the larger ttl for cleanup threshold)
	cleanupTTL := time.Duration(e.cfg.Engine.ObserveSymbolCooldownSeconds) * time.Second
	if cleanupTTL <= 0 {
		cleanupTTL = 2 * time.Hour
	}
	cleanupCutoff := now.Add(-cleanupTTL)
	for k, t := range e.symbolLast {
		if t.Before(cleanupCutoff) {
			delete(e.symbolLast, k)
		}
	}

	return out, dropped
}

func (e *Engine) applyMaxEventsPerRun(events []notifier.Event) ([]notifier.Event, int) {
	max := e.cfg.Engine.MaxEventsPerRun
	if max <= 0 {
		return events, 0
	}

	// Apply tier caps first (action), then fill with observe.
	actionCap := e.cfg.Engine.ActionMaxEventsPerRun
	observeCap := e.cfg.Engine.ObserveMaxEventsPerRun

	var action, observe, other []notifier.Event
	for _, ev := range events {
		switch eventTier(ev) {
		case "action":
			action = append(action, ev)
		case "observe":
			observe = append(observe, ev)
		default:
			other = append(other, ev)
		}
	}

	trim := func(xs []notifier.Event, cap int) ([]notifier.Event, int) {
		if cap <= 0 || len(xs) <= cap {
			return xs, 0
		}
		return xs[:cap], len(xs) - cap
	}

	action, droppedA := trim(action, actionCap)
	observe, droppedO := trim(observe, observeCap)

	out := make([]notifier.Event, 0, len(action)+len(observe)+len(other))
	out = append(out, action...)
	out = append(out, observe...)
	out = append(out, other...)

	dropped := droppedA + droppedO
	if len(out) <= max {
		return out, dropped
	}
	return out[:max], dropped + (len(out) - max)
}

func (e *Engine) applyDailyCaps(events []notifier.Event, tradeDate string) ([]notifier.Event, int, int) {
	// Keep only current trade_date keys.
	for k := range e.dailySent {
		if !strings.HasPrefix(k, tradeDate+"|") {
			delete(e.dailySent, k)
		}
	}

	actionCap := e.cfg.Engine.ActionMaxEventsPerDay
	observeCap := e.cfg.Engine.ObserveMaxEventsPerDay
	perSignal := e.cfg.Engine.ActionMaxEventsPerSignalPerDay
	if e.recoQuotas != nil && len(e.recoQuotas) > 0 {
		perSignal = e.recoQuotas
	}

	out := make([]notifier.Event, 0, len(events))
	droppedA := 0
	droppedO := 0
	for _, ev := range events {
		tier := eventTier(ev)

		// If action budgets are exceeded, downgrade to observe (broad coverage) rather than dropping.
		// This keeps coverage while still enforcing "action" quality/budget.
		if tier == "action" {
			if actionCap > 0 && e.dailySent[tradeDate+"|action"] >= actionCap {
				ev = downgradeTier(ev, "daily_action_cap")
				tier = "observe"
			}
			if tier == "action" && perSignal != nil {
				if cap, ok := perSignal[ev.Source]; ok && cap > 0 {
					keySig := tradeDate + "|action|" + ev.Source
					if e.dailySent[keySig] >= cap {
						ev = downgradeTier(ev, "per_signal_action_cap")
						tier = "observe"
					}
				}
			}
		}

		switch tier {
		case "observe":
			key := tradeDate + "|observe"
			if observeCap > 0 && e.dailySent[key] >= observeCap {
				droppedO++
				continue
			}
			e.dailySent[key] = e.dailySent[key] + 1
		default:
			key := tradeDate + "|action"
			if actionCap > 0 && e.dailySent[key] >= actionCap {
				droppedA++
				continue
			}
			e.dailySent[key] = e.dailySent[key] + 1
			if perSignal != nil {
				if cap, ok := perSignal[ev.Source]; ok && cap > 0 {
					keySig := tradeDate + "|action|" + ev.Source
					e.dailySent[keySig] = e.dailySent[keySig] + 1
				}
			}
		}
		out = append(out, ev)
	}
	return out, droppedA, droppedO
}

func eventTier(e notifier.Event) string {
	if e.Tags == nil {
		return "action"
	}
	if t, ok := e.Tags["tier"]; ok {
		switch strings.TrimSpace(strings.ToLower(t)) {
		case "observe":
			return "observe"
		case "action":
			return "action"
		}
	}
	return "action"
}

func downgradeTier(e notifier.Event, reason string) notifier.Event {
	if e.Tags == nil {
		e.Tags = map[string]string{}
	}
	e.Tags["tier"] = "observe"
	if e.Data == nil {
		e.Data = map[string]interface{}{}
	}
	e.Data["policy_downgrade_reason"] = reason
	return e
}

func eventKey(e notifier.Event) string {
	h := sha256.New()
	_, _ = h.Write([]byte(e.Source))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(e.TradeDate))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(e.Title))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(e.Body))
	return hex.EncodeToString(h.Sum(nil))
}
