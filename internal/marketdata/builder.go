package marketdata

import (
	"fmt"
	"strings"
	"time"

	"value-sniffer-radar/internal/config"
)

func Build(cfg config.MarketdataConfig) (Fusion, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	var providers []Provider
	for _, pc := range cfg.Providers {
		switch strings.TrimSpace(pc.Type) {
		case "eastmoney_repo":
			providers = append(providers, NewEastmoneyRepo(EastmoneyRepoOptions{
				Name:        pc.Name,
				BaseURL:     pc.BaseURL,
				Fields:      pc.Fields,
				RateDivisor: pc.RateDivisor,
				Timeout:     time.Duration(cfg.TimeoutMS) * time.Millisecond,
			}))
		case "tencent_repo":
			providers = append(providers, NewTencentRepo(TencentRepoOptions{
				Name:     pc.Name,
				QuoteURL: pc.QuoteURL,
				Timeout:  time.Duration(cfg.TimeoutMS) * time.Millisecond,
			}))
		default:
			return nil, fmt.Errorf("marketdata unknown provider type: %s", pc.Type)
		}
	}
	f, err := NewFusion(providers, FusionConfig{
		Timeout:          time.Duration(cfg.TimeoutMS) * time.Millisecond,
		RequiredSources:  cfg.RequiredSources,
		MaxAbsDiff:       cfg.MaxAbsDiff,
		Staleness:        time.Duration(cfg.StalenessSec) * time.Second,
		MinValid:         cfg.MinValid,
		MaxValid:         cfg.MaxValid,
		FailThreshold:    cfg.FailThreshold,
		OutlierThreshold: cfg.OutlierThreshold,
		Cooldown:         time.Duration(cfg.CooldownSec) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

