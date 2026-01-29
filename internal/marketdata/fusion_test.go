package marketdata

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeProvider struct {
	name string
	rate float64
	err  error
	ts   time.Time
}

func (p fakeProvider) Name() string { return p.name }

func (p fakeProvider) Fetch(ctx context.Context, symbol string) (Snapshot, error) {
	_ = ctx
	return Snapshot{
		Provider: p.name,
		Symbol:   symbol,
		TS:       p.ts,
		RatePct:  p.rate,
	}, p.err
}

func TestFusionConsensusPass(t *testing.T) {
	now := time.Date(2026, 1, 29, 15, 0, 0, 0, time.FixedZone("CST", 8*3600))
	f, err := NewFusion([]Provider{
		fakeProvider{name: "a", rate: 5.01, ts: now},
		fakeProvider{name: "b", rate: 5.03, ts: now},
		fakeProvider{name: "c", rate: 8.00, ts: now},
	}, FusionConfig{
		RequiredSources: 2,
		MaxAbsDiff:      0.05,
		Staleness:       10 * time.Second,
		Now:             func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := f.FetchFusion(context.Background(), "204001.SH")
	if err != nil {
		t.Fatal(err)
	}
	if out.Confidence != ConfidencePass {
		t.Fatalf("expected PASS, got=%s reason=%s", out.Confidence, out.Reason)
	}
	if out.ConsensusRatePct < 5.0 || out.ConsensusRatePct > 5.05 {
		t.Fatalf("unexpected consensus: %.4f", out.ConsensusRatePct)
	}
}

func TestFusionNoValidSources(t *testing.T) {
	now := time.Date(2026, 1, 29, 10, 0, 0, 0, time.UTC)
	f, err := NewFusion([]Provider{
		fakeProvider{name: "a", err: errors.New("boom")},
	}, FusionConfig{
		RequiredSources: 1,
		Now:             func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := f.FetchFusion(context.Background(), "204001.SH")
	if err != nil {
		t.Fatal(err)
	}
	if out.Confidence != ConfidenceFail || out.Reason != "no_valid_sources" {
		t.Fatalf("expected FAIL/no_valid_sources got=%s/%s", out.Confidence, out.Reason)
	}
}

func TestFusionCircuitBreaker(t *testing.T) {
	now := time.Date(2026, 1, 29, 10, 0, 0, 0, time.UTC)
	cur := now
	p := fakeProvider{name: "a", err: errors.New("boom")}
	f, err := NewFusion([]Provider{p}, FusionConfig{
		RequiredSources: 1,
		FailThreshold:   2,
		Cooldown:        1 * time.Minute,
		Now:             func() time.Time { return cur },
	})
	if err != nil {
		t.Fatal(err)
	}

	_, _ = f.FetchFusion(context.Background(), "x") // fail 1
	_, _ = f.FetchFusion(context.Background(), "x") // fail 2 -> circuit open

	out, err := f.FetchFusion(context.Background(), "x")
	if err != nil {
		t.Fatal(err)
	}
	// Provider should be skipped as circuit_open; still returns a fusion snapshot.
	foundCircuit := false
	for _, pr := range out.Providers {
		if pr.Error == "circuit_open" {
			foundCircuit = true
			break
		}
	}
	if !foundCircuit {
		t.Fatalf("expected circuit_open provider result")
	}

	// After cooldown, it should attempt fetch again.
	cur = now.Add(2 * time.Minute)
	out2, _ := f.FetchFusion(context.Background(), "x")
	foundCircuit = false
	for _, pr := range out2.Providers {
		if pr.Error == "circuit_open" {
			foundCircuit = true
		}
	}
	if foundCircuit {
		t.Fatalf("expected circuit to be closed after cooldown")
	}
}

