package optimizer

import (
	"fmt"
	"math/rand"
	"sort"
)

type ArmKey struct {
	Signal string
	Variant string // threshold/provider variant label
}

func (k ArmKey) String() string {
	if k.Variant == "" {
		return k.Signal
	}
	return k.Signal + "|" + k.Variant
}

type Bandit struct {
	Arms map[string]*BetaBernoulliArm
}

func NewBandit() *Bandit {
	return &Bandit{Arms: map[string]*BetaBernoulliArm{}}
}

func (b *Bandit) Ensure(key string) *BetaBernoulliArm {
	if key == "" {
		key = "unknown"
	}
	if a, ok := b.Arms[key]; ok {
		return a
	}
	arm := NewArm(key)
	b.Arms[key] = &arm
	return &arm
}

func (b *Bandit) Update(key string, reward int) {
	b.Ensure(key).Update(reward)
}

type Allocation struct {
	Key   string
	Score float64
	Mean  float64
	N     int
}

// SuggestAllocation uses Thompson Sampling to allocate a limited number of "action slots".
// It returns the top keys by sampled score.
func (b *Bandit) SuggestAllocation(r *rand.Rand, slots int) ([]Allocation, error) {
	if slots <= 0 {
		return nil, fmt.Errorf("slots must be >0")
	}
	var out []Allocation
	for k, a := range b.Arms {
		out = append(out, Allocation{
			Key:   k,
			Score: a.Sample(r),
			Mean:  a.Mean(),
			N:     a.N,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Mean > out[j].Mean
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > slots {
		out = out[:slots]
	}
	return out, nil
}

