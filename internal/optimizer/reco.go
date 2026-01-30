package optimizer

import (
	"math/rand"
	"sort"
	"time"

	"value-sniffer-radar/internal/reco"
)

// SuggestQuotas allocates "slots" by running Thompson sampling repeatedly.
// Each iteration samples all arms and assigns 1 slot to the best sampled arm.
// Deterministic given rng seed and bandit state.
func SuggestQuotas(rng *rand.Rand, b *Bandit, slots int) map[string]int {
	out := map[string]int{}
	if b == nil || len(b.Arms) == 0 || slots <= 0 {
		return out
	}
	for i := 0; i < slots; i++ {
		bestKey := ""
		bestScore := -1.0
		for k, a := range b.Arms {
			s := a.Sample(rng)
			if bestKey == "" || s > bestScore {
				bestKey = k
				bestScore = s
			}
		}
		if bestKey == "" {
			break
		}
		out[bestKey]++
	}
	return out
}

func BuildRecommendation(now time.Time, inputPaper string, inputLabels string, primaryWindowSec int, slots int, quotas map[string]int, b *Bandit) reco.Recommendation {
	var qs []reco.SignalQuota
	for k, a := range b.Arms {
		qs = append(qs, reco.SignalQuota{
			Signal:              k,
			MeanReward:          a.Mean(),
			N:                   a.N,
			SuggestedDailyQuota: quotas[k],
		})
	}
	sort.Slice(qs, func(i, j int) bool {
		if qs[i].SuggestedDailyQuota == qs[j].SuggestedDailyQuota {
			if qs[i].MeanReward == qs[j].MeanReward {
				return qs[i].Signal < qs[j].Signal
			}
			return qs[i].MeanReward > qs[j].MeanReward
		}
		return qs[i].SuggestedDailyQuota > qs[j].SuggestedDailyQuota
	})

	return reco.Recommendation{
		Version:          "reco.v1",
		GeneratedAt:      now,
		InputPaper:       inputPaper,
		InputLabels:      inputLabels,
		PrimaryWindowSec: primaryWindowSec,
		Slots:            slots,
		Quotas:           qs,
	}
}

func WriteReco(path string, r reco.Recommendation) error {
	return reco.Write(path, r)
}
