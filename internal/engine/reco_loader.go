package engine

import (
	"log"

	"value-sniffer-radar/internal/reco"
)

func (e *Engine) loadRecoIfConfigured() {
	path := ""
	if e.cfg != nil {
		path = e.cfg.Engine.RecoPath
	}
	if path == "" {
		return
	}
	r, err := reco.Read(path)
	if err != nil {
		log.Printf("reco load failed path=%s err=%v", path, err)
		return
	}
	m := map[string]int{}
	for _, q := range r.Quotas {
		if q.Signal == "" || q.SuggestedDailyQuota <= 0 {
			continue
		}
		m[q.Signal] = q.SuggestedDailyQuota
	}
	if len(m) == 0 {
		log.Printf("reco loaded path=%s but no quotas found", path)
		return
	}
	e.recoQuotas = m
	log.Printf("reco loaded path=%s quotas=%d window_sec=%d", path, len(m), r.PrimaryWindowSec)
}
