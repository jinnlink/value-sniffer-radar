package marketdata

import (
	"context"
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

type FusionConfig struct {
	Timeout time.Duration

	RequiredSources int
	MaxAbsDiff      float64
	Staleness       time.Duration
	MinValid        float64
	MaxValid        float64

	FailThreshold    int
	OutlierThreshold int
	Cooldown         time.Duration

	Now func() time.Time
}

type FusionEngine struct {
	providers []Provider
	cfg       FusionConfig

	mu    sync.Mutex
	state map[string]*providerState
}

type providerState struct {
	score             float64
	consecutiveFails  int
	consecutiveOutlier int
	disabledUntil     time.Time
}

func NewFusion(providers []Provider, cfg FusionConfig) (*FusionEngine, error) {
	if len(providers) == 0 {
		return nil, errors.New("marketdata: no providers")
	}
	if cfg.RequiredSources <= 0 {
		cfg.RequiredSources = 2
	}
	if cfg.MaxAbsDiff <= 0 {
		cfg.MaxAbsDiff = 0.05
	}
	if cfg.Staleness <= 0 {
		cfg.Staleness = 10 * time.Second
	}
	if cfg.MaxValid <= 0 {
		cfg.MaxValid = 20
	}
	if cfg.FailThreshold <= 0 {
		cfg.FailThreshold = 3
	}
	if cfg.OutlierThreshold <= 0 {
		cfg.OutlierThreshold = 3
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = 2 * time.Minute
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 1500 * time.Millisecond
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}

	st := make(map[string]*providerState, len(providers))
	for _, p := range providers {
		st[p.Name()] = &providerState{score: 0.5}
	}
	return &FusionEngine{
		providers: providers,
		cfg:       cfg,
		state:     st,
	}, nil
}

func (f *FusionEngine) FetchFusion(ctx context.Context, symbol string) (FusionSnapshot, error) {
	now := f.cfg.Now()
	results := make([]ProviderResult, 0, len(f.providers))

	type one struct {
		name string
		snap Snapshot
		err  error
	}

	outs := make(chan one, len(f.providers))
	var wg sync.WaitGroup

	for _, p := range f.providers {
		p := p
		name := p.Name()

		if f.isDisabled(name, now) {
			results = append(results, ProviderResult{
				Provider: name,
				Error:    "circuit_open",
			})
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			cctx, cancel := context.WithTimeout(ctx, f.cfg.Timeout)
			defer cancel()
			snap, err := p.Fetch(cctx, symbol)
			outs <- one{name: name, snap: snap, err: err}
		}()
	}

	wg.Wait()
	close(outs)

	// Collect, then do quality filtering.
	var candidates []float64
	candidateByProvider := map[string]float64{}
	for o := range outs {
		if o.err != nil {
			results = append(results, ProviderResult{
				Provider: o.name,
				Snapshot: o.snap,
				Error:    o.err.Error(),
			})
			continue
		}
		stale := false
		if !o.snap.TS.IsZero() && now.Sub(o.snap.TS) > f.cfg.Staleness {
			stale = true
		}
		r := o.snap.RatePct
		invalid := (r < f.cfg.MinValid) || (r > f.cfg.MaxValid) || math.IsNaN(r) || math.IsInf(r, 0)
		if stale || invalid {
			results = append(results, ProviderResult{
				Provider: o.name,
				Snapshot: o.snap,
				Stale:    stale,
				Error:    "invalid_or_stale",
			})
			continue
		}
		candidates = append(candidates, r)
		candidateByProvider[o.name] = r
		results = append(results, ProviderResult{
			Provider: o.name,
			Snapshot: o.snap,
		})
	}

	if len(candidates) == 0 {
		f.updateStates(now, results, nil)
		return FusionSnapshot{
			Symbol:     symbol,
			TS:         now,
			Confidence: ConfidenceFail,
			Reason:     "no_valid_sources",
			Providers:  results,
		}, nil
	}

	consensus := median(candidates)
	inliers := map[string]bool{}
	for name, val := range candidateByProvider {
		if math.Abs(val-consensus) <= f.cfg.MaxAbsDiff {
			inliers[name] = true
		}
	}

	conf := ConfidenceFail
	reason := "insufficient_consensus"
	if len(inliers) >= f.cfg.RequiredSources {
		conf = ConfidencePass
		reason = "consensus_pass"
	}

	// Mark inlier/outlier in ProviderResult.
	for i := range results {
		pr := &results[i]
		if pr.Error != "" {
			continue
		}
		if inliers[pr.Provider] {
			pr.Inlier = true
		} else {
			pr.Outlier = true
		}
	}

	f.updateStates(now, results, inliers)

	return FusionSnapshot{
		Symbol:           symbol,
		TS:               now,
		ConsensusRatePct: consensus,
		Confidence:       conf,
		Reason:           reason,
		Providers:        results,
	}, nil
}

func (f *FusionEngine) isDisabled(provider string, now time.Time) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	st := f.state[provider]
	if st == nil {
		st = &providerState{score: 0.5}
		f.state[provider] = st
	}
	return now.Before(st.disabledUntil)
}

func (f *FusionEngine) updateStates(now time.Time, results []ProviderResult, inliers map[string]bool) {
	const alpha = 0.15
	f.mu.Lock()
	defer f.mu.Unlock()

	// If no results were collected (e.g. disabled all), nothing to update.
	if results == nil {
		return
	}

	for _, r := range results {
		st := f.state[r.Provider]
		if st == nil {
			st = &providerState{score: 0.5}
			f.state[r.Provider] = st
		}

		if r.Error != "" {
			st.score = st.score * (1 - alpha)
			st.consecutiveFails++
			st.consecutiveOutlier++
		} else {
			st.consecutiveFails = 0
			if inliers != nil && inliers[r.Provider] {
				st.score = st.score + alpha*(1-st.score)
				st.consecutiveOutlier = 0
			} else {
				st.score = st.score * (1 - alpha)
				st.consecutiveOutlier++
			}
		}

		if st.consecutiveFails >= f.cfg.FailThreshold || st.consecutiveOutlier >= f.cfg.OutlierThreshold {
			st.disabledUntil = now.Add(f.cfg.Cooldown)
			st.consecutiveFails = 0
			st.consecutiveOutlier = 0
		}
	}
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	ys := make([]float64, len(xs))
	copy(ys, xs)
	sort.Float64s(ys)
	m := len(ys) / 2
	if len(ys)%2 == 1 {
		return ys[m]
	}
	return (ys[m-1] + ys[m]) / 2
}
