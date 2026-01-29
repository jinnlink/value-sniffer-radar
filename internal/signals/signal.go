package signals

import (
	"context"
	"fmt"
	"time"

	"value-sniffer-radar/internal/config"
	"value-sniffer-radar/internal/marketdata"
	"value-sniffer-radar/internal/notifier"
	"value-sniffer-radar/internal/tushare"
)

type Signal interface {
	Name() string
	MinInterval() time.Duration
	Evaluate(ctx context.Context, client *tushare.Client, tradeDate string, md marketdata.Fusion) ([]notifier.Event, error)
}

func BuildAll(cfgs []config.SignalConfig) ([]Signal, error) {
	var out []Signal
	for _, c := range cfgs {
		if !c.Enabled {
			continue
		}
		switch c.Type {
		case "cb_premium":
			out = append(out, NewCBPremium(c))
		case "cb_double_low":
			out = append(out, NewCBDoubleLow(c))
		case "fund_premium":
			out = append(out, NewFundPremium(c))
		case "cn_repo_sniper":
			out = append(out, NewCNRepoSniper(c))
		case "cn_repo_realtime":
			out = append(out, NewCNRepoRealtime(c))
		default:
			return nil, fmt.Errorf("unknown signal type: %s", c.Type)
		}
	}
	return out, nil
}
