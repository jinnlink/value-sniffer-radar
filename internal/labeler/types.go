package labeler

import "time"

type Label struct {
	EventID   string    `json:"event_id"`
	EventTS   time.Time `json:"event_ts"`
	Source    string    `json:"source"`
	Symbol    string    `json:"symbol"`
	TradeDate string    `json:"trade_date"`

	WindowSec int `json:"window_sec"`
	GraceSec  int `json:"grace_sec"`
	LateBySec int `json:"late_by_sec"`

	Threshold    float64 `json:"threshold"`
	EntryRatePct float64 `json:"entry_rate_pct"`
	ExitRatePct  float64 `json:"exit_rate_pct"`

	Confidence string `json:"confidence"`
	Reward     int    `json:"reward"`
	Reason     string `json:"reason"`
}

