package optimizer

import (
	"math"
	"math/rand"
)

// BetaBernoulliArm models rewards in {0,1} with a Beta(a,b) posterior.
// This is the core used for Thompson Sampling.
type BetaBernoulliArm struct {
	Key string
	A   float64 // successes + 1
	B   float64 // failures + 1
	N   int
}

func NewArm(key string) BetaBernoulliArm {
	return BetaBernoulliArm{Key: key, A: 1, B: 1}
}

func (a *BetaBernoulliArm) Update(reward int) {
	if reward != 0 {
		a.A += 1
	} else {
		a.B += 1
	}
	a.N++
}

func (a BetaBernoulliArm) Mean() float64 {
	den := a.A + a.B
	if den <= 0 {
		return 0.5
	}
	return a.A / den
}

// Sample draws one sample from Beta(a,b) via Gamma sampling.
func (a BetaBernoulliArm) Sample(r *rand.Rand) float64 {
	x := gammaSample(r, a.A)
	y := gammaSample(r, a.B)
	if x+y == 0 {
		return 0.5
	}
	return x / (x + y)
}

// gammaSample draws from Gamma(k, 1) for k>0 using Marsaglia and Tsang (2000).
func gammaSample(r *rand.Rand, k float64) float64 {
	if k <= 0 {
		return 0
	}
	// Boost for k<1: Gamma(k) = Gamma(k+1) * U^(1/k)
	if k < 1 {
		u := r.Float64()
		return gammaSample(r, k+1) * math.Pow(u, 1/k)
	}

	d := k - 1.0/3.0
	c := 1.0 / math.Sqrt(9*d)
	for {
		x := r.NormFloat64()
		v := 1 + c*x
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := r.Float64()
		if u < 1-0.0331*(x*x)*(x*x) {
			return d * v
		}
		if math.Log(u) < 0.5*x*x+d*(1-v+math.Log(v)) {
			return d * v
		}
	}
}

