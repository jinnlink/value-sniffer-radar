package marketdata

import (
	"context"
	"time"
)

type Snapshot struct {
	Provider string
	Symbol   string
	TS       time.Time

	// For VS_0005 we start with repo only: rate in percentage points (e.g. 1.85 means 1.85%).
	RatePct float64

	Raw map[string]any
}

type Provider interface {
	Name() string
	Fetch(ctx context.Context, symbol string) (Snapshot, error)
}

type Confidence string

const (
	ConfidencePass Confidence = "PASS"
	ConfidenceFail Confidence = "FAIL"
)

type FusionSnapshot struct {
	Symbol string
	TS     time.Time

	ConsensusRatePct float64
	Confidence       Confidence
	Reason           string

	Providers []ProviderResult
}

type ProviderResult struct {
	Provider string
	Snapshot Snapshot
	Error    string

	Stale   bool
	Outlier bool
	Inlier  bool
}

type Fusion interface {
	FetchFusion(ctx context.Context, symbol string) (FusionSnapshot, error)
}

